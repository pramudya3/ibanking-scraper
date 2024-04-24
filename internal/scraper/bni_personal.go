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

const linkBNIPersonal = "https://ibank.bni.co.id"

func (i *Ibanking) StartBNIPersonal() {
Login:
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		page, err := browser.NewPage()
		if err != nil {
			logger.Debug(err)
		}
		i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
		if len(browser.Contexts()) != 0 {
			page, err := i.LoginBNIPersonal(page, i.BankAccount.UserId.String, i.BankAccount.Password.String)
			if err != nil {
				time.Sleep(2 * time.Second)
				if len(browser.Contexts()) != 0 {
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
					i.Log("terjadi error, mencoba login ulang")
					page.Close()
					goto Login
				} else {
					return
				}
			} else {
			Scrape:
				i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
				page, err := i.resultBNIPersonal(page)
				if err != nil {
					if len(browser.Contexts()) != 0 {
						i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
						if strings.Contains(err.Error(), "session") {
							i.Log(err.Error())
							logger.Debug(err.Error())
							page.Close()
							time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
							goto Login
						} else {
							i.Log("terjadi error, mencoba scraping ulang")
							time.Sleep(2 * time.Second)
							if err := i.backToHomeBNIPersonal(page); err != nil {
								if len(browser.Contexts()) != 0 {
									i.UpdateLoginStatus(domain.BelumLogin)
									i.Log("tidak bisa kembali ke menu utama, logout")
									if err := i.logoutBNIPersonal(page); err != nil {
										i.Log("proses logout gagal, menutup halaman")
										page.Close()
									}
									goto Login
								} else {
									return
								}
							}
							goto Scrape
						}
					} else {
						return
					}
				} else {
					if i.BankAccount.AutoLogout.Bool {
						i.UpdateLoginStatus(domain.BelumLogin)
						i.logoutBNIPersonal(page)
						i.Log("Browser masih terbuka.")
					}
					i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
				}
			}
		} else {
			return
		}
	} else {
		return
	}
}

