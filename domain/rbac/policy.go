package rbac

import "strings"

type Policy struct {
	Action     Action `db:"action" json:"action"`
	Menu       string `db:"menu" json:"menu"`
	Asset      string `db:"asset" json:"asset"`
	Griyabayar bool   `db:"griyabayar" json:"griyabayar"`
	Onpay      bool   `db:"onpay" json:"onpay"`
}

func (p Policy) IsGranted(act, asset string, isGriyabayar, isOnpay bool) bool {
	if p.Asset == "/" {
		granted := p.Action.FromHTTPMethod(act) && strings.HasPrefix(asset, p.Asset)
		if p.Griyabayar && isGriyabayar {
			return granted
		} else if p.Onpay && isOnpay {
			return granted
		} else {
			return false
		}
	}
	return p.Action.FromHTTPMethod(act) && strings.HasPrefix(asset, p.Asset) && p.Griyabayar == isGriyabayar && p.Onpay == isOnpay
}
