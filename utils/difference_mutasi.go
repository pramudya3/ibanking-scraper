package utils

import (
	"fmt"
	"ibanking-scraper/domain"
)

func DifferenceMutasi(mutasi1, mutasi2 []*domain.Mutasi) []*domain.Mutasi {
	map1 := make(map[interface{}]struct{}, len(mutasi2))
	for _, mutasi := range mutasi2 {
		key := fmt.Sprint(",,,,,", mutasi.TglBank, mutasi.Rekening, mutasi.Keterangan, mutasi.Jumlah, mutasi.Saldo, mutasi.Griyabayar)
		map1[key] = struct{}{}
	}

	var diff []*domain.Mutasi
	for _, mutasi := range mutasi1 {
		key := fmt.Sprint(",,,,,", mutasi.TglBank, mutasi.Rekening, mutasi.Keterangan, mutasi.Jumlah, mutasi.Saldo, mutasi.Griyabayar)
		if _, found := map1[key]; !found {
			diff = append(diff, mutasi)
		}
	}

	return diff
}
