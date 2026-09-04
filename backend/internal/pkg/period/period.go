// Package period menghitung window transaksi & label periode "YYYY-MM" untuk household yang
// punya siklus budget custom (budget_cycle_start_day) — misal gajian tanggal 25, bukan tanggal 1.
//
// Package leaf (tidak import apapun dari repository/usecase) supaya bisa dipakai baik di layer
// usecase maupun repository/postgres tanpa import cycle.
package period

import "time"

// MaxCycleStartDay dibatasi 28 (bukan 31) supaya aritmetik AddDate(-1 bulan)/(+1 bulan) pada
// tanggal cutoff tidak pernah overflow bulan pendek (Februari) — menghindari kebutuhan logic
// clamp seperti DueDateForPeriod, cukup dengan membatasi domain input.
const MaxCycleStartDay = 28

// CyclePeriodWindow mengembalikan [start, end) transaksi yang termasuk period "YYYY-MM" untuk
// household dengan cycle_start_day tertentu.
//
// startDay<=1: identik perilaku kalender lama, TIDAK ADA shift sama sekali — households yang
// tidak pernah ubah setting ini (default) tidak akan lihat perubahan apapun.
// startDay>1: period M mencakup [hari startDay bulan M-1, hari startDay bulan M) — transaksi
// SEBELUM tanggal cutoff dianggap milik bulan berikutnya (bulan pemakaian), bukan bulan diterima.
func CyclePeriodWindow(period string, startDay int) (start, end time.Time, err error) {
	base, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if startDay <= 1 {
		return base, base.AddDate(0, 1, 0), nil
	}
	end = time.Date(base.Year(), base.Month(), startDay, 0, 0, 0, 0, time.UTC)
	start = end.AddDate(0, -1, 0)
	return start, end, nil
}

// CurrentCyclePeriod mengembalikan label "YYYY-MM" yang memuat `now`, konsisten dengan
// CyclePeriodWindow. Next-month dihitung dari first-of-month, BUKAN now.AddDate(0,1,0)
// langsung — now bisa hari 29/30/31 dan overflow ke bulan yang salah (bug klasik Go time.Time).
func CurrentCyclePeriod(now time.Time, startDay int) string {
	if startDay <= 1 || now.Day() < startDay {
		return now.Format("2006-01")
	}
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return firstOfMonth.AddDate(0, 1, 0).Format("2006-01")
}

// PreviousCyclePeriod mengembalikan label bulan sebelumnya — aman terhadap overflow karena
// selalu beroperasi di tanggal 1 (period selalu diparse ke first-of-month).
func PreviousCyclePeriod(period string) (string, error) {
	base, err := time.Parse("2006-01", period)
	if err != nil {
		return "", err
	}
	return base.AddDate(0, -1, 0).Format("2006-01"), nil
}
