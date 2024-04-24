package scraper

import (
	"context"
	"fmt"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/pkg/constant"
	"ibanking-scraper/utils"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const linkBRICMS = "https://ibank.bri.co.id/cms/Logon.aspx"

func (i *Ibanking) StartBRICMS() {
Login:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.LoginBRICMS(page, i.BankAccount.CompanyId.String, i.BankAccount.UserId.String, i.BankAccount.Password.String)
			if err != nil {
				time.Sleep(time.Second)
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("Terjadi error, mencoba login ulang")
					page.Close()
					goto Login
				} else {
					return
				}
			} else {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
			Scrape:
				if len(browser.Contexts()) != 0 {
					page, err := i.resultBRICMS(page)
					if err != nil {
						time.Sleep(time.Second)
						if len(browser.Contexts()) != 0 {
							i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
							i.Log("terjadi error, mencoba scraping ulang")
							if err = i.backToHomeBRICMS(page); err != nil {
								if len(browser.Contexts()) != 0 {
									i.UpdateLoginStatus(domain.BelumLogin)
									i.Log("tidak bisa kembali ke menu utama, logout...")
									if err := i.LogoutBRICMS(page); err != nil {
										i.Log("proses logout gagal, menutup halaman")
										page.Close()
									}
									goto Login
								} else {
									return
								}
							}
							goto Scrape
						} else {
							return
						}
					} else {
						if i.BankAccount.AutoLogout.Bool {
							i.UpdateLoginStatus(domain.BelumLogin)
							i.LogoutBRICMS(page)
							i.Log("Browser masih terbuka")
						}
						i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					}
				} else {
					return
				}
			}
		} else {
			return
		}
	}
}

