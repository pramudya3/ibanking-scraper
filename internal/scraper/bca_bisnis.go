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

const linkBCABisnis = "https://vpn.klikbca.com/+CSCOE+/logon.html"

func (i *Ibanking) StartBCABisnis() {
Login1:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.Login1BCABisnis(page, i.BankAccount.CompanyId.String, i.BankAccount.UserId.String)
			if err != nil {
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("terjadi error, mengulang proses login")
					logger.Debug(fmt.Sprint("login bca bisnis error: ", err))
					page.Close()
					goto Login1
				} else {
					return
				}
			} else {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
			Menu:
				if err := i.selectMenuBCABisnis(page); err != nil {
					logger.Debug(err)
					goto Menu
				} else {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
				Login2:
					if len(browser.Contexts()) != 0 {
						page, err := i.Login2BCABisnis(page, i.BankAccount.CompanyId.String, i.BankAccount.UserId.String)
						if err != nil {
							if len(browser.Contexts()) != 0 {
								if !strings.Contains(err.Error(), "sudah aktif") {
									i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
									i.Log(fmt.Sprint("login 2 bca bisnis error"))
									page.Close()
									goto Login1
								} else {
									page.Reload()
									goto Login2
								}
							} else {
								return
							}
						} else {
							i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
						Scrape:
							if len(browser.Contexts()) != 0 {
								res3, err := i.resultBCABisnis(page)
								if err != nil {
									time.Sleep(time.Second)
									if len(browser.Contexts()) != 0 {
										i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &res3
										i.Log("Terjadi error, mencoba scraping ulang")
										if err := i.backToHomeBCABisnis(res3); err != nil {
											i.UpdateLoginStatus(domain.BelumLogin)
											i.Log("tidak bisa kembali ke menu utama, logout...")
											if err := i.softLogout(res3); err != nil {
												i.Log("proses logout gagal, menutup halaman")
												page.Close()
											}
											goto Login1
										}
										goto Scrape
									} else {
										return
									}
								} else {
									i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &res3
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
		} else {
			return
		}
	} else {
		return
	}
}

func (i *Ibanking) Login1BCABisnis(page playwright.Page, corpID, username string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("menuju website BCA Bisnis")
	logger.Debug("menuju website BCA Bisnis")

	if _, err := page.Goto(linkBCABisnis, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateCommit,
	}); err != nil {
		logger.Debug("Gagal menuju website BCA Bisnis :", err)
		return page, err
	}

	i.Log("proses login ke-1")
	logger.Debug("proses login 1")

	userCorp := corpID + username
	userCorp = strings.TrimSpace(userCorp)

	usernameField, err := page.WaitForSelector("input#username", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element usernameField :", err)
		page.Reload()
		return page, err
	}
	if err = usernameField.Type(userCorp); err == nil {
		logger.Debug("mengetik username.")
		i.Log("mengetik username")
	}
	i.Log("Menunggu token")
	var token string
	interval := 2 * time.Second
	maxDuration := time.Duration(int(i.BankAccount.IntervalCek.Int64)) * time.Minute
	maxAttempt := int(maxDuration.Seconds() / interval.Seconds())
Attempt:
	for attempt := 1; attempt <= maxAttempt; attempt++ {
		if value, _ := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil {
			token1, err := i.UcBankAccount.GetToken(context.Background(), uint64(i.BankAccount.ID))
			if err != nil {
				logger.Debug(err)
			} else {
				if token1.String != "" {
					i.UpdateLoginStatus(domain.ProsesLogin)
					token = token1.String
					break Attempt
				} else {
					i.UpdateLoginStatus(domain.ButuhToken)
				}
			}
			time.Sleep(interval)
		} else {
			i.UpdateLoginStatus(domain.TidakAktif)
			break Attempt
		}
	}

	if token != "" {
		logger.Debug(token)
		passwordField, err := page.WaitForSelector("#password_input", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element passwordField :", err)
			return page, err
		}
		page.WaitForTimeout(250)
		if err = passwordField.Type(token); err == nil {
			i.Log("mengetik token.")
			logger.Debug("mengetik token.")
		}

		if err := i.UcBankAccount.DeleteToken(context.Background(), token, uint64(i.BankAccount.ID)); err != nil {
			logger.Error(err)
			return page, err
		}

		btnLogin, err := page.WaitForSelector("#submit", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element btnLogin :", err)
			return page, err
		}
		page.WaitForTimeout(1000)
		if err = btnLogin.Click(); err == nil {
			logger.Debug("login clicked.")
			forceLogin, err := page.WaitForSelector("body > center > table > tbody > tr > td > table > tbody > tr:nth-child(3) > td:nth-child(1) > a", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(2000),
			})
			if err != nil {
				msgError, err := page.WaitForSelector("#swal2-content", playwright.PageWaitForSelectorOptions{
					Timeout: playwright.Float(1000),
				})
				if err == nil {
					errorMsg, err := msgError.TextContent()
					if err == nil {
						if errorMsg == "Username atau password wajib untuk diisi" {
							return page, errors.New("Username atau password wajib diisi.")
						}
						if errorMsg == "Username atau password tidak sesuai" {
							logger.Debug(errorMsg)
							i.UpdateLoginStatus(domain.TidakAktif)
							i.ErrorLog(errorMsg)
							i.StopScrape()
							return nil, nil
						}
					}
				}
			} else {
				if err := forceLogin.Click(); err != nil {
					logger.Debug(err)
				}
				logger.Debug("force login clicked")
			}
		}
	} else {
		return page, errors.New("Token kosong")
	}

	return page, nil
}

func (i *Ibanking) Login2BCABisnis(page playwright.Page, corpId, username string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("proses login ke-2 BCA Bisnis")
	logger.Debug("proses login 2")

	page.WaitForTimeout(1000)
	corpIdfield, err := page.WaitForSelector("input[name='corp_cd']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element corpIdField :", err)
		return page, err
	}
	if err = corpIdfield.Type(corpId); err == nil {
		logger.Debug("mengetik company id")
		i.Log("mengetik company id.")
	}

	usernameField, err := page.WaitForSelector("input[name='user_cd']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element usernameField :", err)
		return page, err
	}
	if err = usernameField.Type(username); err == nil {
		logger.Debug("mengetik username")
		i.Log("mengetik username.")
	}

	var token string
	interval := 2 * time.Second
	maxDuration := time.Duration(int(i.BankAccount.IntervalCek.Int64)) * time.Minute
	maxAttempt := int(maxDuration.Seconds() / interval.Seconds())
