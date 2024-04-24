package cron_initiate

import "go.uber.org/fx"

var Module = fx.Provide(
	NewCron,
)
