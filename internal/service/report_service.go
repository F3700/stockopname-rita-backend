package service

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"stockopname-rita-backend/internal/dto"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/xuri/excelize/v2"
)

//go:embed assets/logoritapasaraya_mini.png
var ritaLogo []byte

var reportLocation = mustLoadReportLocation()

type ReportService interface {
	SessionPDF(ctx context.Context, id int) ([]byte, string, error)
	CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error)
	SessionExcel(ctx context.Context, id int) ([]byte, string, error)
}

type ReportServiceImpl struct {
	SesiService        SesiService
	CoordinatorService CoordinatorService
	InspectorService   InspectorService
	StockOpnameService StockOpnameService
}

func NewReportService(sesi SesiService, coordinator CoordinatorService, inspector InspectorService, stock StockOpnameService) ReportService {
	return &ReportServiceImpl{SesiService: sesi, CoordinatorService: coordinator, InspectorService: inspector, StockOpnameService: stock}
}

func (r *ReportServiceImpl) SessionPDF(ctx context.Context, id int) ([]byte, string, error) {
	session, err := r.SesiService.FindById(ctx, id)
	if err != nil {
		return nil, "", err
	}
	coordinators, err := r.CoordinatorService.FindAllSummary(ctx, &id)
	if err != nil {
		return nil, "", err
	}
	results, err := r.StockOpnameService.FindAll(ctx, &dto.Pagination{Page: 1, Limit: 100000}, "", nil, &id)
	if err != nil {
		return nil, "", err
	}
	pdf := newPDF("STOCK OPNAME REPORT")
	pdf.Ln(4)
	section(pdf, "SESSION INFORMATION")
	label(pdf, "Session Code", session.Code)
	label(pdf, "Location", session.Location)
	label(pdf, "Status", session.Status)
	label(pdf, "Started At", formatReportDate(session.StartDate))
	label(pdf, "Ended At", formatReportDate(session.EndDate))
	section(pdf, "COORDINATORS")
	table(pdf, []string{"Code", "Status"}, func() [][]string {
		rows := make([][]string, 0, len(coordinators))
		for _, c := range coordinators {
			rows = append(rows, []string{c.Code, c.Status})
		}
		return rows
	}())
	section(pdf, "STOCK OPNAME RESULTS")
	resultTable(pdf, results)
	footer(pdf)
	data, err := pdfBytes(pdf)
	return data, fmt.Sprintf("%s.pdf", session.Code), err
}

func (r *ReportServiceImpl) CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error) {
	coordinator, err := r.CoordinatorService.FindByIdSummary(ctx, id)
	if err != nil {
		return nil, "", err
	}
	inspectors, err := r.InspectorService.FindAllSummary(ctx, &id)
	if err != nil {
		return nil, "", err
	}
	results, err := r.StockOpnameService.FindAll(ctx, &dto.Pagination{Page: 1, Limit: 100000}, "", &id, nil)
	if err != nil {
		return nil, "", err
	}
	report, err := r.CoordinatorService.FindByIdReport(ctx, id)
	if err != nil {
		return nil, "", err
	}
	pdf := newPDF("COORDINATOR STOCK OPNAME")
	pdf.Ln(4)
	section(pdf, "COORDINATOR INFORMATION")
	label(pdf, "Coordinator Code", coordinator.Code)
	label(pdf, "Session", report.SessionCode)
	label(pdf, "Status", coordinator.Status)
	section(pdf, "INSPECTORS")
	table(pdf, []string{"Code", "Rack Assigned", "Rack Completed", "Items"}, func() [][]string {
		rows := make([][]string, 0, len(inspectors))
		for _, i := range inspectors {
			rows = append(rows, []string{i.Code, fmt.Sprint(i.RackAssigned), fmt.Sprint(i.RackCompleted), fmt.Sprint(i.TotalItems)})
		}
		return rows
	}())
	section(pdf, "STOCK OPNAME RESULTS")
	resultTable(pdf, results)
	footer(pdf)
	data, err := pdfBytes(pdf)
	return data, fmt.Sprintf("%s.pdf", coordinator.Code), err
}

