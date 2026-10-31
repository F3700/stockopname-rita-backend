package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"stockopname-rita-backend/internal/dto"
	"strconv"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/valentin-kaiser/go-dbase/dbase"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
)

var reportLocation = mustLoadReportLocation()

type ReportService interface {
	SessionPDF(ctx context.Context, id int) ([]byte, string, error)
	CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error)
	CoordinatorUploadSummaryPDF(ctx context.Context, id int) ([]byte, string, error)
	UploadPDF(ctx context.Context, id int) ([]byte, string, error)
	SessionExcel(ctx context.Context, id int) ([]byte, string, error)
	SessionDBF(ctx context.Context, id int) ([]byte, string, error)
}

type ReportServiceImpl struct {
	SesiService        SesiService
	CoordinatorService CoordinatorService
	UploadService      UploadService
	StockOpnameService StockOpnameService
}

func NewReportService(sesi SesiService, coordinator CoordinatorService, upload UploadService, stock StockOpnameService) ReportService {
	return &ReportServiceImpl{SesiService: sesi, CoordinatorService: coordinator, UploadService: upload, StockOpnameService: stock}
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
	exports, err := r.StockOpnameService.FindAllForExport(ctx, &id, nil)
	if err != nil {
		return nil, "", err
	}
	groups := groupByRack(exports)
	pdf := newPDF("Session Report")
	reportPair(pdf, "Session", session.Code)
	reportPair(pdf, "Location", session.Location)
	reportPair(pdf, "Status", session.Status)
	pdf.Ln(4.6)
	reportPair(pdf, "Started At", formatReportDate(session.StartDate))
	reportPair(pdf, "Ended At", formatReportDate(session.EndDate))
	reportPair(pdf, "Coordinators", fmt.Sprint(len(coordinators)))
	pdf.Ln(5)
	section(pdf, "COORDINATORS")
	table(pdf, []string{"Code", "Status"}, func() [][]string {
		rows := make([][]string, 0, len(coordinators))
		for _, c := range coordinators {
			rows = append(rows, []string{c.Code, c.Status})
		}
		return rows
	}())
	section(pdf, "REKAP RAK")
	rackRecap(pdf, groups, true)
	if len(groups) > 0 {
		section(pdf, "DETAIL PER RAK")
		rackDetails(pdf, groups, true)
		reportGrandTotal(pdf, groups, "TOTAL SESI")
	}
	footer(pdf)
	data, err := pdfBytes(pdf)
	return data, fmt.Sprintf("%s.pdf", session.Code), err
}

func (r *ReportServiceImpl) CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error) {
	coordinator, err := r.CoordinatorService.FindByIdDetail(ctx, id)
	if err != nil {
		return nil, "", err
	}
	exports, err := r.StockOpnameService.FindAllForExport(ctx, nil, &id)
	if err != nil {
		return nil, "", err
	}
	groups := groupByRack(exports)
	pdf := newPDF("Coordinator Report")
	reportPair(pdf, "Coordinator", coordinator.Code)
	reportPair(pdf, "Session", coordinator.SessionCode)
	reportPair(pdf, "Location", coordinator.SessionLocation)
	pdf.Ln(4.6)
	reportPair(pdf, "Status", coordinator.Status)
	reportPair(pdf, "Racks", fmt.Sprint(coordinator.RackAssigned))
	reportPair(pdf, "Items", fmt.Sprint(len(exports)))
	pdf.Ln(5)
	section(pdf, "REKAP RAK")
	rackRecap(pdf, groups, false)
	if len(groups) > 0 {
		section(pdf, "DETAIL PER RAK")
		rackDetails(pdf, groups, false)
		reportGrandTotal(pdf, groups, "TOTAL KOORDINATOR")
	}
	footer(pdf)
	data, err := pdfBytes(pdf)
	return data, fmt.Sprintf("%s.pdf", coordinator.Code), err
}

