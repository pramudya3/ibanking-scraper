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

const linkBCAPersonal = "https://ibank.klikbca.com"

func (i *Ibanking) StartBCAPersonal() {
Login:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.loginBCAPersonal(page, i.BankAccount.UserId.String, i.BankAccount.Password.String)
			if err != nil {
				time.Sleep(time.Second)
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("terjadi error, mengulang login")
					page.Close()
					goto Login
				}
			} else {
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
			Scrape:
				if len(browser.Contexts()) != 0 {
					page, err := i.resultBCAPersonal(page)
					if err != nil {
						time.Sleep(time.Second)
						if len(browser.Contexts()) != 0 {
							i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
							i.Log("terjadi error, mencoba scraping ulang")
							if err := i.backToHomeBCAPersonal(page); err != nil {
								if len(browser.Contexts()) != 0 {
									i.UpdateLoginStatus(domain.BelumLogin)
									i.Log("tidak bisa kembali ke menu utama, logout...")
									if err := i.logoutBCAPersonal(page); err != nil {
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
							i.logoutBCAPersonal(page)
							i.Log("Browser masih terbuka")
						}
						i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
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

func (i *Ibanking) loginBCAPersonal(page playwright.Page, userID, password string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)
	i.Log("menuju website BCA Personal")
	logger.Debug("menuju website BCA Personal")

	if _, err := page.Goto(linkBCAPersonal, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(100000),
	}); err != nil {
		logger.Debug("Gagal menuju :", err)
		return page, err
	}

	i.Log("proses login")
	logger.Debug("proses login")

	inputUserId, err := page.WaitForSelector("input#user_id", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom userId :", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = inputUserId.Type(userID); err != nil {
		logger.Debug("Gagal mengetik userId :", err)
		return page, err
	}
	i.Log("mengetik userId")
	logger.Debug("mengetik userId")

	inputPassword, err := page.WaitForSelector("input#pswd", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom password :", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = inputPassword.Type(password); err != nil {
		logger.Debug("Gagal mengetik password :", err)
	}
	i.Log("mengetik password.")
	logger.Debug("mengetik password")

	btnLogin, err := page.WaitForSelector("input[value='LOGIN']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(2000)})
	if err != nil {
		logger.Debug("Gagal mendapatkan btn login:", err)
		return page, err
	}
	page.WaitForTimeout(2000)
	if err = btnLogin.Click(); err != nil {
		logger.Debug("Gagal login :", err)
		return page, err
	}
	popup, err := page.WaitForEvent("dialog", playwright.PageWaitForEventOptions{
		Timeout: playwright.Float(4000),
	})
	if popup != nil {
		dialog := popup.(playwright.Dialog)
		if strings.Contains(dialog.Message(), "Anda dapat melakukan login kembali setelah 5 menit") {
			i.UpdateLoginStatus(domain.BelumLogin)
			i.Log(dialog.Message())
			logger.Debug(dialog.Message())
			i.Log(fmt.Sprintf("menunggu %d menit", i.BankAccount.IntervalCek.Int64))
			logger.Debug(fmt.Sprintf("menunggu %d menit", i.BankAccount.IntervalCek.Int64))
			time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
			return page, errors.New(dialog.Message())
		} else {
			i.UpdateLoginStatus(domain.TidakAktif)
			logger.Debug(dialog.Message())
			i.ErrorLog(dialog.Message())
			i.StopScrape()
			return nil, nil
		}
		dialog.Dismiss()
	}

	//if i.isSslVisible(page) {
	//	i.skipSSLBCA(page)
	//	page, err := i.btnLoginBCAPersonal(page, userID, password)
	//	if err != nil {
	//		return page, err
	//	}
	//} else {
	//page, err := i.btnLoginBCAPersonal(page, userID, password)
	//	if err != nil {
	//		return page, err
	//	}
	//}

	return page, nil
}

func (i *Ibanking) scrapeBCAPersonal(page playwright.Page, nomorRekening string, totalCheckDate int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesScraping)
	i.Log("mulai proses scraping")
	logger.Debug("mulai proses scraping")

	frameAtm, err := page.WaitForSelector("frame[name='atm']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan frame atm :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	contentFrameAtm, err := frameAtm.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan contentFrameAtm :", err)
		return nil, page, err
	}

	page.WaitForTimeout(1000)
	nama, err := contentFrameAtm.WaitForSelector("body > table:nth-child(7) > tbody:nth-child(1) > tr:nth-child(3) > td:nth-child(1) > center:nth-child(1) > b:nth-child(1) > font", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan nama :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	msg, err := nama.InnerText()
	if err != nil {
		logger.Debug("Gagal mendapatkan namaPemilik :", err)
		return nil, page, err
	}

	pemilikRekening := strings.ToTitle(strings.Split(msg, ",")[0])
	logger.Debug(pemilikRekening)

	frameMenu, err := page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameMenuContent, err := frameMenu.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu content:", err)
		return nil, page, err
	}

	menuInfoRek, err := frameMenuContent.WaitForSelector("a[href='account_information_menu.htm']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(40000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan menu info rek content:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuInfoRek.Click(); err != nil {
		logger.Debug("Klik error :", err)
		return nil, page, err
	}

	frameMenu, err = page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameMenuContent, err = frameMenu.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu content:", err)
		return nil, page, err
	}

	menuInfoSaldo, err := frameMenuContent.WaitForSelector("a[onclick=\"javascript:goToPage('balanceinquiry.do');return false;\"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan menu info saldo:", err)
		return nil, page, err
	}
	page.WaitForTimeout(2000)
	if err = menuInfoSaldo.Click(); err != nil {
		logger.Debug("Gagal Klik menu info saldo :", err)
		return nil, page, err
	}

	frameDetail, err := page.WaitForSelector("frame[name='atm']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element frame detail:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameDetailContent, err := frameDetail.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan element frame detail content:", err)
		return nil, page, err
	}

	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	rekening := &domain.Rekening{}
	rekening.TipeBank = domain.BankTypeBCA
	rekening.PemilikRekening = pemilikRekening

	if _, err = frameDetailContent.WaitForSelector("table:nth-of-type(3) tr"); err != nil {
		logger.Debug("Gagal mendapatkan table saldo:", err)
		return nil, page, err
	} else {
		page.WaitForTimeout(5000)
		tableRowSaldo, err := frameDetailContent.QuerySelectorAll("table:nth-of-type(3) tr")
		if err != nil {
			logger.Debug("Gagal mendapatkan table saldo:", err)
			return nil, page, err
		}

		for tableSaldoRowIndex, rowSaldo := range tableRowSaldo {
			if tableSaldoRowIndex > 0 {
				saldoCells, err := rowSaldo.QuerySelectorAll("td font")
				if err != nil {
					logger.Debug("Gagal mendapatkan saldo cell:", err)
					return nil, page, err
				}
				for saldoCellIndex, saldoCell := range saldoCells {
					valueCell, err := saldoCell.InnerText()
					if err != nil {
						logger.Debug("Gagal mendapatkan cell value:", err)
						return nil, page, err
					}

					switch saldoCellIndex {
					case 0:
						rekening.Rekening = valueCell
					case 3:
						rekening.SaldoStr = valueCell
						saldo := strings.Replace(valueCell, ",", "", -1)
						saldoSplit := strings.Split(saldo, ".")
						saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "\u00a0", "")
						saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "&nbsp;", "")
						saldoInt, err := strconv.ParseInt(saldoSplit[0], 0, 64)
						if err != nil {
							logger.Debug("parse to int64 error :", err)
						}
						rekening.Saldo = saldoInt
					}
				}
			}
		}

		if rekening.Saldo > 0 {
			rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
		}
	}

	i.Log("Daftar Rekening: ")
	no := 1
	for _, rek := range rekeningMutasiWraper.Rekening {
		i.Log(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", no, rek.Rekening, rek.PemilikRekening))
		logger.Debug(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", no, rek.Rekening, rek.PemilikRekening))
		no++
	}
	i.Log("proses scraping rekening saldo")
	logger.Debug("proses scraping rekening saldo")

	i.Log(fmt.Sprintf("rekening: %s, a/n: %s, update saldo menjadi: Rp%s", nomorRekening, pemilikRekening, rekening.SaldoStr))
	logger.Debug(fmt.Sprintf("rekening: %s, a/n: %s, update saldo menjadi: Rp%s", nomorRekening, pemilikRekening, rekening.SaldoStr))

	frameMenu, err = page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameMenuContent, err = frameMenu.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan frame menu content:", err)
		return nil, page, err
	}

	menuMutasiRekening, err := frameMenuContent.WaitForSelector("a[onclick=\"javascript:goToPage('accountstmt.do?value(actions)=acct_stmt');return false;\"]", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan menu info saldo:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuMutasiRekening.Click(); err != nil {
		logger.Debug("Klik error :", err)
		return nil, page, err
	}

	frameDetail, err = page.WaitForSelector("frame[name='atm']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element frame detail:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	frameDetailContent, err = frameDetail.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan element frame detail content:", err)
		return nil, page, err
	}

	noRekOption, err := frameDetailContent.WaitForSelector("select#D1", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan nomor rekening option:", err)
		return nil, page, err
	}

	page.WaitForTimeout(4000)
	reks, err := frameDetailContent.QuerySelectorAll("select#D1 option")
	if err != nil {
		logger.Debug(err)
	}
	for _, rek := range reks {
		page.WaitForTimeout(2000)
		norek, err := rek.InnerText()
		if err != nil {
			logger.Debug(err)
		}
		if norek == nomorRekening {
			val, err := rek.GetAttribute("value")
			if err != nil {
				logger.Debug(err)
			}
			if val, err := noRekOption.SelectOption(playwright.SelectOptionValues{
				Values: playwright.StringSlice(val),
			}); err != nil {
				logger.Debug(err)
			} else {
				logger.Debug(val)
			}
			logger.Debug("rekening selected")

			now := time.Now().AddDate(0, 0, -int(totalCheckDate))

			dayFormated := now.Format("02")
			i.Log(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, rekening.PemilikRekening, now.Format("02/01/2006")))
			logger.Debug(fmt.Sprintf("proses scraping mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, rekening.PemilikRekening, now.Format("02/01/2006")))

			selectDay, err := frameDetailContent.WaitForSelector("#startDt", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan startDate :", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			//if err = selectDay.Click(); err != nil {
			//	logger.Debug("Gagal klik pilih tanggal :", err)
			//	return nil, page, err
			//}
			if val, err := selectDay.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(dayFormated)}); err != nil {
				logger.Debug("Gagal mengetik tanggal mulai :", err)
				return nil, page, err
			} else {
				logger.Debug(val)
			}

			btnSubmit, err := frameDetailContent.WaitForSelector("input[name=\"value(submit1)\"]", playwright.PageWaitForSelectorOptions{ //input[name="value(submit1)"]
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan btn lihat mutasi rekening:", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			if err = btnSubmit.Click(); err != nil {
				logger.Debug("Gagal Klik btnLihatMutasiRekening :", err)
				return nil, page, err
			}

			frameDetail, err = page.WaitForSelector("frame[name='atm']", playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(60000),
			})
			if err != nil {
				logger.Debug("Gagal mendapatkan element frame detail:", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			frameDetailContent, err = frameDetail.ContentFrame()
			if err != nil {
				logger.Debug("Gagal mendapatkan element frame detail content:", err)
				return nil, page, err
			}

			nama, err := frameDetailContent.WaitForSelector("body > table:nth-child(4) > tbody > tr:nth-child(1) > td > table > tbody > tr:nth-child(3) > td:nth-child(3) > font", playwright.PageWaitForSelectorOptions{
				State: playwright.WaitForSelectorStateVisible,
			})
			if err != nil {
				logger.Debug("gagal mendapatkan element nama :", err)
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			pemilikRekening, err := nama.InnerText()
			if err != nil {
				logger.Debug("gagal mendapatkan pemilik rekening :", err)
				return nil, page, err
			}

			tahun, err := frameDetailContent.WaitForSelector("body > table:nth-child(4) > tbody > tr:nth-child(1) > td > table > tbody > tr:nth-child(4) > td:nth-child(3) > font")
			if err != nil {
				logger.Debug("could not get tahun")
				return nil, page, err
			}
			page.WaitForTimeout(1000)
			tahunString, err := tahun.InnerText()
			if err != nil {
				logger.Debug("gagal mendapat tahunString element :", err)
				return nil, page, err
			}
			year := strings.Split(tahunString, "/")

			page.WaitForTimeout(5000)
			dataMutasiRows, err := frameDetailContent.QuerySelectorAll("table:nth-of-type(3) tr:nth-of-type(2) table tr")
			if err != nil {
				logger.Debug("Gagal mendapatkan table mutasi:", err)
				return nil, page, err
			}

			i.Log("Proses scraping mutasi")
			logger.Debug("Proses scraping mutasi")
			for _, dataMutasiRow := range dataMutasiRows[1:] {
				mutasi := &domain.Mutasi{}

				mutasiCells, err := dataMutasiRow.QuerySelectorAll("td font")
				if err != nil {
					logger.Debug("Gagal mendapatkan mutasi cell:", err)
					return nil, page, err
				}
			rowMutasi:
				for mutasiCellIndex, mutasiCell := range mutasiCells {
					mutasiCellValue, err := mutasiCell.InnerText()
					if err != nil {
						logger.Debug("Gagal mendapatkan mutasi cell value:", err)
						return nil, page, err
					}

					switch mutasiCellIndex {
					case 0:
						mutasi = &domain.Mutasi{}
						mutasi.TipeBank = domain.BankTypeBCA
						mutasi.Rekening = nomorRekening
						mutasi.PemilikRekening = pemilikRekening
						if mutasiCellValue == "Tgl." {
							continue rowMutasi
						}
						if mutasiCellValue == "PEND" {
							pgDate := types.PGDate{Time: time.Now()}
							mutasi.TglBank = pgDate
						} else {
							tanggal := strings.Replace(mutasiCellValue, "/", "-", -1)
							tanggalBank := tanggal + "-" + year[4]
							date, err := time.Parse(constant.LayoutDateISO8602, tanggalBank)
							if err != nil {
								logger.Debug("gagal parse date :", err)
							}
							pgDate := types.PGDate{Time: date}
							mutasi.TglBank = pgDate
						}
					case 1:
						if mutasiCellValue == "Keterangan" {
							continue rowMutasi
						} else {
							keterangan := strings.Replace(mutasiCellValue, "\n", " ", -1)
							mutasi.Keterangan = keterangan
						}
					case 3:
						if mutasiCellValue == "Mutasi" {
							continue
						} else {
							jumlah := strings.Replace(mutasiCellValue, ",", "", -1)
							jumlahSplit := strings.Split(jumlah, ".")
							jumlahSplit[0] = strings.ReplaceAll(jumlahSplit[0], "\u00a0", "")
							jumlahSplit[0] = strings.ReplaceAll(jumlahSplit[0], "&nbsp;", "")
							jumlahInt, err := strconv.ParseInt(jumlahSplit[0], 0, 64)
							if err != nil {
								logger.Debug("parsing from string to int64 error :", err)
							}
							mutasi.Jumlah = jumlahInt
						}
					case 4:
						if mutasiCellValue != "CR" {
							mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
						} else {
							mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
						}
					case 5:
						if mutasiCellValue == "Saldo" {
							continue
						} else {
							saldoString := strings.Replace(mutasiCellValue, ",", "", -1)
							saldoSplit := strings.Split(saldoString, ".")
							saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "\u00a0", "")
							saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "&nbsp;", "")
							saldoInt, err := strconv.ParseInt(saldoSplit[0], 0, 64)
							if err != nil {
								logger.Debug("parsing from string to int64 error :", err)
							}
							mutasi.Saldo = saldoInt
						}
					}
				}
				if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
					rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
				}
			}
		} else {
			logger.Debug("rekening tidak ditemukan")
			i.Log("Rekening tidak ditemukan")
		}
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	i.Log("proses scraping selesai")
	logger.Debug("proses scraping selesai")

	frameMenu, err = page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(250)
	frameMenuContent, err = frameMenu.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	backToMenu, err := frameMenuContent.WaitForSelector("body > table > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(8) > td > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendaptakan element kembali ke menu :", err)
		return nil, page, err
	}
	page.WaitForTimeout(250)
	if err = backToMenu.Click(); err != nil {
		logger.Debug("Gagal Klik kembali ke menu :", err)
		return nil, page, err
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) logoutBCAPersonal(page playwright.Page) error {
	frameHeader, err := page.WaitForSelector("frame[name='header']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan frame header:", err)
		return err
	}
	frameHeaderContent, err := frameHeader.ContentFrame()
	if err != nil {
		logger.Debug("Gagal mendapatkan frame header content:", err)
		return err
	}
	i.Log("proses logout")
	logger.Debug("proses logout")

	btnLogout, err := frameHeaderContent.WaitForSelector("div#gotohome a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan btn logout:", err)
		return err
	}
	if err = btnLogout.Click(); err != nil {
		logger.Debug("Gagal Klik btnLogout :", err)
		return err
	}

	if err := page.Close(); err != nil {
		logger.Debug(err)
	}
	i.Log("logout berhasil.")
	logger.Debug("logout berhasil.")

	return nil
}

