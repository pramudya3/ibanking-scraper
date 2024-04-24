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

const linkMandiriMCM = "https://mcm2.bankmandiri.co.id/corporate/#!/login"

func (i *Ibanking) StartMandiriMCM() {
Login:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.LoginMandiriKopra(page, i.BankAccount.CompanyId.String, i.BankAccount.UserId.String, i.BankAccount.Password.String)
			if err != nil {
				time.Sleep(time.Second)
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("Terjadi error, mencoba login ulang")
					logger.Debug("Mencoba login ulang Mandiri Kopra", err.Error())
					page.Close()
					goto Login
				} else {
					return
				}
			} else {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
			Scrape:
				if len(browser.Contexts()) != 0 {
					page, err := i.resultMandiriMCM(page)
					if err != nil {
						time.Sleep(time.Second)
						if len(browser.Contexts()) != 0 {
							i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
							i.Log("terjadi error, mencoba scraping ulang")
							if err := i.backToHomeMandiriMCM(page); err != nil {
								if len(browser.Contexts()) != 0 {
									i.UpdateLoginStatus(domain.BelumLogin)
									i.Log("tidak bisa kembali ke menu utama, logout...")
									if err := i.LogoutMandiriMCM(page); err != nil {
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
							i.LogoutMandiriMCM(page)
							i.Log("browser masih terbuka")
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

func (i *Ibanking) LoginMandiriKopra(page playwright.Page, clientID, username, password string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("menuju website Mandiri MCM")
	logger.Debug("menuju website Mandiri MCM")

	if _, err := page.Goto(linkMandiriMCM, playwright.PageGotoOptions{
		Timeout: playwright.Float(100000),
	}); err != nil {
		logger.Debug("Gagal menuju website Mandiri MCM: ", err)
		return page, err
	}

	i.Log("proses login")
	logger.Debug("proses login")

	companyIdField, err := page.WaitForSelector("#content > div > ng-include > div > div > div.col-md-4.col-sm-5 > div > form > div:nth-child(1) > div > div > input", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = companyIdField.Type(clientID); err == nil {
		i.Log("mengetik company id.")
		logger.Debug("mengetik company id")
	}

	UsernameField, err := page.WaitForSelector("#content > div > ng-include > div > div > div.col-md-4.col-sm-5 > div > form > div:nth-child(2) > div > div > input", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = UsernameField.Type(username); err == nil {
		i.Log("mengetik usermame.")
		logger.Debug("mengetik Username")
	}

	passwordField, err := page.WaitForSelector("#content > div > ng-include > div > div > div.col-md-4.col-sm-5 > div > form > div.form-group.clearfix > div > div > input", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = passwordField.Type(password); err == nil {
		i.Log("mengetik password.")
		logger.Debug("mengetik password")
	}

	btnLogin, err := page.WaitForSelector("#content > div > ng-include > div > div > div.col-md-4.col-sm-5 > div > form > button", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = btnLogin.Click(); err == nil {
		msgError, err := page.WaitForSelector("#content > div > ng-include > div > div > div.col-md-4.col-sm-5 > div > form > div.small.eroholder.clearfix.ng-isolate-scope > div.head.clearfix.alert.alert-danger.header-msg.ng-scope > h2", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(2000),
		})
		if err == nil {
			page.WaitForTimeout(1000)
			msgerror, err := msgError.InnerText()
			if err == nil {
				if strings.Contains(msgerror, "still login") {
					i.Log(msgerror)
					logger.Debug(msgerror)
					i.Log(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
					logger.Debug(fmt.Sprintf("Menunggu %d menit", i.BankAccount.IntervalCek.Int64))
					time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
					return page, errors.New(strings.TrimPrefix(msgerror, " "))
				} else {
					i.UpdateLoginStatus(domain.TidakAktif)
					logger.Debug(strings.TrimPrefix(msgerror, " "))
					i.ErrorLog(strings.TrimPrefix(msgerror, " "))
					i.StopScrape()
					return nil, nil
				}
			}
		}
	}

	return page, nil
}

func (i *Ibanking) ScrapeMandiriMCM(page playwright.Page, nomorRekening string, totalCekHari int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.Log("mulai proses scraping")
	i.UpdateLoginStatus(domain.ProsesScraping)
	logger.Debug("mulai proses scraping")

	closePopup, err := page.WaitForSelector("button[class='btn btn-primary']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(1000),
	})
	if err == nil {
		if err = closePopup.Click(playwright.ElementHandleClickOptions{
			Timeout: playwright.Float(1000),
		}); err != nil {
			logger.Debug("tidak ada popup.")
		}
	}

	Accounts, err := page.WaitForSelector("#header > header > div.navbar-menu.yamm > div > ul > li:nth-child(3) > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = Accounts.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	accountList, err := page.WaitForSelector("#header > header > div.navbar-menu.yamm > div > ul > li.dropdown.pull-left.yamm-fw.open > ul > li > div > div > ul:nth-child(1) > li:nth-child(2) > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = accountList.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	rekenings := []*domain.Rekening{}
	if _, err = page.WaitForSelector("div.tbody.tbody-vs-repeat.clearfix div.tr.clearfix.ng-scope"); err == nil {
		page.WaitForTimeout(2000)
		tableRowRekening, err := page.QuerySelectorAll("div.tbody.tbody-vs-repeat.clearfix div.tr.clearfix.ng-scope")
		if err != nil {
			logger.Debug(err)
			return nil, page, err
		}

		for _, rows := range tableRowRekening {
			rekening := &domain.Rekening{}
			rekening.TipeBank = domain.BankTypeMandiri
			datas, err := rows.QuerySelectorAll("div[class='td ']")
			if err != nil {
				logger.Debug("Gagal mendapatkan element datas :", err)
				return nil, page, err
			}

			for a, data := range datas {
				cellValue, err := data.InnerText()
				if err != nil {
					logger.Debug(err)
					return nil, page, err
				}
				switch a {
				case 0:
					norek := strings.TrimSpace(cellValue)
					norek = strings.Replace(norek, "\n", "", -1)
					rekening.Rekening = norek
				case 2:
					pemilikRekening := strings.Replace(cellValue, "\n", "", -1)
					rekening.PemilikRekening = pemilikRekening
				case 6:
					rekening.SaldoStr = cellValue
					saldo := strings.Split(cellValue, ".")
					saldo[0] = strings.Replace(saldo[0], ",", "", -1)
					saldo[0] = strings.Replace(saldo[0], "\n", "", -1)
					saldoInt, _ := strconv.ParseInt(saldo[0], 0, 64)
					rekening.Saldo = saldoInt
				}
			}
			if rekening.Saldo > 0 {
				rekenings = append(rekenings, rekening)
				if rekening.Rekening == nomorRekening {
					rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
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

	i.Log("mulai proses scraping rekening saldo")

	for _, rekening := range rekeningMutasiWraper.Rekening {
		i.Log(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
		logger.Debug(fmt.Sprintf("rekening: %s, a/n: %s update saldo menjadi: Rp%s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
		//logger.Debug(fmt.Sprintf("Rekening: %s | Pemilik Rekening: %s\nUpdate saldo menjadi: Rp%s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
	}

	// scraping mutasi saldo
	Accounts, err = page.WaitForSelector("#header > header > div.navbar-menu.yamm > div > ul > li.dropdown.pull-left.yamm-fw.active > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = Accounts.Click(); err != nil {
		logger.Debug("Gagal klik Accounts :", err)
		return nil, page, err
	}

	accountStatement, err := page.WaitForSelector("#header > header > div.navbar-menu.yamm > div > ul > li.dropdown.pull-left.yamm-fw.active.open > ul > li > div > div > ul:nth-child(3) > li.ng-scope > a > span", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = accountStatement.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	selectRekening, err := page.WaitForSelector("span.select2-chosen.ng-binding[aria-hidden='false']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = selectRekening.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = selectRekening.Type(nomorRekening); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = selectRekening.Press(`Enter`); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	page.WaitForTimeout(2000)
	nama, err := page.WaitForSelector("span[ng-if='selectValue3']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	page.WaitForTimeout(2000)
	namaString, err := nama.InnerText()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	namaPemilik := strings.TrimPrefix(namaString, " ")
	namaPemilik = strings.Replace(namaPemilik, "-", "", -1)

	norek, err := page.WaitForSelector("#content > ng-include:nth-child(2) > div.container.custom-container.ng-scope > div > section.no-print > div.content.p_20 > ng-include:nth-child(2) > form > div:nth-child(1) > div > div > div > a > span:nth-child(2) > span:nth-child(1)", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	norekString, err := norek.InnerText()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	setCustom, err := page.WaitForSelector("#content > ng-include:nth-child(2) > div.container.custom-container.ng-scope > div > section.no-print > div.content.p_20 > ng-include:nth-child(2) > form > div:nth-child(2) > div.col-md-2 > select", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if _, err = setCustom.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice("custom")}); err != nil {
		logger.Debug("Failed set to custom :", err)
		return nil, page, err
	}

	startdate := time.Now().AddDate(0, 0, -int(totalCekHari))
	startdateString := startdate.Format(constant.LayoutMandiriKopra)
	calendar, err := page.WaitForSelector("form.ng-dirty > div:nth-child(2) > div:nth-child(3) > div:nth-child(2) > md-datepicker:nth-child(1) > div:nth-child(2) > input", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = calendar.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	logger.Debug(startdateString)
	dates := fmt.Sprintf("td[aria-label='%s']", startdateString)
	page.WaitForTimeout(1000)
	// td[aria-label='Thursday August 10 2023']
	dateSelected, err := page.WaitForSelector(dates, playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = dateSelected.Click(); err != nil {
		return nil, page, err
	}
	//logger.Debugf("Scrape mutasi dari tanggal: %s", startdate.Format("02/01/2006"))
	//i.Log(fmt.Sprintf("Scrape mutasi dari tanggal: %s", startdate.Format("02/01/2006")))
	i.Log(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s dari tanggal: %s", nomorRekening, namaPemilik, startdate.Format("02/01/2006")))
	logger.Debug(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s dari tanggal: %s", nomorRekening, namaPemilik, startdate.Format("02/01/2006")))

	page.WaitForTimeout(2000)
	btnView, err := page.QuerySelectorAll("div.clearfix:nth-child(4) > div:nth-child(1) > div:nth-child(1) > button")
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	if len(btnView) > 1 {
		page.WaitForTimeout(2000)
		if err = btnView[0].Click(); err != nil {
			return nil, page, err
		}
	}

	msgError, err := page.WaitForSelector("#content > ng-include:nth-child(2) > div.container.custom-container.ng-scope > div > section:nth-child(2) > section.content.p_20 > div > div.width-table-acc-statement.ng-scope.ng-isolate-scope > div.contain-regular-table > div > div:nth-child(2) > div.no-print.clearfix.ng-isolate-scope > div", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err == nil {
		msg, err := msgError.InnerText()
		if err != nil {
			logger.Debug(err)
		}
		if strings.Contains(msg, "No Records Found") {
			i.Log("mutasi tidak ditemukan")
			logger.Debug(msg)
		} else {
			logger.Debug(msg)
			pages, err := page.WaitForSelector("#content > ng-include:nth-child(2) > div.container.custom-container.ng-scope > div > section:nth-child(2) > section.content.p_20 > div > div > div.contain-regular-table > div > div:nth-child(2) > div:nth-child(3) > nav > ul > li:nth-child(3) > a", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(2000)
			pagesString, err := pages.InnerText()
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			pageString := strings.Split(pagesString, "of")
			pageString[1] = strings.TrimSpace(pageString[1])
			totalPagesInt, err := strconv.Atoi(pageString[1])
			if err != nil {
				logger.Debug("Gagal parsing string to int :", err)
				return nil, page, err
			}
			logger.Debug(totalPagesInt)

			btnNext, err := page.WaitForSelector(".width-table-acc-statement > div:nth-child(2) > div:nth-child(1) > div:nth-child(2) > div:nth-child(3) > nav:nth-child(1) > ul:nth-child(1) > li:nth-child(4) > a", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug(err)
				return nil, page, err
			}
			page.WaitForTimeout(2000)

			var tabelMutasiSaldo []playwright.ElementHandle
			for k := 1; k <= totalPagesInt; k++ {
				logger.Debug(fmt.Sprintf("proses scraping mutasi, halaman: %d", k))
				i.Log(fmt.Sprintf("proses scraping mutasi, halaman: %d", k))

				tabelMutasiSaldo, err = page.QuerySelectorAll("div.tbody.tbody-vs-repeat.clearfix div.tr") // .table-div.table-scroll.table-striped.table-hover.table-div-non-auto .tr div[class='td']
				if err != nil {
					logger.Debug(err)
					return nil, page, err
				}

				for _, rows := range tabelMutasiSaldo {
					mutasi := &domain.Mutasi{}

					mutasiCells, err := rows.QuerySelectorAll("div[class='td']")
					if err != nil {
						logger.Debug(err)
						return nil, page, err
					}
					for k, cellValue := range mutasiCells {
						mutasiCellValue, err := cellValue.InnerText()
						if err != nil {
							logger.Debug(err)
							return nil, page, err
						}
						switch k {
						case 0:
							tanggal := strings.Split(mutasiCellValue, " ")
							date, err := time.Parse(constant.LayoutDateMandiri, tanggal[0])
							if err != nil {
								logger.Debug("Gagal parse time :", err)
							}
							pgDate := types.PGDate{Time: date}
							mutasi = &domain.Mutasi{}
							mutasi.TglBank = pgDate
							mutasi.TipeBank = domain.BankTypeMandiri
							mutasi.PemilikRekening = namaPemilik
							mutasi.Rekening = norekString
						case 1:
							ket := strings.TrimSuffix(mutasiCellValue, "\n")
							mutasi.Keterangan = ket
						case 3:
							jumlah := strings.Replace(mutasiCellValue, ",", "", -1)
							jumlah = strings.Replace(jumlah, "\n", "", -1)
							jumlahSep := strings.Split(jumlah, ".")
							jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "\u00a0", "")
							jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "&nbsp;", "")
							if jumlahSep[0] != "0" {
								jumlahInt, err := strconv.Atoi(jumlahSep[0])
								if err != nil {
									logger.Debug("Gagal parse dari string ke int :", err)
								}
								mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
								mutasi.Jumlah = int64(jumlahInt)
							} else {
								continue
							}
						case 4:
							jumlah := strings.Replace(mutasiCellValue, ",", "", -1)
							jumlah = strings.Replace(jumlah, "\n", "", -1)
							jumlahSep := strings.Split(jumlah, ".")
							jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "\u00a0", "")
							jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "&nbsp;", "")
							if jumlahSep[0] != "0" {
								jumlahInt, err := strconv.Atoi(jumlahSep[0])
								if err != nil {
									logger.Debug("Gagal parse dari string ke int :", err)
								}
								mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
								mutasi.Jumlah = int64(jumlahInt)
							} else {
								continue
							}
						case 5:
							saldo := strings.Split(mutasiCellValue, ".")
							saldo[0] = strings.Replace(saldo[0], ",", "", -1)
							saldoInt, err := strconv.Atoi(saldo[0])
							if err != nil {
								logger.Debug("Gagal parse string ke int :", err)
							}
							mutasi.Saldo = int64(saldoInt)
						}
					}
					if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
						rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
					}
				}

				page.WaitForTimeout(1500)
				if err = btnNext.Click(); err != nil {
					page.WaitForTimeout(1500)
				}
			}
		}
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	i.Log("proses scraping selesai.")
	logger.Debug("proses scraping selesai.")

	backToHome, err := page.WaitForSelector(".icon-home", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = backToHome.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) LogoutMandiriMCM(page playwright.Page) error {
	page.WaitForTimeout(2000)
	btnLogout, err := page.WaitForSelector(".nav-logout", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}

	i.Log("proses logout")
	logger.Debug("proses logout")

	page.WaitForTimeout(500)
	if err = btnLogout.Click(); err != nil {
		logger.Debug(err)
		return err
	}

	relogin, err := page.WaitForSelector("#content > div > ng-include > div > div > section.section-footer.p_20 > div > button", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		logger.Debug(err)
		return err
	} else {
		if err := relogin.Click(); err != nil {
			logger.Debug(err)
			return err
		}
		logger.Debug("logout oke")
	}

	if err = page.Close(); err != nil {
		logger.Debug(err)
	}
	logger.Debug("logout successfully.")
	i.Log("logout berhasil")

	return nil
}

func (i *Ibanking) isMandiriMCMLogin(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input[placeholder=\"Enter your Company ID\"]", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) resultMandiriMCM(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.ScrapeMandiriMCM(
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
					logger.Debug("Gagal bulk insert rekening mandiri kopra:", err)
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
			result, page1, err := i.ScrapeMandiriMCM(
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

func (i *Ibanking) backToHomeMandiriMCM(page playwright.Page) error {
	homeBtn, err := page.WaitForSelector(".icon-home", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	if err := homeBtn.Click(); err != nil {
		logger.Debug(err)
		return err
	}
	i.Log("terjadi error, kembali ke menu utama")
	logger.Debug("terjadi error, kembali ke menu utama")
	return nil
}
