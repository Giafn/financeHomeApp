package export

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// GenerateExcel membuat workbook 3 sheet (Tren, Kategori, Anggota) di memori — tidak pernah
// ditulis ke disk (Phase 12 §Item Penting: stream dari buffer).
func GenerateExcel(data ReportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet Tren
	trendSheet := "Tren Bulanan"
	f.NewSheet(trendSheet)
	f.SetCellValue(trendSheet, "A1", "Bulan")
	f.SetCellValue(trendSheet, "B1", "Pemasukan")
	f.SetCellValue(trendSheet, "C1", "Pengeluaran")
	for i, t := range data.Trend {
		row := i + 2
		f.SetCellValue(trendSheet, fmt.Sprintf("A%d", row), t.Month)
		f.SetCellValue(trendSheet, fmt.Sprintf("B%d", row), t.Income)
		f.SetCellValue(trendSheet, fmt.Sprintf("C%d", row), t.Expense)
	}

	// Sheet Kategori
	catSheet := "Breakdown Kategori"
	f.NewSheet(catSheet)
	f.SetCellValue(catSheet, "A1", "Kategori")
	f.SetCellValue(catSheet, "B1", "Total")
	f.SetCellValue(catSheet, "C1", "Persentase")
	for i, c := range data.CategoryBreakdown {
		row := i + 2
		f.SetCellValue(catSheet, fmt.Sprintf("A%d", row), c.CategoryName)
		f.SetCellValue(catSheet, fmt.Sprintf("B%d", row), c.Total)
		f.SetCellValue(catSheet, fmt.Sprintf("C%d", row), c.Percentage)
	}

	// Sheet Anggota
	memberSheet := "Kontribusi Anggota"
	f.NewSheet(memberSheet)
	f.SetCellValue(memberSheet, "A1", "Nama")
	f.SetCellValue(memberSheet, "B1", "Total Pengeluaran")
	f.SetCellValue(memberSheet, "C1", "Total Pemasukan")
	for i, m := range data.MemberBreakdown {
		row := i + 2
		f.SetCellValue(memberSheet, fmt.Sprintf("A%d", row), m.Name)
		f.SetCellValue(memberSheet, fmt.Sprintf("B%d", row), m.TotalExpense)
		f.SetCellValue(memberSheet, fmt.Sprintf("C%d", row), m.TotalIncome)
	}

	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
