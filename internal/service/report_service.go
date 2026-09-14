package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"stockopname-rita-backend/internal/dto"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/valentin-kaiser/go-dbase/dbase"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
)

//go:embed assets/logoritapasaraya_mini.png
var ritaLogo []byte

var reportLocation = mustLoadReportLocation()

type ReportService interface {
	SessionPDF(ctx context.Context, id int) ([]byte, string, error)
	CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error)
	SessionExcel(ctx context.Context, id int) ([]byte, string, error)
	SessionDBF(ctx context.Context, id int) ([]byte, string, error)
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

	headers := []string{"Barcode", "Product", "Buy Price", "Sell Price", "Quantity", "Rak", "Inspector", "Coordinator", "Session Code", "Updated At"}
	if err := file.SetSheetRow(sheet, "A1", &headers); err != nil {
		return nil, "", err
	}
	for i, export := range exports {
		row := []interface{}{export.Barcode, export.Name, export.BuyPrice, export.SellPrice, export.Quantity, export.RackName, export.InspectorCode, export.CoordinatorCode, session.Code, export.UpdatedAt}
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

// dbfExportRow is the row layout for the session DBF export. Field names
// follow dBase limits (max 10 characters, uppercased by the library).
type dbfExportRow struct {
	Barcode   string  `dbase:"BARCODE"`
	Product   string  `dbase:"PRODUCT"`
	BuyPrice  float64 `dbase:"BUYPRICE"`
	SellPrice float64 `dbase:"SELLPRICE"`
	Quantity  int     `dbase:"QTY"`
	Rak       string  `dbase:"RAK"`
	Inspector string  `dbase:"INSPECTOR"`
	CoorCode  string  `dbase:"COOR"`
	SesiCode  string  `dbase:"SESI_CODE"`
	Updated   string  `dbase:"UPDATED"`
}

func sessionDBFColumns() ([]*dbase.Column, error) {
	specs := []struct {
		name     string
		dataType dbase.DataType
		length   uint8
		decimals uint8
	}{
		{"BARCODE", dbase.Character, 20, 0},
		{"PRODUCT", dbase.Character, 150, 0},
		{"BUYPRICE", dbase.Numeric, 15, 2},
		{"SELLPRICE", dbase.Numeric, 15, 2},
		{"QTY", dbase.Numeric, 10, 0},
		{"RAK", dbase.Character, 15, 0},
		{"INSPECTOR", dbase.Character, 20, 0},
		{"COOR", dbase.Character, 20, 0},
		{"SESI_CODE", dbase.Character, 20, 0},
		{"UPDATED", dbase.Character, 25, 0},
	}
	columns := make([]*dbase.Column, 0, len(specs))
	for _, spec := range specs {
		column, err := dbase.NewColumn(spec.name, spec.dataType, spec.length, spec.decimals, false)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func (r *ReportServiceImpl) SessionDBF(ctx context.Context, id int) ([]byte, string, error) {
	session, err := r.SesiService.FindById(ctx, id)
	if err != nil {
		return nil, "", err
	}
	exports, err := r.StockOpnameService.FindAllForExport(ctx, id)
	if err != nil {
		return nil, "", err
	}

	columns, err := sessionDBFColumns()
	if err != nil {
		return nil, "", err
	}

	// go-dbase only supports file-backed creation, so stage through a
	// unique temp file which is removed before returning.
	tmp, err := os.CreateTemp("", "stockopname-*.dbf")
	if err != nil {
		return nil, "", err
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(tmpName)
	defer func() { _ = os.Remove(tmpName) }()

	// FoxBasePlus (0x03) emits a dBASE III file: the widest-supported
	// variant for legacy readers and the surrounding retail toolchain.
	// All field types used here (Character, Numeric) are dBASE III native.
	table, err := dbase.NewTable(
		dbase.FoxBasePlus,
		&dbase.Config{
			Filename:  tmpName,
			Converter: dbase.NewDefaultConverter(charmap.Windows1252),
		},
		columns,
		0,
		nil,
	)
	if err != nil {
		return nil, "", err
	}

	for _, export := range exports {
		row, err := table.RowFromStruct(&dbfExportRow{
			Barcode:   export.Barcode,
			Product:   export.Name,
			BuyPrice:  export.BuyPrice,
			SellPrice: export.SellPrice,
			Quantity:  export.Quantity,
			Rak:       export.RackName,
			Inspector: export.InspectorCode,
			CoorCode:  export.CoordinatorCode,
			SesiCode:  session.Code,
			Updated:   export.UpdatedAt,
		})
		if err != nil {
			_ = table.Close()
			return nil, "", err
		}
		if err := row.Add(); err != nil {
			_ = table.Close()
			return nil, "", err
		}
	}

	if err := table.Close(); err != nil {
		return nil, "", err
	}
	raw, err := os.ReadFile(tmpName)
	if err != nil {
		return nil, "", err
	}
	data, err := compactDBFFile(raw, len(columns))
	if err != nil {
		return nil, "", err
	}
	return data, fmt.Sprintf("%s.dbf", session.Code), nil
}

// compactDBFFile rewrites a go-dbase generated file into strict dBase III
// layout: the library computes a wrong first-row offset (296 + 32*fields,
// leaving a zero gap after the field descriptors) and omits the 0x1A
// end-of-file marker, which makes strict DBF readers reject the file.
// This sets the header offset to 32 + 32*fields + 1, drops the gap and
// appends the EOF marker. Row bytes themselves are untouched.
func compactDBFFile(raw []byte, fieldCount int) ([]byte, error) {
	if len(raw) < 32 {
		return nil, fmt.Errorf("dbf file too short: %d bytes", len(raw))
	}
	headerLen := 32 + 32*fieldCount + 1
	if len(raw) < headerLen {
		return nil, fmt.Errorf("dbf file too short for %d fields: %d bytes", fieldCount, len(raw))
	}
	if raw[headerLen-1] != 0x0D {
		return nil, fmt.Errorf("dbf header terminator missing at offset %d", headerLen-1)
	}
	rowLen := int(binary.LittleEndian.Uint16(raw[10:12]))
	recordCount := int(binary.LittleEndian.Uint32(raw[4:8]))
	claimedFirst := int(binary.LittleEndian.Uint16(raw[8:10]))
	end := claimedFirst + recordCount*rowLen
	if claimedFirst < headerLen || end > len(raw) {
		return nil, fmt.Errorf("dbf record section out of bounds: first=%d end=%d size=%d", claimedFirst, end, len(raw))
	}

	fixed := make([]byte, 0, headerLen+recordCount*rowLen+1)
	fixed = append(fixed, raw[:headerLen]...)
	binary.LittleEndian.PutUint16(fixed[8:10], uint16(headerLen))
	fixed = append(fixed, raw[claimedFirst:end]...)
	fixed = append(fixed, 0x1A)
	return fixed, nil
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
