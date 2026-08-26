package export

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

func formatIDR(v float64) string {
	return fmt.Sprintf("Rp%.0f", v)
}

// GeneratePDF membuat ringkasan laporan (tren, breakdown kategori, breakdown anggota) sebagai
// PDF di memori — tidak pernah ditulis ke disk (Phase 12 §Item Penting: stream dari buffer).
func GeneratePDF(data ReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, "Laporan Keuangan Rumah Tangga")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 8, "Periode: "+data.PeriodLabel)
	pdf.Ln(12)

	// Tren bulanan
	pdf.SetFont("Helvetica", "B", 13)
	pdf.Cell(0, 8, "Tren Bulanan")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(50, 7, "Bulan", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, "Pemasukan", "1", 0, "R", false, 0, "")
	pdf.CellFormat(60, 7, "Pengeluaran", "1", 0, "R", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("Helvetica", "", 10)
	for _, t := range data.Trend {
		pdf.CellFormat(50, 7, t.Month, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 7, formatIDR(t.Income), "1", 0, "R", false, 0, "")
		pdf.CellFormat(60, 7, formatIDR(t.Expense), "1", 0, "R", false, 0, "")
		pdf.Ln(7)
	}
	pdf.Ln(8)

	// Breakdown kategori
	pdf.SetFont("Helvetica", "B", 13)
	pdf.Cell(0, 8, "Breakdown Pengeluaran per Kategori")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(80, 7, "Kategori", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 7, "%", "1", 0, "R", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("Helvetica", "", 10)
	for _, c := range data.CategoryBreakdown {
		pdf.CellFormat(80, 7, c.CategoryName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 7, formatIDR(c.Total), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%.1f%%", c.Percentage), "1", 0, "R", false, 0, "")
		pdf.Ln(7)
	}
	pdf.Ln(8)

	// Breakdown anggota
	pdf.SetFont("Helvetica", "B", 13)
	pdf.Cell(0, 8, "Kontribusi per Anggota")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(60, 7, "Nama", "1", 0, "L", false, 0, "")
	pdf.CellFormat(65, 7, "Pengeluaran", "1", 0, "R", false, 0, "")
	pdf.CellFormat(65, 7, "Pemasukan", "1", 0, "R", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("Helvetica", "", 10)
	for _, m := range data.MemberBreakdown {
		pdf.CellFormat(60, 7, m.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(65, 7, formatIDR(m.TotalExpense), "1", 0, "R", false, 0, "")
		pdf.CellFormat(65, 7, formatIDR(m.TotalIncome), "1", 0, "R", false, 0, "")
		pdf.Ln(7)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