func (i *Ibanking) LoginBRICMS(page playwright.Page, corpId, userId, pass string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("Menuju website BRI CMS")
	logger.Debug("menuju website bri cms")

	if _, err := page.Goto(linkBRICMS, playwright.PageGotoOptions{
		Timeout: playwright.Float(100000),
	}); err != nil {
		i.Log("Gagal menuju website BRI CMS")
		logger.Debug("Gagal menuju website BRI CMS: ", err)
		return page, errors.New("Gagal menuju website BRI CMS")
	}

	i.Log("mulai proses login")
	logger.Debug("mulai proses login")

	closePopup, err := page.WaitForSelector("#TB_closeWindowButton", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(1000)})
	if err == nil {
		closePopup.Click()
	}
	page.Keyboard().Press(`Escape`)

	clientID, err := page.WaitForSelector("input#ClientID", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = clientID.Type(corpId); err == nil {
		i.Log("mengetik company id.")
		logger.Debug("mengetik company id.")
	}

	userID, err := page.WaitForSelector("input#UserID", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = userID.Type(userId); err == nil {
		logger.Debug("mengetik user id")
		i.Log("mengetik user id")
	}

	password, err := page.WaitForSelector("input#Password", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = password.Type(pass); err == nil {
		logger.Debug("mengetik password")
		i.Log("mengetik password")
	}

	btnLogin, err := page.WaitForSelector("#btnLogin", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	} else {
		page.WaitForTimeout(1000)
		if err = btnLogin.Click(); err == nil {
			msgError, err := page.WaitForSelector("#Msg", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(2000),
			})
			if err == nil {
				page.WaitForTimeout(1000)
				msg, err := msgError.TextContent()
				if err == nil {
					if msg == "You are already login from other device. Please wait for 15 minutes after last login" || strings.Contains(msg, "already login") {
						i.Log(msg)
						logger.Debug(msg)
						i.Log(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						logger.Debug(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
						return page, errors.New(msg)
					} else {
						i.UpdateLoginStatus(domain.TidakAktif)
						logger.Debug(msg)
						i.ErrorLog(msg)
						i.StopScrape()
						return nil, nil
					}
				}
			}
			i.Log("login berhasil")
			logger.Debug("login berhasil")
		}
	}

	return page, nil
}

func (i *Ibanking) ScrapeBRICMS(page playwright.Page, nomorRekening string, totalCheckDate int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.Log("mulai proses scraping")
	i.UpdateLoginStatus(domain.ProsesScraping)
	logger.Debug("mulai proses scraping")

	frameHead, err := page.WaitForSelector("frame[name='head']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(100000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	page.WaitForTimeout(1000)
	frameHeadContent, err := frameHead.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	menuAccountInformation, err := frameHeadContent.WaitForSelector("a[target='menu']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuAccountInformation.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	frameMenu, err := page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameMenuContent, err := frameMenu.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	accountInformation, err := frameMenuContent.WaitForSelector("td[onclick=\"javascript:hideAllMenu();visible('m_240')\"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = accountInformation.Click(); err != nil {
		logger.Debug("Gagal klik Account Information :", err)
		return nil, page, err
	}

	accountSummary, err := frameMenuContent.WaitForSelector("a[onclick=\"javascript:setBiggerFont('sm_247', 'm_240');\"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = accountSummary.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	frameChannel, err := page.WaitForSelector("frame[name='channel']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameChannelContent, err := frameChannel.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	btnSubmitSummary, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_btnOk", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(40000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = btnSubmitSummary.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(8000)

	frameChannel, err = page.WaitForSelector("frame[name='channel']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameChannelContent, err = frameChannel.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	frameReport, err := frameChannelContent.WaitForSelector("#ReportFramectl00_TransactionForm_ReportViewer1", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	frameReportContent, err := frameReport.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	frameReportViewer, err := frameReportContent.WaitForSelector("#report", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	frameReportViewerContent, err := frameReportViewer.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	var pemilikrekening string
	rekenings := []*domain.Rekening{}
	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	if _, err = frameReportViewerContent.WaitForSelector("table.a224"); err == nil {
		logger.Debug("get table saldo rekening.")
		tableRekenings, err := frameReportViewerContent.QuerySelectorAll("table.a224")
		if err != nil {
			logger.Debug(err)
			return nil, page, err
		}

		for _, tableRekening := range tableRekenings {
			tableRowRekening, err := tableRekening.QuerySelectorAll("tr")
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}

		row:
			for _, rekRow := range tableRowRekening {
				cells, err := rekRow.QuerySelectorAll("td div")
				if err != nil {
					logger.Debug(err)
					return nil, page, err
				}
				rekening := &domain.Rekening{}
				for j, cell := range cells {
					cellValue, err := cell.InnerText()
					if err != nil {
						logger.Debug(err)
						return nil, page, err
					}
					switch j {
					case 1:
						noRek := strings.TrimSuffix(cellValue, "\n")
						noRek = strings.ReplaceAll(noRek, "\u00a0", "")
						noRek = strings.ReplaceAll(noRek, "&nbsp;", "")
						if noRek == "0.00" || cellValue == "No Rekening" {
							continue row
						} else {
							rekening = &domain.Rekening{}
							rekening.TipeBank = domain.BankTypeBRI
							rekening.Rekening = noRek
						}
					case 2:
						if cellValue == "Nama" || strings.Contains(cellValue, ",") {
							continue row
						}
						pemilikRekening := strings.TrimSuffix(cellValue, "\n")
						rekening.PemilikRekening = pemilikRekening
					case 5:
						if cellValue == "Ledger" || cellValue == "" {
							continue row
						}
						rekening.SaldoStr = cellValue
						saldo := strings.Replace(cellValue, ",", "", -1)
						saldo = strings.Replace(saldo, "\n", "", -1)
						saldoSep := strings.Split(saldo, ".")
						saldoSep[0] = strings.ReplaceAll(saldoSep[0], "\u00a0", "")
						saldoSep[0] = strings.ReplaceAll(saldoSep[0], "&nbsp;", "")
						saldoInt, err := strconv.ParseInt(saldoSep[0], 0, 64)
						if err != nil {
							//logger.Debug(err)
						}
						rekening.Saldo = saldoInt
					}
				}
				if rekening.Rekening != "" && rekening.PemilikRekening != "" {
					rekenings = append(rekenings, rekening)
					if rekening.Rekening == nomorRekening {
						if rekening.Saldo > 0 {
							pemilikrekening = rekening.PemilikRekening
							rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
						}
					}
				}
			}
		}
	}

	i.Log("Daftar Rekening:")
	logger.Debug("Daftar Rekening:")
	nomor := 1
	for _, rek := range rekenings {
		i.Log(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", nomor, rek.Rekening, rek.PemilikRekening))
		logger.Debug(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", nomor, rek.Rekening, rek.PemilikRekening))
		nomor++
	}
	rekenings = nil

	logger.Debug("Proses scraping saldo rekening.")
	i.Log("proses scraping saldo rekening.")

	for _, rekening := range rekeningMutasiWraper.Rekening {
		i.Log(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", nomorRekening, rekening.PemilikRekening, rekening.SaldoStr))
		logger.Debug(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", nomorRekening, rekening.PemilikRekening, rekening.SaldoStr))
		//logger.Debug(fmt.Sprintf("Rekening: %s | Pemilik Rekening: %s\nUpdate saldo menjadi: Rp%s", nomorRekening, rekening.PemilikRekening, rekening.SaldoStr))
	}

	frameMenu, err = page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameMenuContent, err = frameMenu.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	page.WaitForTimeout(2000)
	accountStatement, err := frameMenuContent.WaitForSelector("a[onclick=\"javascript:setBiggerFont('sm_249', 'm_240');\"]", playwright.PageWaitForSelectorOptions{ //#sm_249 > a
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = accountStatement.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	frameChannel, err = page.WaitForSelector("html > frameset > frameset > frame:nth-child(2)", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameChannelContent, err = frameChannel.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(5000)
	fieldNomorRekening, err := frameChannelContent.WaitForSelector("input#ctl00_TransactionForm_txtNoRek", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(100000),
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = fieldNomorRekening.Type(nomorRekening); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	now := time.Now()
	totalDateDuration := time.Duration(totalCheckDate)
	totalDate := now.Add(-totalDateDuration * 24 * time.Hour)
	startCheckDate := totalDate.Format("02/01/2006")
	i.Log(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, pemilikrekening, startCheckDate))
	logger.Debug(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, pemilikrekening, startCheckDate))

	fillDate, err := frameChannelContent.WaitForSelector("input[name='ctl00$TransactionForm$txtstartdate']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = fillDate.Fill(startCheckDate); err != nil {
		logger.Debug("gagal klik start date :", err)
		return nil, page, err
	}

	cbLedger, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_rdioLedger", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = cbLedger.Check(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	btnSubmit, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_btnSubmit", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(500)
	if err = btnSubmit.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	information, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_Msg", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	})
	if err != nil {
		var totalPage int
		page.WaitForTimeout(10000)
		pages, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_ReportViewer1_ctl01_ctl01_ctl04", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(60000),
		})
		if err != nil {
			logger.Debug(err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		pagesString, err := pages.InnerText()
		if err != nil {
			logger.Debug(err)
		} else {
			if pagesString != "0" {
				totalPages, err := strconv.Atoi(pagesString)
				if err != nil {
					logger.Debug(err)
					return nil, page, err
				}
				logger.Debug(totalPages)
				totalPage = totalPages
			} else {
				logger.Debug("total page not loaded")
				return nil, page, errors.New("page not loaded")
			}
		}

		if totalPage > 0 {
			logger.Debug("mutasi ditemukan")
			i.Log("mutasi ditemukan")

			frameReport, err = frameChannelContent.WaitForSelector("#ReportFramectl00_TransactionForm_ReportViewer1", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			frameReportContent, err = frameReport.ContentFrame()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}

			frameReportViewer, err = frameReportContent.WaitForSelector("frame#report", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			frameReportViewerContent, err = frameReportViewer.ContentFrame()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}

			accountName, err := frameReportViewerContent.WaitForSelector("td[class=\"a55l\"] div", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			accountNameString, err := accountName.InnerText()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}

			frameReportViewer, err = frameReportContent.WaitForSelector("frame#report", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			frameReportViewerContent, err = frameReportViewer.ContentFrame()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}

			for k := 1; k <= totalPage; k++ {
				logger.Debug(fmt.Sprintf("Proses scraping mutasi rekening, halaman: %d", k))
				i.Log(fmt.Sprintf("Proses scraping mutasi, halaman: %d", k))

				frameReport, err = frameChannelContent.WaitForSelector("#ReportFramectl00_TransactionForm_ReportViewer1", playwright.PageWaitForSelectorOptions{
					State:   playwright.WaitForSelectorStateAttached,
					Timeout: playwright.Float(60000),
				})
				if err != nil {
					return nil, page, err
				}
				frameReportContent, err = frameReport.ContentFrame()
				if err != nil {
					return nil, page, err
				}
				frameReportViewer, err = frameReportContent.WaitForSelector("frame#report", playwright.PageWaitForSelectorOptions{
					State:   playwright.WaitForSelectorStateAttached,
					Timeout: playwright.Float(60000),
				})
				if err != nil {
					return nil, page, err
				}
				frameReportViewerContent, err = frameReportViewer.ContentFrame()
				if err != nil {
					return nil, page, err
				}
				tblMutasiRows, err := frameReportViewerContent.QuerySelectorAll("table.a159 tr[valign]")
				if err != nil {
					logger.Debug(err)
					return nil, page, err
				}

			rowMutasi:
				for _, mutasiRow := range tblMutasiRows {
					mutasi := &domain.Mutasi{}
					cells, err := mutasiRow.QuerySelectorAll("td")
					if err != nil {
						logger.Debug(err)
						return nil, page, err
					}
					for k, cell := range cells {
						cellValue, err := cell.InnerText()
						if err != nil {
							logger.Debug(err)
							return nil, page, err
						}
						switch k {
						case 0:
							date := strings.Trim(cellValue, "\n")
							if date == "DATE" {
								continue rowMutasi
							} else {
								date, _ := time.Parse(constant.LayoutDate, date)
								pgDate := types.PGDate{
									Time: date,
								}
								mutasi = &domain.Mutasi{}
								mutasi.TglBank = pgDate
								mutasi.TipeBank = domain.BankTypeBRI
								mutasi.Rekening = nomorRekening
								mutasi.PemilikRekening = accountNameString
							}
						case 2:
							ket := strings.Trim(cellValue, "\n")
							if ket != "REMARK" {
								mutasi.Keterangan = ket
							} else {
								continue rowMutasi
							}
						case 3:
							if cellValue != "0.00" {
								jumlah := strings.Replace(cellValue, ",", "", -1)
								jumlahSep := strings.Split(jumlah, ".")
								jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "\u00a0", "")
								jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "&nbsp;", "")
								jumlahInt, err := strconv.ParseInt(jumlahSep[0], 0, 64)
								if err != nil {
									logger.Debug("parsing dari string to float error :", err)
								}
								mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
								mutasi.Jumlah = jumlahInt
							} else {
								continue
							}
						case 4:
							if cellValue != "0.00" {
								jumlah := strings.Replace(cellValue, ",", "", -1)
								jumlahSep := strings.Split(jumlah, ".")
								jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "\u00a0", "")
								jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "&nbsp;", "")
								jumlahInt, err := strconv.ParseInt(jumlahSep[0], 0, 64)
								if err != nil {
									logger.Debug("parsing dari string to float error :", err)
								}
								mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
								mutasi.Jumlah = jumlahInt
							} else {
								continue
							}
						case 5:
							saldo := strings.Replace(cellValue, ",", "", -1)
							saldoSep := strings.Split(saldo, ".")
							saldoSep[0] = strings.ReplaceAll(saldoSep[0], "\u00a0", "")
							saldoSep[0] = strings.ReplaceAll(saldoSep[0], "&nbsp;", "")
							saldoInt, err := strconv.ParseInt(saldoSep[0], 0, 64)
							if err != nil {
								logger.Debug("parsing dari string to float error :", err)
							}
							mutasi.Saldo = saldoInt
						}
					}
					if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
						rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
					}
				}

				btnNext, err := frameChannelContent.WaitForSelector("#ctl00_TransactionForm_ReportViewer1_ctl01_ctl01_ctl05_ctl00", playwright.PageWaitForSelectorOptions{
					State:   playwright.WaitForSelectorStateVisible,
					Timeout: playwright.Float(2000),
				})
				if err == nil {
					btnNextHidden, err := btnNext.IsHidden()
					if err == nil {
						if btnNextHidden {
							break
						} else {
							if err = btnNext.Click(); err == nil {
								page.WaitForTimeout(2500)
							}
						}
					}
				}
			}
		} else {
			logger.Debug("mutasi tidak ditemukan")
		}
	} else {
		page.WaitForTimeout(1000)
		msg, err := information.InnerText()
		if err != nil {
			logger.Debug(err)
		}
		logger.Debug(msg)
		if strings.Contains(msg, "Not Authorize") {
			i.Log("Rekening Tidak Ditemukan")
			logger.Debug(msg)
		}
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	logger.Debug("Proses scraping mutasi selesai.")
	i.Log("Proses scraping selesai.")

	backToMenu, err := frameHeadContent.WaitForSelector("a[onclick=\"loadtwo('LeftMenu.aspx','Welcome.aspx') \"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("could not get backToMenu :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = backToMenu.Click(); err != nil {
		logger.Debug("could not click backToMenu :", err)
		return nil, page, err
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) LogoutBRICMS(page playwright.Page) error {
	menuFrame, err := page.WaitForSelector("frame[name='head']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	menuFrameContent, err := menuFrame.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return err
	}

	i.Log("proses logout")
	logger.Debug("proses logout")

	btnLogout, err := menuFrameContent.WaitForSelector("a[id='btnLogout']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	if err = btnLogout.Click(); err != nil {
		logger.Debug(err)
		return err
	}

	closeBtn, err := page.WaitForSelector("#TB_closeWindowButton", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
	} else {
		closeBtn.Click()
		logger.Debug("close popup")
	}

	if err = page.Close(); err != nil {
		logger.Debug(err)
	}

	i.Log("logout berhasil")
	logger.Debug("logout berhasil")

	return nil
}

func (i *Ibanking) IsBRICMSLogin(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input#ClientID", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) resultBRICMS(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.ScrapeBRICMS(
				page,
				rekOp,
				i.BankAccount.TotalCekHari.Int64,
			)
			if err != nil {
				return page1, err
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1

			if len(result.Rekening) > 0 {
				if err := i.UcRekening.BulkInsertRekOnpay(context.Background(), result.Rekening); err != nil {
					logger.Debug(err)
				} else {
					result.Rekening = nil
				}
			}

			mutasis, err := i.UcMutasi.Fetch(context.Background(), rekOp)
			if err != nil {
				logger.Debug(err)
			}
			diff := utils.DifferenceMutasi(i.RemoveDuplicateMutasi(result.Mutasi), mutasis)
			if len(diff) > 0 {
				if err := i.UcMutasi.BulkInsertRekOnpay(context.Background(), diff); err != nil {
					logger.Debug(err)
				} else {
					result.Mutasi = nil
				}
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
		}
	}

	if len(i.BankAccount.RekGriyabayar) > 0 {
		for _, rekGb := range i.BankAccount.RekGriyabayar {
			result, page1, err := i.ScrapeBRICMS(
				page,
				rekGb,
				i.BankAccount.TotalCekHari.Int64,
			)
			if err != nil {
				return page1, err
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1

			if len(result.Rekening) > 0 {
				if err := i.UcRekening.BulkInsertRekGriyabayar(context.Background(), result.Rekening); err != nil {
					logger.Debug(err)
				} else {
					result.Rekening = nil
				}
			}

			mutasis, err := i.UcMutasi.Fetch(context.Background(), rekGb)
			if err != nil {
				logger.Debug(err)
			}
			diff := utils.DifferenceMutasi(i.RemoveDuplicateMutasi(result.Mutasi), mutasis)
			if len(diff) > 0 {
				if err := i.UcMutasi.BulkInsertRekGriyabayar(context.Background(), diff); err != nil {
					logger.Debug(err)
				} else {
					result.Mutasi = nil
				}
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
		}
	}

	return page, nil
}

func (i *Ibanking) backToHomeBRICMS(page playwright.Page) error {
	frameHead, err := page.WaitForSelector("frame[name=\"head\"]", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		return err
	}
	page.WaitForTimeout(500)
	frameHeadContent, err := frameHead.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return err
	}
	homeBtn, err := frameHeadContent.WaitForSelector("div.menu-item:nth-child(1) > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	if err := homeBtn.Click(playwright.ElementHandleClickOptions{ClickCount: playwright.Int(4)}); err != nil {
		logger.Debug(err)
		return err
	}
	i.Log("terjadi error, kembali ke menu utama")
	logger.Debug("terjadi error, kembali ke menu utama")

	return nil
}
