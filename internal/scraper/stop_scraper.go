package scraper

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/logger"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (i *Ibanking) StopScrape() {
	i.UpdateLoginStatus(domain.TidakAktif)
	var page playwright.Page
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); ok {
		if value != nil {
			browserCtx := value.(playwright.Browser).Contexts()
			if len(browserCtx) != 0 {
				if i.SavedBrowser.Page[uint64(i.BankAccount.ID)] != nil {
					page = *i.SavedBrowser.Page[uint64(i.BankAccount.ID)]
					logger.Debug("oke")
				} else {
					logger.Debug("tidak oke")
					page = browserCtx[0].Pages()[0]
				}

				i.Log("Proses menutup browser")
				logger.Debug("Proses menutup browser")

				switch i.BankAccount.TipeAkun {
				case domain.AccountTypeBNIDirect:
					i.LogoutBNIDirect(page)
				case domain.AccountTypeBRICMS:
					if i.IsBRICMSLogin(page) {
						i.LogoutBRICMS(page)
					}
				case domain.AccountTypeMandiriMCM:
					if i.isMandiriMCMLogin(page) {
						i.LogoutMandiriMCM(page)
					}
				case domain.AccountTypeBCABisnis:
					if i.isLogin1BCABisnis(page) {
						if i.isLogin2BCABisnis(page) {
							i.LogoutBCABisnis(page)
						}
					}
				case domain.AccountTypeBCAPersonal:
					if i.isLoginBCAPersonal(page) {
						i.logoutBCAPersonal(page)
					}
				case domain.AccountTypeBNIPersonal:
					if i.IsBNIPersonalLogin(page) {
						i.logoutBNIPersonal(page)
					}
				}
			}

			i.CloseBrowser()
			i.SavedBrowser.Page[uint64(i.BankAccount.ID)] = nil
			i.SavedBrowser.Browser.Delete(i.BankAccount.ID)
			time.Sleep(2 * time.Second)
			logger.Debug("menutup browser")
			i.Log("menutup browser")
			i.UpdateLoginStatus(domain.TidakAktif)
		}
	}
}
