package saved_browser

import (
	"github.com/playwright-community/playwright-go"
	"sync"
)

type SavedBrowser struct {
	Browser sync.Map
	Page    map[uint64]*playwright.Page
}

var lock = &sync.Mutex{}
var sb *SavedBrowser

func GetSavedBrowser() *SavedBrowser {
	if sb == nil {
		lock.Lock()
		defer lock.Unlock()
		if sb == nil {
			sb = &SavedBrowser{
				Browser: sync.Map{},
				Page:    make(map[uint64]*playwright.Page),
			}
		}
	}

	return sb
}