func (i *Ibanking) isLoginBCAPersonal(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input#user_id", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) isSslVisible(page playwright.Page) bool {
	ssl, err := page.WaitForSelector("#advancedButton", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(2000)})
	if err != nil {
		logger.Debug("ssl is not visible :", err)
		return false
	}
	if visible, _ := ssl.IsVisible(); visible {
		logger.Debug("ssl is visible")
		return true
	}
	return true
}

func (i *Ibanking) skipSSLBCA(page playwright.Page) {
	ignoreSsl, err := page.WaitForSelector("#advancedButton", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(2000)})
	if err != nil {
		logger.Debug("Gagal mendapatkan advanced btn :", err)
	}
	if err = ignoreSsl.Click(playwright.ElementHandleClickOptions{Timeout: playwright.Float(250)}); err != nil {
		logger.Debug("Klik error :", err)
	}
	acceptRisk, err := page.QuerySelector("#exceptionDialogButton")
	if err != nil {
		logger.Debug("error :", err)
	} else {
		if err = acceptRisk.Click(playwright.ElementHandleClickOptions{Timeout: playwright.Float(250)}); err != nil {
			logger.Debug("Klik error :", err)
		}
	}
}

func (i *Ibanking) btnLoginBCAPersonal(page playwright.Page, userID, password string) (playwright.Page, error) {
	inputUserId, err := page.WaitForSelector("input#user_id")
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom userId :", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = inputUserId.Type(userID); err != nil {
		logger.Debug("Gagal mengetik userId :", err)
		return page, err
	}
	i.Log("mengetik userId")
	logger.Debug("mengetik userId")

	inputPassword, err := page.WaitForSelector("input#pswd")
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom password :", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = inputPassword.Type(password); err != nil {
		logger.Debug("Gagal mengetik password :", err)
	}
	i.Log("mengetik password.")
	logger.Debug("mengetik password")

	btnLogin, err := page.WaitForSelector("input[value='LOGIN']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(2000)})
	if err != nil {
		logger.Debug("Gagal mendapatkan btn login:", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = btnLogin.Click(); err != nil {
		logger.Debug("Gagal login :", err)
		return page, err
	}

	popup, err := page.WaitForEvent("dialog", playwright.PageWaitForEventOptions{
		Timeout: playwright.Float(4000),
	})
	if popup != nil {
		dialog := popup.(playwright.Dialog)
		if dialog.Message() != "" {
			i.UpdateLoginStatus(domain.TidakAktif)
			logger.Debug(dialog.Message())
			i.ErrorLog(dialog.Message())
			i.StopScrape()
			return nil, nil
		}
		dialog.Dismiss()
	}
	//btnOk, err := page.WaitForSelector("button[class='swal2-confirm bg-col-bca-blue error-btn swal2-styled']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(3000)})
	//if err != nil {
	//	break A
	//}
	//if err := btnOk.Click(); err != nil {
	//	logger.Debug(err)
	//}
	return page, nil
}

func (i *Ibanking) backToHomeBCAPersonal(page playwright.Page) error {
	page.WaitForTimeout(2000)
	frameMenu, err := page.WaitForSelector("frame[name='menu']", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		logger.Debug(err)
		return err
	}
	page.WaitForTimeout(250)
	frameMenuContent, err := frameMenu.ContentFrame()
	if err != nil {
		logger.Debug(err)
		return err
	}
	backToMenu, err := frameMenuContent.WaitForSelector("body > table > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(8) > td > a", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendaptakan element kembali ke menu :", err)
		return err
	}
	page.WaitForTimeout(250)
	if err = backToMenu.Click(); err != nil {
		logger.Debug("Gagal Klik kembali ke menu :", err)
		return err
	}
	i.Log("terjadi error, kembali ke menu utama")
	logger.Debug("terjadi error, kembali ke menu utama")

	return nil
}

func (i *Ibanking) resultBCAPersonal(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.scrapeBCAPersonal(
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
			result, page1, err := i.scrapeBCAPersonal(
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
