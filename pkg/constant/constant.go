package constant

import "ibanking-scraper/internal/types"

const (
	ApiKey                  = "1f4282c35b46fd1451a1e40e71ba0ee5"
	ContextKeyUser          = types.ContextKey("user")
	LayoutDateISO8601       = "2006-01-02"
	LayoutDateISO8602       = "02-01-2006"
	LayoutDateTimeISO8601   = "2006-01-02T15:04:05.000Z"
	LayoutTime              = "15:04:05"
	LayoutTimeWithFloat     = "15:04:05.999999999"
	LayoutTimeWithoutSecond = "15:04"
	LayoutDate              = "02/01/06"
	LayoutDateTimeBNI       = "02-Jan-2006 15:04:05"
	LayoutDateBNI           = "02-Jan-2006"
	LayoutDateMandiri       = "02/01/2006"
	Key                     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	LayoutMandiriKopra      = "Monday January 2 2006"
)
