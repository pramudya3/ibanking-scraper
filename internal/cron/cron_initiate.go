package cron_initiate

import "github.com/makelifemorefun/cron"

func NewCron() (*cron.Cron, error) {
	return cron.New(), nil
}
