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

const linkBNIDirect = "https://bnidirect.bni.co.id/corp/common/login.do?action=loginRequest"

func (i *Ibanking) StartBNIDirect() {
Login:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.LoginBNIDirect(page, i.BankAccount.CompanyId.String, i.BankAccount.UserId.String, i.BankAccount.Password.String)
			if err != nil {
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("terjadi error ketika login")
					page.Close()
					goto Login
				} else {
					return
				}
			} else {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
			Scrape:
				if len(browser.Contexts()) != 0 {
					page, err := i.resultBNIDirect(page)
					if err != nil {
						time.Sleep(time.Second)
						if len(browser.Contexts()) != 0 {
							i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
							i.Log("terjadi error, mencoba scraping ulang")
							if err := i.backToHomeBNIDirect(page); err != nil {
								if len(browser.Contexts()) != 0 {
									i.UpdateLoginStatus(domain.BelumLogin)
									i.Log("tidak bisa kembali ke menu utama, logout...")
									if err := i.LogoutBNIDirect(page); err != nil {
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
							i.LogoutBNIDirect(page)
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

func (i *Ibanking) LoginBNIDirect(page playwright.Page, corpId, username, password string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("menuju website BNI Direct")
	logger.Debug("Menuju website BNI Direct")

	if _, err := page.Goto(linkBNIDirect, playwright.PageGotoOptions{
		Timeout: playwright.Float(100000),
	}); err != nil {
		i.Log("Gagal menuju website BNI Direct")
		logger.Debug("Gagal menuju website BNI Direct: ", err)
		return page, err
	}

	i.Log("proses login")
	logger.Debug("Proses login")

	pageNotFound, err := page.WaitForSelector("body > font > table:nth-child(2) > tbody > tr > td > font > h2", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(500),
	})
	if err == nil {
		msg, err := pageNotFound.InnerText()
		if err == nil {
			logger.Debug(msg)
		}
	}

	i.closePopupBNIDirect(page)

	corpIdField, err := page.WaitForSelector("input#corpId_placeholder", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom corpId :", err)
		return page, err
	}
	page.WaitForTimeout(500)
	if err = corpIdField.Type(corpId); err != nil {
		logger.Debug("Gagal mengetik corpId :", err)
	}
	logger.Debug("mengetik company id")
	i.Log("mengetik company id")

	usernameField, err := page.WaitForSelector("input#userName_placeholder", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom username :", err)
		return page, err
	}
	page.WaitForTimeout(500)
	if err = usernameField.Type(username); err != nil {
		logger.Debug("could not type username :", err)
	}
	logger.Debug("mengetik username")
	i.Log("mengetik username")

	passwordField, err := page.WaitForSelector("input#password_placeholder", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom password :", err)
		return page, err
	}
	page.WaitForTimeout(500)
	if err = passwordField.Type(password); err != nil {
		logger.Debug("Gagal mengetik password  :", err)
	}
	logger.Debug("mengetik password")
	i.Log("mengetik password")

	btnLogin, err := page.WaitForSelector("input[name='submit1']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	} else {
		if err = btnLogin.Click(); err == nil {
			msgError, err := page.WaitForSelector("#bni_login_form > div > div:nth-child(1) > div.col-12.col-xl-11.m-0.BNI-Content-OrWh.text-center > b > font", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(2000)})
			if err == nil {
				msg, err := msgError.TextContent()
				if err == nil {
					if strings.Contains(msg, "Pengguna atau Kata Sandi Salah") {
						i.UpdateLoginStatus(domain.TidakAktif)
						i.ErrorLog("Pengguna atau Kata Sandi Salah")
						i.StopScrape()
						return nil, errors.New("Pengguna atau Kata Sandi Salah")
					} else if strings.Contains(msg, "Pengguna sedang login") {
						//i.Log(msg)
						//logger.Debug(msg)
						//i.Log(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						//logger.Debug(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						//time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
						btnYes, err := page.WaitForSelector("button.btn:nth-child(1)", playwright.PageWaitForSelectorOptions{
							Timeout: playwright.Float(5000),
							State:   playwright.WaitForSelectorStateVisible,
						})
						if err == nil {
							btnYes.Click()
						}
					}
				}
			}
			i.Log("login berhasil")
		}
	}

	return page, nil
}

func (i *Ibanking) scrapeBNIDirect(page playwright.Page, nomorRekening string, totalCheckDate int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.Log("mulai proses scraping")
	i.UpdateLoginStatus(domain.ProsesScraping)
	logger.Debug("memulai proses scraping")

	var norek string
	norekSplit := strings.Split(nomorRekening, "")
	if norekSplit[0] == "0" {
		norek = strings.Join(norekSplit[1:], "")
	} else {
		norek = nomorRekening
	}

	iframeSidebar, err := page.WaitForSelector("iframe#sidebar", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe sidebar menu utama : ", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	iframeSidebarContent, err := iframeSidebar.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe sidebar content menu utama: ", err)
		return nil, page, err
	}

	accountInformationMenu, err := iframeSidebarContent.WaitForSelector("#parent-MNU_GCME_040000", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan account information menu: ", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = accountInformationMenu.Click(); err != nil {
		logger.Debug("Gagal klik account information menu :", err)
		return nil, page, err
	}

	balanceMenu, err := iframeSidebarContent.WaitForSelector("#child-MNU_GCME_040100", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan account information menu: ", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = balanceMenu.Click(); err != nil {
		logger.Debug("Gagal klik balance menu :", err)
		return nil, page, err
	}

	iframeMain, err := page.WaitForSelector("iframe#content")
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe main: ", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	iframeMainContent, err := iframeMain.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe main content: ", err)
		return nil, page, err
	}

	reks := []*DaftarRekening{}
	page.WaitForTimeout(2000)
	btnShowRek, err := iframeMainContent.QuerySelectorAll(".col-5 > img")
	if err != nil {
		logger.Debug(err)
	} else {
		if len(btnShowRek) > 1 {
			page.WaitForTimeout(2000)
			if err = btnShowRek[0].Click(); err != nil {
				logger.Debug(err)
			} else {
				page.WaitForTimeout(4000)
				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					browser := value.(playwright.Browser)
					if len(browser.Contexts()) != 0 {
						browserCtx := value.(playwright.Browser).Contexts()[0]
						if len(browserCtx.Pages()) > 1 {
							tableRek, err := browserCtx.Pages()[1].QuerySelectorAll(".BNI-Table-Aprisma tr")
							if err != nil {
								logger.Debug(err)
								return nil, page, err
							}
						row:
							for _, rek := range tableRek {
								datas, err := rek.QuerySelectorAll("td")
								if err != nil {
									logger.Debug(err)
								}
								rekening := &DaftarRekening{}
								for a, value := range datas {
									cellValue, err := value.InnerText()
									if err != nil {
										logger.Debug(err)
									}
									switch a {
									case 0:
										if strings.Contains(cellValue, "No") {
											continue row
										}
									case 1:
										rekening.Rekening = cellValue
									case 2:
										rekening.PemilikRekening = cellValue
									}
								}
								reks = append(reks, rekening)
							}
							page.WaitForTimeout(1000)
							if len(browser.Contexts()) != 0 {
								if err = browserCtx.Pages()[1].Close(); err != nil {
									logger.Debug(err)
								}
							}
						}
					}
				}
			}
		}
	}

	i.Log("Daftar Rekening:")
	logger.Debug("Daftar Rekening:")
	o := 1
	for _, rek := range reks {
		if len(rek.Rekening) != 10 {
			i.Log(fmt.Sprintf("No: %d | Rekening: 0%s | Pemilik Rekening: %s", o, rek.Rekening, rek.PemilikRekening))
			logger.Debug(fmt.Sprintf("No: %d | Rekening: 0%s | Pemilik Rekening: %s", o, rek.Rekening, rek.PemilikRekening))
		} else {
			i.Log(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", o, rek.Rekening, rek.PemilikRekening))
			logger.Debug(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", o, rek.Rekening, rek.PemilikRekening))
		}
		o++
	}

	fieldNomorRekening, err := iframeMainContent.WaitForSelector("input.picklist-form-control", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom NomorRekening :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = fieldNomorRekening.Type(norek); err != nil {
		logger.Debug("Gagal mengetik di kolom Nomor Rekening :", err)
		return nil, page, err
	}

	btnShow, err := iframeMainContent.WaitForSelector("body > form > div > div > div:nth-child(2) > div > input[type=\"button\"]:nth-child(1)", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan btn show: ", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = btnShow.Click(); err != nil {
		logger.Debug("Gagal klik btn show :", err)
		return nil, page, err
	}

	rekening := &domain.Rekening{}
	rekening.TipeBank = domain.BankTypeBNI
	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	nomorRekeningNotFound, err := iframeMainContent.WaitForSelector("font[class='BNI-Message']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(4000)})
	if err != nil {
		iframeMain, err = page.WaitForSelector("iframe#content", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateAttached,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe main: ", err)
			return nil, page, err
		}
		page.WaitForTimeout(1000)
		iframeMainContent, err = iframeMain.ContentFrame()
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe main content: ", err)
			return nil, page, err
		}

		if _, err = iframeMainContent.WaitForSelector("table.BNI-table td[width='70%']", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(2000),
		}); err == nil {
			page.WaitForTimeout(2000)
			balanceDatas, err := iframeMainContent.QuerySelectorAll("table.BNI-table td[width='70%']")
			if err != nil {
				logger.Debug("Gagal mendapatkan data rekening saldo : ", err)
				return nil, page, err
			}

			logger.Debug("proses scraping rekening saldo.")
			i.Log("proses scraping rekening saldo.")

			var saldoStr string
			for j, balanceData := range balanceDatas {
				balanceValue, err := balanceData.InnerText()
				if err != nil {
					logger.Debug("Gagal mendapatkan rekening saldo value string : ", err)
					return nil, page, err
				}
				switch j {
				case 0:
					rekening = &domain.Rekening{}
					rekening.TipeBank = domain.BankTypeBNI
					rekening.Rekening = nomorRekening
				case 4:
					pemilikRekening := strings.Split(balanceValue, "/")[1]
					pemilikRekening = strings.TrimSpace(pemilikRekening)
					pemilikRekening = strings.Replace(pemilikRekening, "(", "", -1)
					pemilikRekening = strings.Replace(pemilikRekening, ")", "", -1)
					pemilikRekening = strings.Replace(pemilikRekening, "IDR", "", -1)
					rekening.PemilikRekening = pemilikRekening
				case 5:
					saldo, _ := balanceData.InnerText()
					saldo = strings.Replace(saldo, ":", "", -1)
					saldoStr = saldo
					saldo = strings.Replace(saldo, ",", "", -1)
					saldo = strings.TrimSpace(saldo)
					saldoSep := strings.Split(saldo, ".")
					saldoSep[0] = strings.ReplaceAll(saldoSep[0], "\u00a0", "")
					saldoSep[0] = strings.ReplaceAll(saldoSep[0], "&nbsp;", "")
					saldoInt, err := strconv.Atoi(saldoSep[0])
					if err != nil {
						logger.Debug("error parse string to float:", err)
					}
					rekening.Saldo = int64(saldoInt)
				}
			}
			if rekening.Saldo > 0 {
				i.Log(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", nomorRekening, strings.TrimSuffix(rekening.PemilikRekening, " "), strings.TrimPrefix(saldoStr, " ")))
				logger.Debug(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", nomorRekening, strings.TrimSuffix(rekening.PemilikRekening, " "), strings.TrimPrefix(saldoStr, " ")))
				rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
			}
		}

		// mutasi menu
		page.WaitForTimeout(1000)
		iframeSidebar, err = page.WaitForSelector("iframe#sidebar", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe sidebar menu utama : ", err)
			return nil, page, err
		}
		page.WaitForTimeout(1000)
		iframeSidebarContent, err = iframeSidebar.ContentFrame()
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe sidebar content menu utama : ", err)
			return nil, page, err
		}

		mutasiMenu, err := iframeSidebarContent.WaitForSelector("a#child-MNU_GCME_040200", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element menu mutasi : ", err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		if err = mutasiMenu.Click(); err != nil {
			logger.Debug("Gagal klik menu mutasi :", err)
			return nil, page, err
		}

		iframeMain, err = page.WaitForSelector("iframe#content", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateAttached,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe main: ", err)
			return nil, page, err
		}
		iframeMainContent, err = iframeMain.ContentFrame()
		page.WaitForTimeout(1000)
		if err != nil {
			logger.Debug("Gagal mendapatkan iframe main content: ", err)
			return nil, page, err
		}

		startDate := time.Now().AddDate(0, 0, -int(totalCheckDate)).Format("02/01/2006")
		i.Log(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, rekening.PemilikRekening, startDate))
		logger.Debug(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, strings.TrimSuffix(rekening.PemilikRekening, " "), startDate))

		btnClearDate, err := iframeMainContent.WaitForSelector("div.row:nth-child(4) > div:nth-child(3) > a:nth-child(7) > img")
		if err != nil {
			logger.Debug("Gagal mendapatkan btnClearDate :", err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		if err = btnClearDate.Click(); err != nil {
			logger.Debug("Gagal klik btnClearDate :", err)
			return nil, page, err
		}

		inputDate, err := iframeMainContent.WaitForSelector("input[name='transferDateDisplay1']", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan inputDate :", err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		if err = inputDate.Fill(startDate); err != nil {
			logger.Debug("fill inputData err :", err)
			return nil, page, err
		}

		fieldNomorRekeningMutasi, err := iframeMainContent.WaitForSelector("input[name='accountDisplay']", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan kolom NomorRekeningMutasi :", err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		if err = fieldNomorRekeningMutasi.Type(norek); err != nil {
			logger.Debug("Gagal mengisi kolom NomorRekeningMutasi :", err)
			return nil, page, err
		}

		btnShow2, err := iframeMainContent.WaitForSelector("input[name='show1']", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element btn show mutasi: ", err)
			return nil, page, err
		}
		page.WaitForTimeout(2000)
		if err = btnShow2.Click(); err != nil {
			logger.Debug("Gagal klik btn show mutasi :", err)
			return nil, page, err
		}

		msgError, err := iframeMainContent.WaitForSelector("body > form > div > div > div:nth-child(1) > div > div.row.BNI-White > div > b > font", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(5000)})
		if err != nil {
			logger.Debug("Rekening Ditemukan")
			iframeMain, err = page.WaitForSelector("iframe#content", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateAttached,
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan iframe main: ", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			iframeMainContent, err = iframeMain.ContentFrame()
			if err != nil {
				logger.Debug("Gagal mendapatkan iframe main content: ", err)
				return nil, page, err
			}

			rekValue, err := iframeMainContent.WaitForSelector("body > section > form > div > div > div:nth-child(1) > div > div:nth-child(2) > div > table > tbody > tr:nth-child(2) > td:nth-child(2)", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan namaPemilik :", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			namaPemilikString, err := rekValue.InnerText()
			if err != nil {
				logger.Debug("Gagal mendapatkan namaPemilikString :", err)
				return nil, page, err
			}
			nama := strings.Split(namaPemilikString, "/")[1]
			nama = strings.Replace(nama, ":", "", -1)
			nama = strings.Replace(nama, "(", "", -1)
			nama = strings.Replace(nama, ")", "", -1)
			nama = strings.Replace(nama, "IDR", "", -1)

			infoPage, err := iframeMainContent.WaitForSelector("body > section > form > div > div > div:nth-child(1) > div > div:nth-child(2) > div > div.row.mt-3 > div > ul > li:nth-child(3) > span:nth-child(4)", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan element pages :", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			pagesString, err := infoPage.InnerText()
			if err != nil {
				logger.Debug("Gagal mendapatkan totalPages string :", err)
				return nil, page, err
			}
			pageString := strings.ReplaceAll(pagesString, "\u00a0", "")
			intPages, err := strconv.Atoi(pageString)
			if err != nil {
				logger.Debug("could not conv string to int :", err)
				return nil, page, err
			}

			if intPages > 0 {
			Mutasi:
				for k := 1; k <= intPages; k++ {
					logger.Debug("proses scrape mutasi, halaman: ", k)
					i.Log(fmt.Sprintf("proses scrape mutasi, halaman: %d", k))
					iframeMain, err = page.WaitForSelector("iframe#content")
					if err != nil {
						return nil, page, err
					}
					page.WaitForTimeout(500)
					iframeMainContent, err = iframeMain.ContentFrame()
					if err != nil {
						return nil, page, err
					}

					if _, err := iframeMainContent.WaitForSelector("table.table-sm.BNI-Table-Aprisma-BorderRadius-None tr.BNI-LightAqua", playwright.PageWaitForSelectorOptions{
						Timeout: playwright.Float(60000),
					}); err == nil {
						mutasiDatas, err := iframeMainContent.QuerySelectorAll("table.table-sm.BNI-Table-Aprisma-BorderRadius-None tr.BNI-LightAqua")
						if err != nil {
							logger.Debug("Gagal mendapatkan mutasi datas: ", err)
							return nil, page, err
						}

						page.WaitForTimeout(500)
						for _, mutasiRow := range mutasiDatas {
							mutasi := &domain.Mutasi{}

							mutasiCells, err := mutasiRow.QuerySelectorAll("td")
							if err != nil {
								logger.Debug("Gagal mendapatkan mutasi row: ", err)
								return nil, page, err
							}
							for k, mutasiCell := range mutasiCells {
								mutasiCellValue, err := mutasiCell.InnerText()
								if err != nil {
									logger.Debug("Gagal mendapatkan mutasi cell value: ", err)
									return nil, page, err
								}
								switch k {
								case 1:
									mutasi = &domain.Mutasi{}
									mutasi.TipeBank = domain.BankTypeBNI
									mutasi.Rekening = nomorRekening
									mutasi.PemilikRekening = nama
									date, err := time.Parse(constant.LayoutDateTimeBNI, mutasiCellValue)
									if err != nil {
										logger.Debug("could not parsing date from mutasiCellValue :", err)
									}
									pgDate := types.PGDate{
										Time: date,
									}
									mutasi.TglBank = pgDate
								case 4:
									mutasi.Keterangan = mutasiCellValue
									//logger.Debug(mutasi.Keterangan)
								case 5:
									jumlahString := strings.Replace(mutasiCellValue, ",", "", -1)
									jumlahStringSep := strings.Split(jumlahString, ".")
									jumlahString = strings.ReplaceAll(jumlahStringSep[0], "\u00a0", "")
									jumlahString = strings.ReplaceAll(jumlahStringSep[0], "&nbsp;", "")
									jumlah, err := strconv.Atoi(jumlahStringSep[0])
									if err != nil {
										logger.Debug("could not parsing jumlah into float :", err)
									}
									mutasi.Jumlah = int64(jumlah)
								case 6:
									if mutasiCellValue == "C" {
										mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
									} else {
										mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
									}
								case 7:
									saldo, _ := mutasiCell.InnerText()
									saldo = strings.Replace(saldo, ",", "", -1)
									saldoSep := strings.Split(saldo, ".")
									saldo = strings.ReplaceAll(saldoSep[0], "\u00a0", "")
									saldo = strings.ReplaceAll(saldoSep[0], "&nbsp;", "")
									saldoInt, err := strconv.Atoi(saldoSep[0])
									if err != nil {
										logger.Debug("error parse string to int:", err)
									}
									mutasi.Saldo = int64(saldoInt)
								}
							}
							if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
								rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
							}
						}

						nextPage, err := iframeMainContent.WaitForSelector("li.page-item:nth-child(4) > a:nth-child(1) > img", playwright.PageWaitForSelectorOptions{
							Timeout: playwright.Float(2000),
						})
						if err != nil {
							break Mutasi
						} else {
							nextPage.Click()
							page.WaitForTimeout(2000)
						}
					}
				}
			}
		} else {
			msg, _ := msgError.InnerText()
			logger.Debug(strings.TrimPrefix(strings.Split(msg, "-")[2], " "))
			i.Log(strings.TrimPrefix(strings.Split(msg, "-")[2], " "))
		}
	} else {
		i.Log("Rekening Tidak Ditemukan.")
		logger.Debug(nomorRekeningNotFound.InnerText())
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	i.Log("proses scraping selesai.")
	logger.Debug("proses scraping selesai.")

	topFrame, err := page.WaitForSelector("iframe[name='topFrame1']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug("Error get top frame element :", err)
		return nil, page, err
	}
	topFrameContent, err := topFrame.ContentFrame()
	if err != nil {
		logger.Debug("Error get top frame content :", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)

	backToHome, err := topFrameContent.WaitForSelector("a[title='Home']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("error get back to home element :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = backToHome.Click(); err != nil {
		logger.Debug("error click back to home :", err)
	}
	page.WaitForTimeout(1000)
	closeAccountInformation, err := iframeSidebarContent.WaitForSelector("div[id='parent-MNU_GCME_040000']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err == nil {
		if err = closeAccountInformation.Click(); err != nil {
			logger.Debug("error click close account information :", err)
		}
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) isLoginBNIDirect(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input#corpId_placeholder", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) LogoutBNIDirect(page playwright.Page) error {
	iframeHeader, err := page.WaitForSelector("iframe[name='topFrame1']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe header: ", err)
		return err
	}
	iframeHeaderContent, err := iframeHeader.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan iframe header content: ", err)
		return err
	}
	time.Sleep(time.Second)
	i.Log("proses logout")
	logger.Debug("proses logout")

	btnLogout, err := iframeHeaderContent.WaitForSelector("a[title='Logout']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan btn loginStatus: ", err)
		return err
	}
	if err = btnLogout.Click(); err != nil {
		logger.Debug("Gagal klik btn loginStatus :", err)
		return err
	}
	if err := page.Close(); err != nil {
		logger.Debug(err)
	}
	i.Log("logout berhasil")
	logger.Debug("logout successfully.")

	return nil
}

func (i *Ibanking) closePopupBNIDirect(page playwright.Page) error {
	popup, err := page.WaitForSelector("#checkBox1", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1000),
	})
	if err != nil {
		//logger.Debug(err)
		return err
	}
	if err := popup.Click(); err != nil {
		logger.Debug(err)
		return err
	}
	btnOk, err := page.WaitForSelector("#btnCheck1", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	if err := btnOk.Click(); err != nil {
		logger.Debug(err)
		return err
	}
	return nil
}

func (i *Ibanking) backToHomeBNIDirect(page playwright.Page) error {
	page.WaitForTimeout(2000)
	iframeHead, err := page.WaitForSelector("iframe[name=\"topFrame1\"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(1000)
	iframeHeadContent, err := iframeHead.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return err
	}
	homeBtn, err := iframeHeadContent.WaitForSelector("#myTop > div > div > a.glyphicon.glyphicon-home.border-left.border-right.border-secondary.px-3.py-1.BNI-TextColor-Black", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	if err := homeBtn.Click(); err != nil {
		logger.Debug(err)
		return err
	}
	i.Log("terjadi error, kembali ke menu utama")
	logger.Debug("terjadi error, kembali ke menu utama")

	return nil
}

func (i *Ibanking) resultBNIDirect(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.scrapeBNIDirect(
				page,
				rekOp,
				i.BankAccount.TotalCekHari.Int64,
			)
			if err != nil {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
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
				logger.Debug("could not Fetch mutasi from database : ", err)
			}
			diff := utils.DifferenceMutasi(i.RemoveDuplicateMutasi(result.Mutasi), mutasis)
			if len(diff) > 0 {
				if err := i.UcMutasi.BulkInsertRekOnpay(context.Background(), diff); err != nil {
					logger.Debug("could not bulk insert mutasi into database :", err)
				} else {
					logger.Debug("bulk insert mutasi success.")
					result.Mutasi = nil
				}
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
		}
	}

	if len(i.BankAccount.RekGriyabayar) > 0 {
		for _, rekGb := range i.BankAccount.RekGriyabayar {
			result, page1, err := i.scrapeBNIDirect(
				page,
				rekGb,
				i.BankAccount.TotalCekHari.Int64,
			)
			if err != nil {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
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
				logger.Debug("could not Fetch mutasi from database : ", err)
			}
			diff := utils.DifferenceMutasi(i.RemoveDuplicateMutasi(result.Mutasi), mutasis)
			if len(diff) > 0 {
				if err := i.UcMutasi.BulkInsertRekGriyabayar(context.Background(), diff); err != nil {
					logger.Debug("could not bulk insert mutasi into database :", err)
				} else {
					result.Mutasi = nil
				}
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
		}
	}
	return page, nil
}
