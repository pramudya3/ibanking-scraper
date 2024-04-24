package scraper

import (
	"context"
	"fmt"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/saved_browser"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/makelifemorefun/cron"
	"github.com/playwright-community/playwright-go"
)

type (
	DaftarRekening struct {
		PemilikRekening string
		Rekening        string
	}

	Ibanking struct {
		C             *cron.Cron
		SavedBrowser  *saved_browser.SavedBrowser
		BankAccount   *domain.BankAccount
		Logs          *domain.Log
		UcLogs        domain.LogUsecase
		UcRekening    domain.RekeningUsecase
		UcMutasi      domain.MutasiUsecase
		UcBankAccount domain.BankAccountUsecase
		C2            *gocron.Scheduler
	}
)

type MyScheduler struct {
	StartAt time.Time
	Every   time.Duration
	Delay   time.Duration
	StopAt  time.Time
}

func (s *MyScheduler) Next(t time.Time) time.Time {
	if t.After(s.StartAt) && t.Add(s.Every).Before(s.StopAt) {
		logger.Debug("Next :", t.Add(s.Every))
		return t.Add(s.Every)
	} else if t.Add(s.Every).After(s.StopAt) {
		s.StartAt = s.StartAt.AddDate(0, 0, 1)
		logger.Debug("Next :", s.StartAt)
	}

	return s.StartAt
}

func (i *Ibanking) Log(msg string) {
	log := &domain.Log{
		AkunBankId: i.BankAccount.ID,
		Tipe:       domain.TypeLog,
		Keterangan: msg,
	}

	if err := i.UcLogs.Create(context.Background(), log); err != nil {
		logger.Debug(err)
	}
}

func (i *Ibanking) ErrorLog(msg string) {
	log := &domain.Log{
		AkunBankId: i.BankAccount.ID,
		Tipe:       domain.TypeError,
		Keterangan: msg,
	}
	if err := i.UcLogs.Create(context.Background(), log); err != nil {
		logger.Debug(err)
	}
}

func (i *Ibanking) UpdateLoginStatus(loginStatus domain.LoginStatus) {
	statusLogin := &domain.BankAccount{
		StatusLogin: loginStatus,
	}

	if err := i.UcBankAccount.UpdateLoginStatus(context.Background(), statusLogin, uint64(i.BankAccount.ID)); err != nil {
		logger.Debug(err)
	}
}

func (i *Ibanking) CloseBrowser() {
	if value, ok := i.SavedBrowser.Browser.Load(i.BankAccount.ID); value != nil && ok {
		browser := value.(playwright.Browser)
		if len(browser.Contexts()) > 0 {
			for _, bws := range browser.Contexts() {
				for _, pgs := range bws.Pages() {
					if err := pgs.Close(); err != nil {
						logger.Debug(err)
					}
					logger.Debug("page closed")
				}
				if err := bws.Close(); err != nil {
					logger.Debug(err)
				}
				logger.Debug("browser context closed")
			}
		}
		if err := browser.Close(); err != nil {
			logger.Debug(err)
		}
		logger.Debug("browser closed")
	}
}

func (i *Ibanking) RemoveDuplicateMutasi(datas []*domain.Mutasi) []*domain.Mutasi {
	map1 := make(map[interface{}]struct{}, len(datas))
	var finalResult []*domain.Mutasi
	for _, data := range datas {
		key := fmt.Sprint(",,,", data.TglBank, data.Keterangan, data.Jumlah, data.Saldo)
		if _, ok := map1[key]; !ok {
			map1[key] = struct{}{}
			finalResult = append(finalResult, data)
		}
	}

	return finalResult
}
