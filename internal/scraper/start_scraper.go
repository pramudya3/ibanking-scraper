package scraper

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/utils"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (i *Ibanking) StartScrape() {
	i.UpdateLoginStatus(domain.ProsesLogin)

	var (
		browser playwright.Browser
		page    playwright.Page
	)

	startTime := i.BankAccount.JamAktifStart.Time
	endTime := i.BankAccount.JamAktifEnd.Time

	startAt := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
	endAt := time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
	interval := time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute
	sleepTime := time.Hour

	switch i.BankAccount.TipeAkun {
	case domain.AccountTypeBRICMS:
		if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
			logger.Debug("scraping masih berjalan")
		} else {
			browser = utils.InitiateBrowser(false)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)
			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start1:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartBRICMS()
						} else {
							logger.Debug("sudah login")
						scrape1:
							if len(browser.Contexts()) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
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
												goto start1
											} else {
												return
											}
										}
										goto scrape1
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
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(sleepTime)
					}
				} else {
					return
				}
			}
		}
	case domain.AccountTypeBNIDirect:
		if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
			logger.Debug("scraping masih berjalan.")
		} else {
			browser = utils.InitiateBrowser(true)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)
			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start2:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartBNIDirect()
						} else {
							logger.Debug("sudah login")
						scrape2:
							if len(browser.Contexts()) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
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
												goto start2
											} else {
												return
											}
										}
										goto scrape2
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
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(sleepTime)
					}
				} else {
					return
				}
			}
		}
	case domain.AccountTypeMandiriMCM:
		if value, _ := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil {
			logger.Debug("scraping masih berjalan.")
		} else {
			browser = utils.InitiateBrowser(false)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)

			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start3:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartMandiriMCM()
						} else {
							logger.Debug("sudah login")
						scrape3:
							if len(browser.Contexts()) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
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
												goto start3
											} else {
												return
											}
										}
										goto scrape3
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
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(sleepTime)
					}
				} else {
					return
				}
			}
		}
	case domain.AccountTypeBCABisnis:
		if value, _ := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil {
			logger.Debug("scraping masih berjalan.")
		} else {
			browser = utils.InitiateBrowser(true)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)

			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start4:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartBCABisnis()
						} else {
						scrape4:
							logger.Debug("sudah login")
							if len(browserCtx) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
								page, err := i.resultBCABisnis(page)
								if err != nil {
									time.Sleep(time.Second)
									if len(browser.Contexts()) != 0 {
										i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
										i.Log("Terjadi error, mencoba scraping ulang")
										if err := i.backToHomeBCABisnis(page); err != nil {
											i.UpdateLoginStatus(domain.BelumLogin)
											i.Log("tidak bisa kembali ke menu utama, logout...")
											if err := i.softLogout(page); err != nil {
												i.Log("proses logout gagal, menutup halaman")
												page.Close()
											}
											goto start4
										}
										goto scrape4
									} else {
										return
									}
								} else {
									i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
								}
							} else {
								return
							}
						}
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(interval)
					}
				} else {
					return
				}
			}
		}
	case domain.AccountTypeBCAPersonal:
		if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
			logger.Debug("scraping masih berjalan.")
		} else {
			browser = utils.InitiateBrowser(false)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)
			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start5:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartBCAPersonal()
						} else {
							logger.Debug("sudah login")
						scrape5:
							if len(browser.Contexts()) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
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
												goto start5
											} else {
												return
											}
										}
										goto scrape5
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
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(sleepTime)
					}
				} else {
					return
				}
			}
		}
	case domain.AccountTypeBNIPersonal:
		if value, _ := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil {
			logger.Debug("scraping masih berjalan")
		} else {
			browser = utils.InitiateBrowser(true)
			i.SavedBrowser.Browser.Store(i.BankAccount.ID, browser)
			for {
				startTime = i.BankAccount.JamAktifStart.Time
				endTime = i.BankAccount.JamAktifEnd.Time
				startAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.Now().Location())
				endAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.Now().Location())
				interval = time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute

				if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
					if time.Now().After(startAt) && time.Now().Before(endAt) {
						browserCtx := value.(playwright.Browser).Contexts()
					start6:
						if len(browserCtx) == 0 {
							logger.Debug("belum login")
							i.StartBNIPersonal()
						} else {
							logger.Debug("sudah login")
						scrape6:
							if len(browser.Contexts()) != 0 {
								page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
								page, err := i.resultBNIPersonal(page)
								if err != nil {
									time.Sleep(time.Second)
									if len(browser.Contexts()) != 0 {
										i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
										if strings.Contains(err.Error(), "session") {
											i.Log(err.Error())
											logger.Debug(err.Error())
											page.Close()
											time.Sleep(time.Duration(i.BankAccount.IntervalCek.Int64) * time.Minute)
											goto start6
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
													goto start6
												} else {
													return
												}
											}
											goto scrape6
										}
									} else {
										return
									}
								} else {
									if i.BankAccount.AutoLogout.Bool {
										i.UpdateLoginStatus(domain.BelumLogin)
										i.logoutBNIPersonal(page)
										i.Log("Browser masih terbuka")
									}
									i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = &page
								}
							} else {
								return
							}
						}
						logger.Debug("next: ", time.Now().Add(interval))
						time.Sleep(interval)
					} else {
						time.Sleep(sleepTime)
					}
				} else {
					return
				}
			}
		}
	}
}
