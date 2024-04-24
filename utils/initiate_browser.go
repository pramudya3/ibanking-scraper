package utils

import (
	"ibanking-scraper/internal/logger"

	"github.com/playwright-community/playwright-go"
)

func InitiateBrowser(isChromium bool) playwright.Browser {
	err := playwright.Install()
	if err != nil {
		logger.Debug("could not install playwright: ", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		logger.Debug("could not start playwright: ", err)
	}

	var browser playwright.Browser
	if isChromium {
		browser, _ = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			//Headless: playwright.Bool(false),
		})
	} else {
		browser, _ = pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
			//Headless: playwright.Bool(false),
		})
	}

	return browser
}