func (r *ReportServiceImpl) SessionExcel(ctx context.Context, id int) ([]byte, string, error) {
	session, err := r.SesiService.FindById(ctx, id)
	if err != nil {
		return nil, "", err
	}
	exports, err := r.StockOpnameService.FindAllForExport(ctx, id)
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	sheet := "Results"
	if _, err := file.NewSheet(sheet); err != nil {
		return nil, "", err
	}
	file.DeleteSheet("Sheet1")

	headers := []string{"Barcode", "Product", "Buy Price", "Sell Price", "Quantity", "Rak", "Inspector", "Coordinator", "Session Code"}
	if err := file.SetSheetRow(sheet, "A1", &headers); err != nil {
		return nil, "", err
	}
	for i, export := range exports {
		row := []interface{}{export.Barcode, export.Name, export.BuyPrice, export.SellPrice, export.Quantity, export.RackName, export.InspectorCode, export.CoordinatorCode, session.Code}
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return nil, "", err
		}
		if err := file.SetSheetRow(sheet, cell, &row); err != nil {
			return nil, "", err
		}
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	return buffer.Bytes(), fmt.Sprintf("%s.xlsx", session.Code), nil
}

func newPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.RegisterImageOptionsReader("rita-logo", gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, bytes.NewReader(ritaLogo))
	pdf.ImageOptions("rita-logo", 15, 12, 14, 0, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.SetY(12)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.SetY(29)
	return pdf
}
func section(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(235, 235, 235)
	pdf.CellFormat(0, 8, title, "B", 1, "L", true, 0, "")
	pdf.Ln(2)
}
func label(pdf *gofpdf.Fpdf, key, value string) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(38, 6, key, "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(0, 6, value, "", "L", false)
}
func table(pdf *gofpdf.Fpdf, headers []string, rows [][]string, widths ...float64) {
	cols := len(headers)
	colWidths := make([]float64, cols)
	if len(widths) == cols {
		copy(colWidths, widths)
	} else {
		even := 180.0 / float64(cols)
		for i := range colWidths {
			colWidths[i] = even
		}
	}
	drawHeader := func() {
		pdf.SetFont("Arial", "B", 8)
		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 7, header, "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}
	drawHeader()
	pdf.SetFont("Arial", "", 8)
	for _, row := range rows {
		lineCounts := make([]int, len(row))
		rowHeight := 6.0
		for index, value := range row {
			lineCounts[index] = len(pdf.SplitLines([]byte(value), colWidths[index]-2))
			if float64(lineCounts[index])*6 > rowHeight {
				rowHeight = float64(lineCounts[index]) * 6
			}
		}
		if pdf.GetY()+rowHeight > 282 {
			pdf.AddPage()
			drawHeader()
			pdf.SetFont("Arial", "", 8)
		}
		startX, startY := pdf.GetXY()
		for index, value := range row {
			x := startX
			for j := 0; j < index; j++ {
				x += colWidths[j]
			}
			pdf.SetXY(x, startY)
			pdf.MultiCell(colWidths[index], 6, value, "1", "L", false)
			pdf.SetXY(x+colWidths[index], startY)
		}
		pdf.SetXY(startX, startY+rowHeight)
	}
	pdf.Ln(2)
}
func resultTable(pdf *gofpdf.Fpdf, results []dto.StockOpnameResponse) {
	table(pdf, []string{"Barcode", "Product", "Rack", "Inspector", "Qty"}, func() [][]string {
		rows := make([][]string, 0, len(results))
		for _, result := range results {
			rows = append(rows, []string{result.Barcode, result.Name, result.RackName, result.InspectorCode, fmt.Sprint(result.Quantity)})
		}
		return rows
	}(), 30, 70, 30, 30, 20)
}
func footer(pdf *gofpdf.Fpdf) {
	pdf.Ln(4)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 6, "Generated At: "+formatReportDate(time.Now().UTC().Format(time.RFC3339)), "", 1, "L", false, 0, "")
}

func mustLoadReportLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}

func formatReportDate(value string) string {
	if value == "" {
		return "-"
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.Year() <= 1 {
		return "-"
	}
	return parsed.In(reportLocation).Format("02 January 2006, 15:04")
}
func pdfBytes(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