func (i *Ibanking) LoginBNIPersonal(page playwright.Page, userID, password string) (playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesLogin)

	if _, err := page.Goto(linkBNIPersonal, playwright.PageGotoOptions{
		Timeout: playwright.Float(100000),
	}); err != nil {
		logger.Debug("Gagal menuju :", err)
		return page, err
	}

	logger.Debug("proses login bni.")
	i.Log("proses login BNI.")

	captcha, err := page.WaitForSelector("img#IMAGECAPTCHA", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(160000),
	})
	if err != nil {
		logger.Debug("gagal mendapatkan captcha :", err)
		return page, err
	}

	if _, err = captcha.Screenshot(playwright.ElementHandleScreenshotOptions{
		Path: playwright.String("captcha.png"),
	}); err != nil {
		logger.Debug("Gagal screenshot captcha :", err)
		return page, err
	}

	err, responseJson := utils.ResolveCaptcha()
	if err != nil {
		logger.Debug("Gagal mendapat captcha: ", err)
		return page, err
	}

	username, err := page.WaitForSelector("input#AuthenticationFG\\.USER_PRINCIPAL")
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom username:", err)
		return page, err
	}
	if err = username.Type(userID); err != nil {
		logger.Debug("Gagal mengetik username :", err)
		return page, err
	}
	logger.Debug("mengetik username")
	i.Log("mengetik username")

	passwordField, err := page.WaitForSelector("input#AuthenticationFG\\.ACCESS_CODE")
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom password:", err)
		return page, err
	}

	if err = passwordField.Type(password); err != nil {
		logger.Debug("Gagal mengetik password :", err)
		return page, err
	}
	logger.Debug("mengetik password")
	i.Log("mengetik password")

	inputCaptcha, err := page.WaitForSelector("input#AuthenticationFG\\.VERIFICATION_CODE")
	if err != nil {
		logger.Debug("Gagal mendapatkan kolom captcha:", err)
		return page, err
	}
	page.WaitForTimeout(1000)
	if err = inputCaptcha.Type(responseJson.Data.Captcha); err != nil {
		logger.Debug("Gagal mengetik captcha :", err)
		return page, err
	}
	logger.Debug(responseJson.Data.Captcha)
	logger.Debug("mengetik captcha")
	i.Log("mengetik captcha")

	bahasa, err := page.WaitForSelector("select[name='AuthenticationFG.PREFERRED_LANGUAGE']")
	if err != nil {
		logger.Debug("Gagal mendapatkan bahasa selector :", err)
		return page, err
	}

	if _, err = bahasa.SelectOption(playwright.SelectOptionValues{
		Values: playwright.StringSlice("002"),
	}); err != nil {
		logger.Debug("Gagal memilih Bahasa Indonesia :", err)
		return page, err
	}

	halamanUtama, err := page.QuerySelector("select#AuthenticationFG\\.MENU_ID")
	if err != nil {
		logger.Debug("Gagal memilih halaman rekening :", err)
		return page, err
	}

	if _, err = halamanUtama.SelectOption(playwright.SelectOptionValues{
		Values: playwright.StringSlice("2"),
	}); err != nil {
		logger.Debug(err)
		return page, err
	}

	btnLogin, err := page.WaitForSelector("#VALIDATE_CREDENTIALS")
	if err != nil {
		logger.Debug("Gagal mendapatkan btnLogin :", err)
		return page, err
	}
	page.WaitForTimeout(1000)

	if err = btnLogin.Click(); err != nil {
		logger.Debug("Gagal klik btnLogin :", err)
		return page, err
	}

	msgError, err := page.WaitForSelector("div[class='redbg']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(3000)})
	if err == nil {
		respError, err := msgError.TextContent()
		if err != nil {
			logger.Debug(err)
			return page, err
		} else {
			if strings.Contains(respError, "sedang login") {
				i.UpdateLoginStatus(domain.BelumLogin)
				resp := strings.Replace(strings.Replace(strings.Join(strings.Split(respError, " ")[3:], " "), "[", "", -1), "]", "", -1)
				i.Log(resp)
				logger.Debug(resp)
				time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
				return page, errors.New(resp)
			} else {
				i.UpdateLoginStatus(domain.TidakAktif)
				logger.Debug(respError)
				i.ErrorLog(respError)
				i.StopScrape()
				return nil, errors.New("user terblokir")
			}
		}
	}

	return page, nil
}