// CoordinatorUploadSummaryPDF renders one summary row per upload for a coordinator.
func (r *ReportServiceImpl) CoordinatorUploadSummaryPDF(ctx context.Context, id int) ([]byte, string, error) {
	coordinator, err := r.CoordinatorService.FindByIdDetail(ctx, id)
	if err != nil {
		return nil, "", err
	}
	rows, err := r.UploadService.FindSummaryByCoordinator(ctx, id)
	if err != nil {
		return nil, "", err
	}

	pdf := newUploadPDF("Upload Summary Report")

	pair := func(label, value string) {
		reportPair(pdf, label, value)
	}
	pair("Koordinator", coordinator.Code)
	pair("Session", coordinator.SessionCode)
	pair("Lokasi", coordinator.SessionLocation)
	pdf.Ln(5)

	tableRows := make([][]string, 0, len(rows))
	var totalQty, totalBuy, totalSell float64
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			row.CoordinatorCode,
			strconv.Itoa(row.UploadNumber),
			formatQty(row.Qty),
			formatMoneyID(row.TotalBuy),
			formatMoneyID(row.TotalSell),
		})
		totalQty += row.Qty
		totalBuy += row.TotalBuy
		totalSell += row.TotalSell
	}
	total := []string{
		"TOTAL",
		"",
		formatQty(totalQty),
		formatMoneyID(totalBuy),
		formatMoneyID(totalSell),
	}
	uploadTable(pdf, []string{"Kode Koor", "Upload No", "Qty", "Total HB", "Total HJ"}, tableRows, total, []float64{30, 25, 25, 53, 53})
	footer(pdf)

	data, err := pdfBytes(pdf)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("%s_%s_upload-summary.pdf", sanitizeReportName(coordinator.SessionCode), sanitizeReportName(coordinator.Code))
	return data, filename, nil
}

// UploadPDF renders the immutable snapshot taken when the batch was uploaded.
func (r *ReportServiceImpl) UploadPDF(ctx context.Context, id int) ([]byte, string, error) {
	detail, err := r.UploadService.FindDetail(ctx, id)
	if err != nil {
		return nil, "", err
	}

	pdf := newUploadPDF("Upload Report")
	uploadInfo(pdf, detail)
	for _, rack := range detail.Racks {
		uploadRackSection(pdf, rack)
	}
	footer(pdf)

	data, err := pdfBytes(pdf)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("%s_%s_upload-%d.pdf", sanitizeReportName(detail.SessionCode), sanitizeReportName(detail.CoordinatorCode), detail.UploadID)
	return data, filename, nil
}