Attempt:
	for attempt := 1; attempt <= maxAttempt; attempt++ {
		if value, _ := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil {
			token1, err := i.UcBankAccount.GetToken(context.Background(), uint64(i.BankAccount.ID))
			if err != nil {
				logger.Debug(err)
			} else {
				if token1.String != "" {
					i.UpdateLoginStatus(domain.ProsesLogin)
					token = token1.String
					break Attempt
				} else {
					i.UpdateLoginStatus(domain.ButuhToken)
				}
			}
			time.Sleep(interval)
		} else {
			i.UpdateLoginStatus(domain.TidakAktif)
			break Attempt
		}
	}

	passwordField, err := page.WaitForSelector("input[name='pswd']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element passwordField :", err)
		return page, err
	}

	if token != "" {
		page.WaitForTimeout(250)
		if err = passwordField.Type(token); err == nil {
			i.Log("mengetik token.")
			logger.Debug("mengetik token.")
		}
		if err := i.UcBankAccount.DeleteToken(context.Background(), token, uint64(i.BankAccount.ID)); err != nil {
			logger.Debug(err)
		}

		btnLogin2, err := page.WaitForSelector("img[name='Image13']", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element btnLogin2 :", err)
			return page, err
		}
		page.WaitForTimeout(1000)
		if err = btnLogin2.Click(); err == nil {
			dialog, err := page.WaitForEvent("dialog", playwright.PageWaitForEventOptions{
				Timeout: playwright.Float(2000),
			})
			if err == nil {
				if dialog != nil {
					newDialog := dialog.(playwright.Dialog)
					if strings.Contains(newDialog.Message(), "User ID Anda sudah aktif") {
						i.Log(newDialog.Message())
						logger.Debug(newDialog.Message())
						i.Log(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						logger.Debug(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
						newDialog.Dismiss()
						time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
						return page, errors.New(newDialog.Message())
					} else if strings.Contains(newDialog.Message(), "User ID telah diblokir.") || strings.Contains(newDialog.Message(), "Corporate ID/User ID/angka yang anda masukkan dari KeyBCA salah") {
						i.UpdateLoginStatus(domain.TidakAktif)
						i.ErrorLog(newDialog.Message())
						logger.Debug(newDialog.Message())
						i.StopScrape()
						return nil, nil
					}
					page.Keyboard().Press(`Enter`)
				}
			}
			i.Log("login ke-2 berhasil.")
		}
	}

	return page, nil
}

func (i *Ibanking) ScrapeBCABisnis(page playwright.Page, nomorRekening string, totalCekHari int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.Log("mulai proses scraping")
	i.UpdateLoginStatus(domain.ProsesScraping)
	logger.Debug("mulai proses scraping")

	page.WaitForTimeout(1000)
	leftFrame, err := page.WaitForSelector("html > frameset > frameset > frame:nth-child(1)")
	if err != nil {
		logger.Debug("Gagal mendapatkan element leftFrame :", err)
		return nil, page, err
	}
	leftFrameContent, err := leftFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapat element leftFrameContent :", err)
		return nil, page, err
	}
	leftFrameContent.WaitForTimeout(1000)

	informasiRekening, err := leftFrameContent.WaitForSelector("#divFold1 > a")
	if err != nil {
		logger.Debug("Gagal mendapatkan element informasiRekening :", err)
		return nil, page, err
	}
	if err = informasiRekening.Click(); err != nil {
		logger.Debug("Gagal klik inforamsiRekening :", err)
		return nil, page, err
	}

	page.WaitForTimeout(2000)
	informasiSaldo, err := leftFrameContent.WaitForSelector("#divFoldSub1_0 > a")
	if err != nil {
		logger.Debug("Gagal mendapat element informasiSaldo :", err)
		return nil, page, err
	}
	if err = informasiSaldo.Click(); err != nil {
		logger.Debug("Gagal klik informasiSaldo :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	workspaceFrame, err := page.WaitForSelector("html > frameset > frameset > frame:nth-child(3)")
	if err != nil {
		logger.Debug("Gagal mendpatkan workspaceFrame :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	workspaceFrameContent, err := workspaceFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan element workspaceFrameContent :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	// daftar rekening = "table.clsForm tr.clsEven td[width='41%']"\
	daftarRek, err := workspaceFrameContent.QuerySelectorAll("table.clsForm tr.clsEven")
	if err != nil {
		logger.Debug(err)
	}

	reks := []*DaftarRekening{}
	for _, rows := range daftarRek {
		rek := &DaftarRekening{}
		datas, err := rows.QuerySelectorAll("td[width='41%']")
		if err != nil {
			logger.Debug(err)
		}
		for _, cell := range datas {
			rek = &DaftarRekening{}
			cellValue, err := cell.InnerText()
			if err != nil {
				logger.Debug(err)
			}
			akunStr := strings.Replace(cellValue, "(", "", -1)
			akunStr = strings.Replace(akunStr, ")", "", -1)
			akunStr = strings.Replace(akunStr, "Rp", "", -1)
			akuns := strings.Split(akunStr, "/")
			for a, akun := range akuns {
				switch a {
				case 0:
					norek := strings.Replace(akun, "-", "", -1)
					rek.Rekening = norek
				case 1:
					rek.PemilikRekening = strings.TrimPrefix(akun, " ")
				}
			}
			reks = append(reks, rek)
		}
	}

	logger.Debug("Daftar Rekening:")
	i.Log("Daftar Rekening:")
	nomor := 1
	for _, rekening := range reks {
		i.Log(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", nomor, rekening.Rekening, rekening.PemilikRekening))
		logger.Debug(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", nomor, rekening.Rekening, rekening.PemilikRekening))
		nomor++
	}

	if _, err = workspaceFrameContent.WaitForSelector("input#AcctGrp"); err == nil {
		checks, err := workspaceFrameContent.QuerySelectorAll("input#AcctGrp")
		if err != nil {
			logger.Debug(err)
			return nil, page, err
		}
	chooseRek:
		for i := 0; i < len(checks); i++ {
			text, err := checks[i].GetAttribute("value")
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			} else {
				if strings.Contains(text, nomorRekening) {
					if err = checks[i].Click(); err != nil {
						logger.Debug(err)
						return nil, page, err
					}
					break chooseRek
				}
			}
		}
	}

	btnKirim, err := workspaceFrameContent.WaitForSelector("#Submit")
	if err != nil {
		logger.Debug("Gagal mendapatkan element btnKirim :", err)
		return nil, page, err
	}
	if err = btnKirim.Click(); err != nil {
		logger.Debug("Gagal klik btnKirim :", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)

	workspaceFrame, err = page.WaitForSelector("html > frameset > frameset > frame:nth-child(3)")
	if err != nil {
		logger.Debug("Gagal mendpatkan workspaceFrame :", err)
		return nil, page, err
	}

	workspaceFrameContent, err = workspaceFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan element workspaceFrameContent :", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)

	i.Log("proses scraping rekening saldo.")
	logger.Debug("proses scraping rekening saldo.")

	rekening := &domain.Rekening{}
	rekening.TipeBank = domain.BankTypeBCA
	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	if _, err = workspaceFrameContent.WaitForSelector("table.clsForm tr.clsEven td"); err == nil {
		rowRekenings, err := workspaceFrameContent.QuerySelectorAll("table.clsform tr.clsEven td") //tr.clsEven
		if err != nil {
			logger.Debug("Gagal mendapat element tabelRekenings :", err)
			return nil, page, err
		}
		for a, dataRek := range rowRekenings {
			data, err := dataRek.InnerText()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			if len(rowRekenings) == 3 {
				switch a {
				case 0:
					caseZero := strings.Replace(data, "-", "", -1)
					rekening.Rekening = caseZero
				case 1:
					rekening.PemilikRekening = data
				case 2:
					rekening.SaldoStr = data
					saldo := strings.Replace(data, "Rp", "", -1)
					saldo = strings.Replace(saldo, ",", "", -1)
					saldo = strings.ReplaceAll(saldo, "\u00a0", "")
					saldo = strings.ReplaceAll(saldo, "&nbsp;", "")
					caseTwo := strings.Split(saldo, ".")
					saldoInt, err := strconv.Atoi(caseTwo[0])
					if err != nil {
						logger.Debug("Gagal parse string to int :", err)
					}
					rekening.Saldo = int64(saldoInt)
				}
			} else if len(rowRekenings) == 4 {
				switch a {
				case 1:
					caseOne := strings.Replace(data, "-", "", -1)
					rekening.Rekening = caseOne
				case 2:
					rekening.PemilikRekening = data
				case 3:
					rekening.SaldoStr = data
					saldo := strings.Replace(data, "Rp", "", -1)
					saldo = strings.Replace(saldo, ",", "", -1)
					saldo = strings.ReplaceAll(saldo, "\u00a0", "")
					saldo = strings.ReplaceAll(saldo, "&nbsp;", "")
					caseThree := strings.Split(saldo, ".")
					saldoInt, err := strconv.Atoi(caseThree[0])
					if err != nil {
						logger.Debug("Gagal parse string to int :", err)
					}
					rekening.Saldo = int64(saldoInt)
				}
			}
		}
		if rekening.Saldo > 0 {
			i.Log(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: %s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
			logger.Debug(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: %s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
			rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
		}
	}

	//for _, rek := range rekeningMutasiWraper.Rekening {
	//	logger.Debug(rek)
	//}

	//menu mutasi
	informasiMutasi, err := leftFrameContent.WaitForSelector("#divFoldSub1_1 > a")
	if err != nil {
		logger.Debug("Gagal mendapat element informasiMutasi :", err)
		return nil, page, err
	}

	page.WaitForTimeout(1000)
	if err = informasiMutasi.Click(); err != nil {
		logger.Debug("Gagal klik informasiMutasi :", err)
		return nil, page, err
	}

	workspaceFrame, err = page.WaitForSelector("html > frameset > frameset > frame:nth-child(3)")
	if err != nil {
		logger.Debug("Gagal mendpatkan workspaceFrame :", err)
	}
	page.WaitForTimeout(1000)
	workspaceFrameContent, err = workspaceFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan element workspaceFrameContent :", err)
		return nil, page, err
	}

	date := time.Now().AddDate(0, 0, -int(totalCekHari))
	day := date.Format("02")
	month := date.Format("1")
	year := date.Format("2006")
	i.Log(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", rekening.Rekening, rekening.PemilikRekening, date.Format("02/01/2006")))
	logger.Debug(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", rekening.Rekening, rekening.PemilikRekening, date.Format("02/01/2006")))

	fromDay, err := workspaceFrameContent.WaitForSelector("select#from_day")
	if err != nil {
		logger.Debug("Gagal mendapat element fromDay :", err)
		return nil, page, err
	}
	if _, err = fromDay.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(day)}); err != nil {
		logger.Debug("Gagal select date :", err)
		return nil, page, err
	}

	fromMonth, err := workspaceFrameContent.WaitForSelector("select#from_mth")
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	if _, err = fromMonth.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(month)}); err != nil {
		logger.Debug("Gagal select date :", err)
		return nil, page, err
	}

	fromYear, err := workspaceFrameContent.WaitForSelector("select#from_year")
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	if _, err = fromYear.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(year)}); err != nil {
		logger.Debug("Gagal select date :", err)
		return nil, page, err
	}

	searchRekening, err := workspaceFrameContent.WaitForSelector("input#acct_display")
	if err != nil {
		logger.Debug("Gagal mendapatkan element searchRekening :", err)
		return nil, page, err
	}
	if err = searchRekening.Type(nomorRekening); err != nil {
		logger.Debug("Gagal klik searchRekening :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = searchRekening.Press(`Enter`); err != nil {
		logger.Debug("Gagal press enter :", err)
		return nil, page, err
	}

	errorMsg, err := workspaceFrameContent.WaitForSelector("h3.clsErrorMsg", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		i.Log("Mutasi Ditemukan")
		logger.Debug("Mutasi Ditemukan")
		workspaceFrame, err = page.WaitForSelector("html > frameset > frameset > frame:nth-child(3)")
		if err != nil {
			logger.Debug("Gagal mendpatkan workspaceFrame :", err)
			return nil, page, err
		}
		page.WaitForTimeout(1000)
		workspaceFrameContent, err = workspaceFrame.ContentFrame()
		if err != nil {
			logger.Debug("Gagal mendapatkan element workspaceFrameContent :", err)
			return nil, page, err
		}

		page.WaitForTimeout(2000)
		tahun, err := workspaceFrameContent.WaitForSelector("body > form > table:nth-child(1) > tbody > tr:nth-child(3) > td:nth-child(3)", playwright.PageWaitForSelectorOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(40000),
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element tahun :", err)
			return nil, page, err
		}
		page.WaitForTimeout(1000)
		tahunString, err := tahun.InnerText()
		if err != nil {
			logger.Debug("Gagal mendapatkan element tahunString :", err)
			return nil, page, err
		}
		yearString := strings.Split(strings.Split(tahunString, "-")[1], "/")[2]
		//logger.Debug(yearString)

		nama, err := workspaceFrameContent.WaitForSelector("body > form > table:nth-child(1) > tbody > tr:nth-child(2) > td:nth-child(3)", playwright.PageWaitForSelectorOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(40000),
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element nama :", err)
			return nil, page, err
		}
		page.WaitForTimeout(1000)
		namaString, err := nama.InnerText()
		if err != nil {
			logger.Debug("Gagal parse to string :", err)
			return nil, page, err
		}
		pemilik := strings.ReplaceAll(strings.Replace(namaString, ":", "", -1), "&nbsp;", "")
		//logger.Debug(pemilik)

		var tableMutasi []playwright.ElementHandle
		p := 1
	Mutasi:
		for {
			i.Log(fmt.Sprintf("proses scraping mutasi, halaman: %d", p))
			logger.Debug(fmt.Sprintf("proses scraping mutasi, halaman: %d", p))
			workspaceFrame, err = page.WaitForSelector("html > frameset > frameset > frame:nth-child(3)", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug("Gagal mendpatkan workspaceFrame :", err)
				return nil, page, err
			}
			workspaceFrameContent, err = workspaceFrame.ContentFrame()
			if err != nil {
				logger.Debug("Gagal mendapatkan element workspaceFrameContent :", err)
				return nil, page, err
			}

			tableMutasi, err = workspaceFrameContent.QuerySelectorAll("table.clsForm tr[class]")
			if err != nil {
				logger.Debug("Gagal mendapatkan element tableMutasi :", err)
				return nil, page, err
			}

			for _, rowMutasi := range tableMutasi {
				mutasi := &domain.Mutasi{}
				mutasi.TipeBank = domain.BankTypeBCA
				mutasi.Rekening = nomorRekening
				mutasi.PemilikRekening = pemilik

				cells, err := rowMutasi.QuerySelectorAll("td")
				if err != nil {
					logger.Debug("Gagal mendapatkan element cell :", err)
					return nil, page, err
				}
				for a, cell := range cells {
					mutasiCell, err := cell.InnerText()
					if err != nil {
						logger.Debug("Gagal mendapat element mutasiCell :", err)
						return nil, page, err
					}
					switch a {
					case 0:
						mutasi = &domain.Mutasi{}
						mutasi.TipeBank = domain.BankTypeBCA
						mutasi.Rekening = nomorRekening
						mutasi.PemilikRekening = pemilik
						if mutasiCell != "PEND" {
							now := time.Now()
							pgdate := types.PGDate{
								Time: now,
							}
							mutasi.TglBank = pgdate
						} else {
							tanggalBank := mutasiCell + yearString
							tanggal, err := time.Parse(constant.LayoutDateMandiri, tanggalBank)
							if err != nil {
								logger.Debug("Gagal parsing tanggal :", err)
							}
							pgDate := types.PGDate{Time: tanggal}
							mutasi.TglBank = pgDate
						}
					case 1:
						ket := strings.Replace(mutasiCell, "\n", "", -1)
						mutasi.Keterangan = ket
					case 3:
						jumlah := strings.Split(mutasiCell, ".")
						jumlah[0] = strings.Replace(jumlah[0], ",", "", -1)
						jumlah[0] = strings.ReplaceAll(jumlah[0], "&nbsp;", "")
						jumlahInt, err := strconv.Atoi(jumlah[0])
						jumlahType := jumlah[1]
						if strings.Contains(jumlahType, "CR") {
							mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
						} else if strings.Contains(jumlahType, "DB") {
							mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
						}
						if err != nil {
							logger.Debug("Gagal parsing string ke int :", err)
						}
						mutasi.Jumlah = int64(jumlahInt)
					case 4:
						saldo := strings.Split(mutasiCell, ".")
						saldo[0] = strings.Replace(saldo[0], ",", "", -1)
						saldo[0] = strings.ReplaceAll(saldo[0], "&nbsp;", "")
						saldo[0] = strings.ReplaceAll(saldo[0], "\u00a0", "")
						jumlahInt, err := strconv.Atoi(saldo[0])
						if err != nil {
							//logger.Debug("Gagal parsing string ke int :", err)
						}
						mutasi.Saldo = int64(jumlahInt)
					}
				}
				if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
					rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
				}
			}

			btnNext, err := workspaceFrameContent.WaitForSelector("input#Next", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(2000),
			})
			if err != nil {
				break Mutasi
			}
			if err := btnNext.Click(); err == nil {
				p++
			}

			page.WaitForTimeout(3000)
		}
	} else {
		i.Log("mutasi tidak ditemukan.")
		logger.Debug("mutasi tidak ditemukan.", errorMsg.String())
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	i.Log("proses scraping selesai.")
	logger.Debug("proses scraping selesai.")

	page.WaitForTimeout(1000)
	closeInformasiRekening, err := leftFrameContent.WaitForSelector("#divFold1 > a")
	if err != nil {
		logger.Debug("Gagal mendapat element closeInformasiRekening :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = closeInformasiRekening.Click(); err != nil {
		logger.Debug("Gagal klik closeInformasiRekening :", err)
		return nil, page, err
	}

	page.WaitForTimeout(1000)
	backToHomeBtn, err := leftFrameContent.WaitForSelector("#divFold0 > a")
	if err != nil {
		logger.Debug("Gagal mendapatkan element backToHomeBtn :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = backToHomeBtn.Click(); err != nil {
		logger.Debug("Gagal klik backToHomeBtn :", err)
		return nil, page, err
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) LogoutBCABisnis(page playwright.Page) error {
	topFrame, err := page.WaitForSelector("html > frameset > frame:nth-child(1)", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element topFrame :", err)
		return err
	}
	page.WaitForTimeout(500)
	topFrameContent, err := topFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapat element topFrameContent :", err)
		return err
	}

	i.Log("proses logout")
	logger.Debug("proses logout")

	btnLogout, err := topFrameContent.WaitForSelector("body > table:nth-child(2) > tbody > tr > td:nth-child(1) > a > img", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapat element btnLogout :", err)
		return err
	}
	page.WaitForTimeout(500)
	if err = btnLogout.Click(); err != nil {
		logger.Debug("Gagal klik btnLogout :", err)
		return err
	}
	i.Log("logout berhasil")
	logger.Debug("logout berhasil")
	//page.GoBack(playwright.PageGoBackOptions{})
	//btnLogout2, err := page.WaitForSelector("body > header > div.expandable > button.bare.text-white.bg-color-bca-blue.opacity-1.user-btn.ng-isolate-scope", playwright.PageWaitForSelectorOptions{
	//	State:   playwright.WaitForSelectorStateVisible,
	//	Timeout: playwright.Float(5000),
	//})
	//if err != nil {
	//	logger.Debug(err)
	//	return err
	//}
	//page.WaitForTimeout(500)
	//if err := btnLogout2.Click(); err != nil {
	//	logger.Debug(err)
	//	return err
	//}
	//logger.Debug("logout oke")

	//logout, err := page.WaitForSelector("body > div:nth-child(9) > div > button", playwright.PageWaitForSelectorOptions{
	//	State:   playwright.WaitForSelectorStateVisible,
	//	Timeout: playwright.Float(5000),
	//})
	//if err != nil {
	//	logger.Debug(err)
	//	return err
	//}
	//page.WaitForTimeout(500)
	//if err = logout.Click(); err != nil {
	//	logger.Debug(err)
	//	return err
	//} else {
	//	logger.Debug("logout successfully")
	//}
	return nil
}

func (i *Ibanking) isLogin1BCABisnis(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input[id='username']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) isLogin2BCABisnis(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input[name='corp_cd']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) resultBCABisnis(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.ScrapeBCABisnis(
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
					logger.Debug("Gagal bulk insert rekening :", err)
				} else {
					result.Rekening = nil
				}
			}

			mutasis, err := i.UcMutasi.Fetch(context.Background(), rekOp)
			if err != nil {
				logger.Debug("could not fetch mutasi from database :", err)
			}
			diff := utils.DifferenceMutasi(i.RemoveDuplicateMutasi(result.Mutasi), mutasis)
			if len(diff) > 0 {
				if err := i.UcMutasi.BulkInsertRekOnpay(context.Background(), diff); err != nil {
					logger.Debug("could not bulk insert mutasi into database :", err)
				} else {
					result.Mutasi = nil
				}
			}
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page1
		}
	}

	if len(i.BankAccount.RekGriyabayar) > 0 {
		for _, rekGb := range i.BankAccount.RekGriyabayar {
			result, page1, err := i.ScrapeBCABisnis(
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
					logger.Debug("Gagal bulk insert rekening :", err)
				} else {
					result.Rekening = nil
				}
			}

			mutasis, err := i.UcMutasi.Fetch(context.Background(), rekGb)
			if err != nil {
				logger.Debug("could not fetch mutasi from database :", err)
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

func (i *Ibanking) softLogout(page2 playwright.Page) error {
	logger.Debug("logout bca bisnis.")
	i.Log("proses logout")
	topFrame, err := page2.WaitForSelector("html > frameset > frame:nth-child(1)", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element topFrame :", err)
		return err
	}
	page2.WaitForTimeout(500)
	topFrameContent, err := topFrame.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapat element topFrameContent :", err)
		return err
	}
	btnLogout, err := topFrameContent.WaitForSelector("body > header > div.expandable > button.bare.text-white.bg-color-bca-blue.opacity-1.user-btn.ng-isolate-scope")
	if err != nil {
		logger.Debug("Gagal mendapat element btnLogout :", err)
		return err
	}
	page2.WaitForTimeout(500)
	if err = btnLogout.Click(); err != nil {
		logger.Debug("Gagal klik btnLogout :", err)
		return err
	}

	if err := page2.Close(); err != nil {
		logger.Debug(err)
	}

	i.Log("logout berhasil")
	logger.Debug("logout clicked")

	return nil
}

func (i *Ibanking) selectMenuBCABisnis(page playwright.Page) error {
	selectMenu, err := page.WaitForSelector("body > div.main > div > div > div.col-12.col-sm-12.col-md-12.col-lg-8.col-xl-6 > div.row.mt-6vh.mb-3 > div:nth-child(1) > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	} else {
		page.WaitForTimeout(1000)
		if err := selectMenu.Click(); err != nil {
			logger.Debug(err)
			return err
		}
	}
	logger.Debug("menu selected")

	return nil
}

func (i *Ibanking) backToHomeBCABisnis(page playwright.Page) error {
	frameLeft, err := page.WaitForSelector("html > frameset > frameset > frame:nth-child(1)", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(1000)
	frameLeftContent, err := frameLeft.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return err
	}
	homeBtn, err := frameLeftContent.WaitForSelector("#divFold0 > a")
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	if err := homeBtn.Click(); err != nil {
		logger.Debug(err)
		return err
	}

	closeInfoRek, err := frameLeftContent.WaitForSelector("#divFold1 > a")
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(500)
	if err := closeInfoRek.Click(); err != nil {
		logger.Debug(err)
		return err
	}
	i.Log("terjadi error, kembali ke menu utama")
	logger.Debug("back to home")
	return nil
}