func (i *Ibanking) scrapeBNIPersonal(page playwright.Page, nomorRekening string, totalCheckDate int64) (*domain.RekeningMutasiWraper, playwright.Page, error) {
	i.UpdateLoginStatus(domain.ProsesScraping)
	i.Log("proses scraping BNI.")
	logger.Debug("proses scraping BNI.")

	menuRekening, err := page.WaitForSelector("a#REKENING", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element menu rekening:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuRekening.Click(); err != nil {
		logger.Debug("Gagal klik menu rekening :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	menuSaldoRekening, err := page.WaitForSelector("a#Informasi-Saldo--Mutasi_Saldo-Rekening", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element menu saldo rekening:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuSaldoRekening.Click(); err != nil {
		logger.Debug("Gagal klik menu saldo rekening :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	firstName, err := page.WaitForSelector("#firstName", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element nama 1: ", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	name1, err := firstName.InnerText()
	if err != nil {
		logger.Debug("Gagal mendapatkan nama 1 string :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	lastName, err := page.WaitForSelector("#lastName", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element nama 2 :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	name2, err := lastName.InnerText()
	if err != nil {
		logger.Debug("Gagal mendapatkan name2 string :", err)
		return nil, page, err
	}
	pemilikRekening := strings.TrimPrefix(name1+name2, " ")

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	page.WaitForTimeout(4000)
	tableRek, err := page.QuerySelectorAll("table[summary='ringkasan semua rekening'] tr[id='0']")
	if err != nil {
		logger.Debug("Gagal mendapatkan tabel saldo rekening :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	rekeningMutasiWraper := &domain.RekeningMutasiWraper{}
	rekening := &domain.Rekening{}
	rekening.TipeBank = domain.BankTypeBNI
	rekening.PemilikRekening = pemilikRekening
	for _, rek := range tableRek {
		rekCells, err := rek.QuerySelectorAll("td")
		if err != nil {
			logger.Debug("Gagal mendapatkan cells of saldo rekening :", err)
			return nil, page, err
		}
		for cellIndex, cell := range rekCells {
			if cellIndex > 0 {
				cellValue, err := cell.InnerText()
				if err != nil {
					logger.Debug("Gagal mendapatkan cell value :", err)
					return nil, page, err
				}
				switch cellIndex {
				case 1:
					if len(cellValue) > 10 {
						rek := strings.Split(cellValue, "")
						norek := strings.Join(rek[7:], "")
						rekening.Rekening = norek
					} else {
						rekening.Rekening = cellValue
					}
				case 6:
					rekening.SaldoStr = cellValue
					saldo := strings.Replace(cellValue, ",", "", -1)
					saldoSep := strings.Split(saldo, ".")
					saldoSep[0] = strings.ReplaceAll(saldoSep[0], "\u00a0", "")
					saldoSep[0] = strings.ReplaceAll(saldoSep[0], "&nbsp;", "")
					saldoInt, err := strconv.ParseInt(saldoSep[0], 0, 64)
					if err != nil {
						logger.Debug("parse to Int64 Gagal :", err)
					}
					rekening.Saldo = saldoInt
				}
			}
		}
	}
	if rekening.Saldo > 0 {
		rekeningMutasiWraper.Rekening = append(rekeningMutasiWraper.Rekening, rekening)
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	i.Log("Daftar Rekening:")
	logger.Debug("Daftar Rekening:")
	no := 1
	for _, rek := range rekeningMutasiWraper.Rekening {
		i.Log(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", no, rek.Rekening, rek.PemilikRekening))
		logger.Debug(fmt.Sprintf("No: %d | Rekening: %s | Pemilik Rekening: %s", no, rek.Rekening, rek.PemilikRekening))
		no++
	}

	logger.Debug("proses scraping rekening saldo.")
	i.Log("proses scraping rekening saldo.")

	i.Log(fmt.Sprintf("Rekening: %s, a/n: %s, update saldo menjadi: Rp%s", rekening.Rekening, rekening.PemilikRekening, rekening.SaldoStr))
	logger.Debug(fmt.Sprintf("Rekening: %s, a/n: %s, update saldo menjadi: Rp%s", nomorRekening, pemilikRekening, rekening.SaldoStr))

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	menuMutasi, err := page.WaitForSelector("a#Informasi-Saldo--Mutasi_Mutasi-Tabungan--Giro", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element menu mutasi :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = menuMutasi.Click(); err != nil {
		logger.Debug("Gagal klik mutasi menu :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	historyTrx, err := page.WaitForSelector("#VIEW_TRANSACTION_HISTORY", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err := historyTrx.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	tampilkanField, err := page.WaitForSelector("img[title='Tampilkan']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan tampilkan kolom :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = tampilkanField.Click(); err != nil {
		logger.Debug("Gagal klik tampilkan kolom :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	chooseRek, err := page.WaitForSelector("#TransactionHistoryFG\\.INITIATOR_ACCOUNT", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	page.WaitForTimeout(2000)
	reks, err := page.QuerySelectorAll("#TransactionHistoryFG\\.INITIATOR_ACCOUNT option")
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	for _, rks := range reks {
		rek, err := rks.GetAttribute("value")
		if err != nil {
			logger.Debug(err)
		}
		if strings.Join(strings.Split(rek, "")[7:], "") == nomorRekening {
			if _, err := chooseRek.SelectOption(playwright.SelectOptionValues{
				Values: playwright.StringSlice(rek),
			}); err != nil {
				logger.Debug(err)
			}
		}
	}

	startDate := time.Now().AddDate(0, 0, -int(totalCheckDate)).Format("02-Jan-2006")
	endDate := time.Now().Format("02-Jan-2006")
	i.Log(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, rekening.PemilikRekening, startDate))
	logger.Debug(fmt.Sprintf("proses scrape mutasi rekening: %s, a/n: %s, dari tanggal: %s", nomorRekening, rekening.PemilikRekening, startDate))
	logger.Debug(endDate)

	cbTanggalAwal, err := page.WaitForSelector("#TransactionHistoryFG\\.SELECTED_RADIO_INDEX", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapat cbTanggalAwal ")
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = cbTanggalAwal.Check(); err != nil {
		logger.Debug("Gagal men-check cbTangalAwal :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	typeTanggal, err := page.WaitForSelector("#TransactionHistoryFG\\.FROM_TXN_DATE", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element kolom tanggal :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = typeTanggal.Type(startDate); err != nil {
		logger.Debug("Gagal mengetik tanggal :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	endDateField, err := page.WaitForSelector("#TransactionHistoryFG\\.TO_TXN_DATE", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	if err := endDateField.Type(endDate); err != nil {
		logger.Debug(err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	btnViewHistory, err := page.WaitForSelector("input#SEARCH", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element btn view history:", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err = btnViewHistory.Click(); err != nil {
		logger.Debug("Gagal klik btnViewHistory :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	page.WaitForTimeout(4000)
	tableHistory, err := page.QuerySelectorAll("table#txnHistoryList tr[id]")
	if err != nil {
		logger.Debug("Gagal mendapatkan tabel mutasi saldo :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	pages, err := page.WaitForSelector("span[class='paginationtxt1']", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("could not get pages :", err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	pagesString, err := pages.InnerText()
	if err != nil {
		logger.Debug("could not get pagesString :", err)
		return nil, page, err
	}

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	totalPages := strings.Split(pagesString, "dari")
	totalPages[1] = strings.TrimSpace(totalPages[1])
	totalPagesInt, err := strconv.Atoi(totalPages[1])
	if err != nil {
		logger.Debug("could not convert string to int :", err)
		return nil, page, err
	}
	logger.Debug("totalPagesInt :", totalPagesInt)
	page.WaitForTimeout(2000)

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	for k := 1; k <= totalPagesInt; k++ {
		logger.Debug(fmt.Sprintf("proses scraping mutasi, halaman: %d", k))
		i.Log(fmt.Sprintf("proses scraping mutasi, halaman: %d", k))

		tableHistory, err = page.QuerySelectorAll("table#txnHistoryList tr[id]")
		if err != nil {
			logger.Debug("Gagal mendapatkan tabel mutasi saldo :", err)
			return nil, page, err
		}

		for _, mutasiRow := range tableHistory {
			mutasi := &domain.Mutasi{}
			mutasi.TipeBank = domain.BankTypeBNI
			mutasi.Rekening = nomorRekening
			mutasi.PemilikRekening = name1 + name2

			cellValues, err := mutasiRow.QuerySelectorAll("td span")
			if err != nil {
				logger.Debug("Gagal mendapatkan cell values mutasi saldo :", err)
				return nil, page, err
			}

			for cellIndex, cell := range cellValues {
				cellValue, err := cell.InnerText()
				if err != nil {
					logger.Debug("Gagal mendapatkan cell value string mutasi saldo :", err)
					return nil, page, err
				}

				switch cellIndex {
				case 0:
					date, _ := time.Parse(constant.LayoutDateBNI, cellValue)
					pgDate := types.PGDate{Time: date}
					mutasi.TglBank = pgDate
				case 1:
					mutasi.Keterangan = cellValue
				case 4:
					if cellValue != "Cr." {
						mutasi.TipeMutasi = domain.MutasiRekeningTypeDebet
					} else {
						mutasi.TipeMutasi = domain.MutasiRekeningTypeKredit
					}
				case 5:
					if cellValue != "0.00" {
						jumlah := strings.Replace(cellValue, ",", "", -1)
						jumlahSep := strings.Split(jumlah, ".")
						jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "\u00a0", "")
						jumlahSep[0] = strings.ReplaceAll(jumlahSep[0], "&nbsp;", "")
						jumlahFloat, err := strconv.ParseInt(jumlahSep[0], 0, 64)
						if err != nil {
							logger.Debug("could not parse from string to int64 :", err)
						}
						mutasi.Jumlah = jumlahFloat
					}
				case 6:
					saldo := strings.Replace(cellValue, ",", "", -1)
					saldoSplit := strings.Split(saldo, ".")
					saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "\u00a0", "")
					saldoSplit[0] = strings.ReplaceAll(saldoSplit[0], "&nbsp;", "")
					saldoInt, err := strconv.ParseInt(saldoSplit[0], 0, 64)
					if err != nil {
						logger.Debug("could not parse to Int64 :", err)
					}
					mutasi.Saldo = saldoInt
				}
			}
			if mutasi.Saldo > 0 && mutasi.Jumlah > 0 {
				rekeningMutasiWraper.Mutasi = append(rekeningMutasiWraper.Mutasi, mutasi)
			}
		}

		if err = i.sessionTimeout(page); err != nil {
			return nil, page, err
		}

		btnNext, err := page.WaitForSelector("input[title='Selanjutnya']", playwright.PageWaitForSelectorOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			logger.Debug("Gagal mendapatkan element btn next:", err)
			return nil, page, err
		} else {
			page.WaitForTimeout(1000)
			if err := btnNext.Click(playwright.ElementHandleClickOptions{
				Timeout: playwright.Float(5000),
			}); err != nil {
				break
			}
			page.WaitForTimeout(2000)
		}
	}

	i.UpdateLoginStatus(domain.SudahLogin)
	i.Log("proses scraping selesai")
	logger.Debug("proses scraping selesai")

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	homeBtn, err := page.WaitForSelector("#BERANDA", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	page.WaitForTimeout(1000)
	if err := homeBtn.Click(); err != nil {
		logger.Debug(err)
		return nil, page, err
	}
	logger.Debug("back to home")

	if err = i.sessionTimeout(page); err != nil {
		return nil, page, err
	}

	return rekeningMutasiWraper, page, nil
}

func (i *Ibanking) logoutBNIPersonal(page playwright.Page) error {
	btnLogout1, err := page.WaitForSelector("#HREF_Logout", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element btn loginStatus 1 :", err)
		return err
	}
	logger.Debug("proses logout")
	i.Log("proses logout")

	page.WaitForTimeout(500)
	if err = btnLogout1.Click(); err != nil {
		logger.Debug("Gagal klik btnLogout1 :", err)
		return err
	}
	btnLogout2, err := page.WaitForSelector("#LOG_OUT", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		logger.Debug("Gagal mendapatkan element btn loginStatus 2:", err)
		return err
	}
	page.WaitForTimeout(500)
	if err = btnLogout2.Click(); err != nil {
		logger.Debug("Gagal klik btnloginStatus2 :", err)
		return err
	}

	if err := page.Close(); err != nil {
		logger.Debug(err)
	}
	i.Log("logout berhasil")
	logger.Debug("logout berhasil")

	return nil
}

func (i *Ibanking) IsBNIPersonalLogin(page playwright.Page) bool {
	if _, err := page.WaitForSelector("input#AuthenticationFG\\.USER_PRINCIPAL", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	}); err != nil {
		return true
	}
	return false
}

func (i *Ibanking) resultBNIPersonal(page playwright.Page) (playwright.Page, error) {
	if len(i.BankAccount.RekOnpay) > 0 {
		for _, rekOp := range i.BankAccount.RekOnpay {
			result, page1, err := i.scrapeBNIPersonal(
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
			result, page1, err := i.scrapeBNIPersonal(
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

func (i *Ibanking) backToHomeBNIPersonal(page playwright.Page) error {
	homeBtn, err := page.WaitForSelector("#BERANDA", playwright.PageWaitForSelectorOptions{
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

	return nil
}

func (i *Ibanking) sessionTimeout(page playwright.Page) error {
	sessOut, err := page.WaitForSelector("#page_content > div > h1", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000),
	})
	if err == nil {
		msg, err := sessOut.InnerText()
		if err != nil {
			logger.Debug(err)
		}
		return errors.New(msg)
	}
	return nil
}