func (r *ReportServiceImpl) SessionExcel(ctx context.Context, id int) ([]byte, string, error) {
	session, err := r.SesiService.FindById(ctx, id)
	if err != nil {
		return nil, "", err
	}
	exports, err := r.StockOpnameService.FindAllForExport(ctx, &id, nil)
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	sheet := "Results"
	if _, err := file.NewSheet(sheet); err != nil {
		return nil, "", err
	}
	file.DeleteSheet("Sheet1")

	headers := []string{"Barcode", "Product", "Buy Price", "Sell Price", "Quantity", "Rak", "Coordinator", "Session Code", "Updated At"}
	if err := file.SetSheetRow(sheet, "A1", &headers); err != nil {
		return nil, "", err
	}
	for i, export := range exports {
		row := []interface{}{export.Barcode, export.Name, export.BuyPrice, export.SellPrice, export.Quantity, export.RackName, export.CoordinatorCode, session.Code, export.UpdatedAt}
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return nil, "", err
		}
		if err := file.SetSheetRow(sheet, cell, &row); err != nil {
			return nil, "", err
		}
	}
	if len(exports) > 0 {
		qtyStyle, err := file.NewStyle(&excelize.Style{NumFmt: 2})
		if err != nil {
			return nil, "", err
		}
		if err := file.SetCellStyle(sheet, "E2", fmt.Sprintf("E%d", len(exports)+1), qtyStyle); err != nil {
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
	Quantity  float64 `dbase:"QTY"`
	Rak       string  `dbase:"RAK"`
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
		{"QTY", dbase.Numeric, 10, 2},
		{"RAK", dbase.Character, 15, 0},
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
	exports, err := r.StockOpnameService.FindAllForExport(ctx, &id, nil)
	if err != nil {
		return nil, "", err
	}

	columns, err := sessionDBFColumns()
	if err != nil {
		return nil, "", err
	}

	// go-dbase's UnixIO/Create uppercases the whole filesystem path
	// (e.g. /tmp/x.dbf -> /TMP/X.DBF), which breaks on case-sensitive
	// Linux filesystems (works on Windows because it is case-insensitive).
	// Use an in-memory buffer via GenericIO so no temp file/dir is needed.
	// This also keeps distroless containers working (no /tmp required).
	// NOTE: dbase.NewBytesReadWriteSeeker can't be used here — it rejects
	// Seek/Write beyond the current length, while DBF creation seeks to
	// offset 32+ on an empty file (like a real sparse file would allow).
	mem := newMemReadWriteSeeker()
	memIO := dbase.GenericIO{Handle: mem}

	// FoxBasePlus (0x03) emits a dBASE III file: the widest-supported
	// variant for legacy readers and the surrounding retail toolchain.
	// All field types used here (Character, Numeric) are dBASE III native.
	table, err := dbase.NewTable(
		dbase.FoxBasePlus,
		&dbase.Config{
			Filename:  "stockopname.dbf",
			Converter: dbase.NewDefaultConverter(charmap.Windows1252),
		},
		columns,
		0,
		memIO,
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
	raw := mem.Data()
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

// memReadWriteSeeker is a file-like in-memory buffer with sparse-write
// semantics (Seek beyond EOF + Write zero-extends, like os.File).
// dbase.NewBytesReadWriteSeeker lacks this, so DBF creation fails with
// "invalid seek position" when the library seeks to offset 32+ on day-zero.
type memReadWriteSeeker struct {
	data []byte
	pos  int64
}

func newMemReadWriteSeeker() *memReadWriteSeeker {
	return &memReadWriteSeeker{data: make([]byte, 0)}
}

func (m *memReadWriteSeeker) Read(p []byte) (int, error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.pos:])
	m.pos += int64(n)
	return n, nil
}

func (m *memReadWriteSeeker) Write(p []byte) (int, error) {
	if m.pos < 0 {
		return 0, fmt.Errorf("negative seek position: %d", m.pos)
	}
	endPos := m.pos + int64(len(p))
	if endPos > int64(len(m.data)) {
		// Zero-extend like a real file (handles Seek past EOF).
		ext := make([]byte, endPos-int64(len(m.data)))
		m.data = append(m.data, ext...)
	}
	copy(m.data[m.pos:], p)
	m.pos = endPos
	return len(p), nil
}

func (m *memReadWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = m.pos + offset
	case io.SeekEnd:
		newPos = int64(len(m.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}
	if newPos < 0 {
		return 0, fmt.Errorf("negative seek position: %d", newPos)
	}
	m.pos = newPos
	return newPos, nil
}

func (m *memReadWriteSeeker) Close() error { return nil }

// Data returns a copy of the buffered bytes.
func (m *memReadWriteSeeker) Data() []byte {
	out := make([]byte, len(m.data))
	copy(out, m.data)
	return out
}

func newPDF(subtitle string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Pasaraya Ritelindo - "+subtitle, false)
	pdf.SetMargins(12, 10, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()
	reportHeader(pdf, subtitle)
	return pdf
}

// newUploadPDF keeps the upload report compact while still showing the brand.
func newUploadPDF(subtitle string) *gofpdf.Fpdf {
	return newPDF(subtitle)
}

func reportHeader(pdf *gofpdf.Fpdf, subtitle string) {
	pdf.SetY(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 7, "Pasaraya Ritelindo", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, subtitle, "", 1, "C", false, 0, "")
	pdf.Ln(3)
}

func section(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, title, "B", 1, "L", false, 0, "")
	pdf.Ln(1.5)
}

func reportPair(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(20, 4.6, label, "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(42, 4.6, value, "", 0, "L", false, 0, "")
}
func table(pdf *gofpdf.Fpdf, headers []string, rows [][]string, widths ...float64) {
	cols := len(headers)
	colWidths := make([]float64, cols)
	if len(widths) == cols {
		copy(colWidths, widths)
	} else {
		even := 186.0 / float64(cols)
		for i := range colWidths {
			colWidths[i] = even
		}
	}
	drawHeader := func() {
		pdf.SetFont("Arial", "B", 7)
		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 5, header, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
	drawHeader()
	pdf.SetFont("Arial", "", 7.5)
	for _, row := range rows {
		rowHeight := 4.6
		for index, value := range row {
			lines := len(pdf.SplitLines([]byte(value), colWidths[index]))
			if float64(lines)*4.6 > rowHeight {
				rowHeight = float64(lines) * 4.6
			}
		}
		if pdf.GetY()+rowHeight > 282 {
			pdf.AddPage()
			drawHeader()
			pdf.SetFont("Arial", "", 7.5)
		}
		startX, startY := pdf.GetXY()
		for index, value := range row {
			x := startX
			for j := 0; j < index; j++ {
				x += colWidths[j]
			}
			pdf.SetXY(x, startY)
			pdf.MultiCell(colWidths[index], 4.6, value, "1", "L", false)
			pdf.SetXY(x+colWidths[index], startY)
		}
		pdf.SetXY(startX, startY+rowHeight)
	}
	pdf.Ln(2)
}

func uploadInfo(pdf *gofpdf.Fpdf, detail dto.UploadDetailResponse) {
	reportPair(pdf, "Coordinator", detail.CoordinatorCode)
	reportPair(pdf, "Session", detail.SessionCode)
	reportPair(pdf, "Location", detail.SessionLocation)
	pdf.Ln(4.6)
	reportPair(pdf, "Uploaded At", formatReportDate(detail.CreatedAt))
	reportPair(pdf, "Racks", fmt.Sprint(detail.RackCount))
	reportPair(pdf, "Items", fmt.Sprint(detail.ItemCount))
	pdf.Ln(5)
}

func uploadRackSection(pdf *gofpdf.Fpdf, rack dto.UploadDetailRackResponse) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(0, 5, rack.RakName, "B", 1, "L", false, 0, "")
	pdf.Ln(1)

	rows := make([][]string, 0, len(rack.Items))
	var totalQty, totalBuy, totalSell float64
	for _, item := range rack.Items {
		rows = append(rows, []string{
			item.Barcode,
			item.PLU,
			item.Name,
			formatQty(item.Qty),
			formatMoneyID(item.BuyPrice),
			formatMoneyID(item.SellPrice),
		})
		totalQty += item.Qty
		totalBuy += item.Qty * item.BuyPrice
		totalSell += item.Qty * item.SellPrice
	}
	total := []string{
		"TOTAL",
		"",
		fmt.Sprintf("%d item", len(rack.Items)),
		formatQty(totalQty),
		formatMoneyID(totalBuy),
		formatMoneyID(totalSell),
	}
	uploadTable(pdf, []string{"Barcode", "PLU", "Produk", "Qty", "Harga Beli", "Harga Jual"}, rows, total, []float64{24, 18, 72, 14, 29, 29})
	pdf.Ln(1)
}

func uploadTable(pdf *gofpdf.Fpdf, headers []string, rows [][]string, total []string, widths []float64) {
	const lineHeight = 4.6
	drawHeader := func() {
		pdf.SetFont("Arial", "B", 7)
		for i, header := range headers {
			pdf.CellFormat(widths[i], 5, header, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
	drawRow := func(row []string, bold bool) {
		if bold {
			pdf.SetFont("Arial", "B", 7.5)
		} else {
			pdf.SetFont("Arial", "", 7.5)
		}
		startX, startY := pdf.GetXY()
		rowHeight := lineHeight
		for index, value := range row {
			lines := len(pdf.SplitLines([]byte(value), widths[index]))
			if height := float64(lines) * lineHeight; height > rowHeight {
				rowHeight = height
			}
		}
		if pdf.GetY()+rowHeight > 282 {
			pdf.AddPage()
			drawHeader()
			if bold {
				pdf.SetFont("Arial", "B", 7.5)
			} else {
				pdf.SetFont("Arial", "", 7.5)
			}
			startX, startY = pdf.GetXY()
		}
		for index, value := range row {
			x := startX
			for j := 0; j < index; j++ {
				x += widths[j]
			}
			pdf.SetXY(x, startY)
			pdf.MultiCell(widths[index], lineHeight, value, "1", "L", false)
			pdf.SetXY(x+widths[index], startY)
		}
		pdf.SetXY(startX, startY+rowHeight)
	}

	drawHeader()
	for _, row := range rows {
		drawRow(row, false)
	}
	drawRow(total, true)
	pdf.Ln(2)
}

type rackReportGroup struct {
	Coordinator string
	Rack        string
	Items       []dto.StockOpnameExportResponse
	ItemCount   int
	TotalQty    float64
	TotalBuy    float64
	TotalSell   float64
}

func groupByRack(rows []dto.StockOpnameExportResponse) []rackReportGroup {
	groups := make([]rackReportGroup, 0)
	index := make(map[string]int)
	for _, row := range rows {
		key := row.CoordinatorCode + "\x00" + row.RackName
		position, ok := index[key]
		if !ok {
			position = len(groups)
			index[key] = position
			groups = append(groups, rackReportGroup{Coordinator: row.CoordinatorCode, Rack: row.RackName})
		}
		group := &groups[position]
		group.Items = append(group.Items, row)
		group.ItemCount++
		group.TotalQty += row.Quantity
		group.TotalBuy += row.Quantity * row.BuyPrice
		group.TotalSell += row.Quantity * row.SellPrice
	}
	return groups
}

func rackRecap(pdf *gofpdf.Fpdf, groups []rackReportGroup, withCoordinator bool) {
	if len(groups) == 0 {
		pdf.SetFont("Arial", "I", 9)
		pdf.CellFormat(0, 6, "Belum ada hasil stock opname.", "", 1, "L", false, 0, "")
		return
	}

	headers := []string{"No", "Coordinator", "Rack", "Item", "Qty", "Total Beli", "Total Jual"}
	widths := []float64{10, 30, 30, 12, 18, 43, 43}
	if !withCoordinator {
		headers = []string{"No", "Rack", "Item", "Qty", "Total Beli", "Total Jual"}
		widths = []float64{10, 48, 15, 18, 47, 48}
	}

	rows := make([][]string, 0, len(groups))
	for i, group := range groups {
		row := []string{strconv.Itoa(i + 1), group.Rack, strconv.Itoa(group.ItemCount), formatQty(group.TotalQty), formatMoneyID(group.TotalBuy), formatMoneyID(group.TotalSell)}
		if withCoordinator {
			row = []string{strconv.Itoa(i + 1), group.Coordinator, group.Rack, strconv.Itoa(group.ItemCount), formatQty(group.TotalQty), formatMoneyID(group.TotalBuy), formatMoneyID(group.TotalSell)}
		}
		rows = append(rows, row)
	}
	table(pdf, headers, rows, widths...)
}

func rackDetails(pdf *gofpdf.Fpdf, groups []rackReportGroup, withCoordinator bool) {
	for _, group := range groups {
		title := group.Rack
		if withCoordinator {
			title += " - " + group.Coordinator
		}
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(0, 5, title, "B", 1, "L", false, 0, "")
		pdf.Ln(1)

		rows := make([][]string, 0, len(group.Items))
		for _, item := range group.Items {
			rows = append(rows, []string{
				item.Barcode,
				item.PLU,
				item.Name,
				formatQty(item.Quantity),
				formatMoneyID(item.BuyPrice),
				formatMoneyID(item.SellPrice),
			})
		}
		total := []string{
			"TOTAL",
			"",
			fmt.Sprintf("%d item", group.ItemCount),
			formatQty(group.TotalQty),
			formatMoneyID(group.TotalBuy),
			formatMoneyID(group.TotalSell),
		}
		uploadTable(pdf, []string{"Barcode", "PLU", "Produk", "Qty", "Harga Beli", "Harga Jual"}, rows, total, []float64{24, 18, 72, 14, 29, 29})
		pdf.Ln(1)
	}
}

func reportGrandTotal(pdf *gofpdf.Fpdf, groups []rackReportGroup, title string) {
	if len(groups) == 0 {
		return
	}
	var itemCount int
	var totalQty, totalBuy, totalSell float64
	for _, group := range groups {
		itemCount += group.ItemCount
		totalQty += group.TotalQty
		totalBuy += group.TotalBuy
		totalSell += group.TotalSell
	}
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(0, 8, fmt.Sprintf("%s: %d item, %s qty | Beli %s | Jual %s", title, itemCount, formatQty(totalQty), formatMoneyID(totalBuy), formatMoneyID(totalSell)), "T", 1, "L", false, 0, "")
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

func sanitizeReportName(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "report"
	}
	return b.String()
}

func formatQty(q float64) string {
	return strings.Replace(strconv.FormatFloat(q, 'f', 2, 64), ".", ",", 1)
}

func formatMoneyID(v float64) string {
	formatted := strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
	parts := strings.SplitN(formatted, ".", 2)
	whole := parts[0]
	var grouped strings.Builder
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(r)
	}
	return fmt.Sprintf("Rp %s,%s", grouped.String(), parts[1])
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
