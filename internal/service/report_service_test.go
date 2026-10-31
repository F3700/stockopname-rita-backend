package service

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestSessionPDF(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{
			findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
				if id != 1 {
					t.Errorf("expected id 1, got %d", id)
				}
				return dto.SesiResponse{ID: 1, Code: "SESI-01", Location: "Gudang A", Status: "COMPLETED", StartDate: "2026-09-10T08:00:00+07:00", EndDate: "2026-09-10T12:00:00+07:00"}, nil
			},
		},
		&fakeCoordinatorService{
			findAllSummaryFunc: func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
				if sesiId == nil || *sesiId != 1 {
					t.Errorf("expected sesiId 1, got %v", sesiId)
				}
				return []*dto.CoordinatorResponse{{ID: 1, Code: "KOR-01", Status: "COMPLETED"}}, nil
			},
		},
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
				if sesiId == nil || *sesiId != 1 {
					t.Errorf("expected sesiId 1, got %v", sesiId)
				}
				if coorId != nil {
					t.Errorf("expected nil coorId, got %d", *coorId)
				}
				return []dto.StockOpnameExportResponse{
					{Id: 1, Barcode: "8991234567890", PLU: "100251", Name: "Indomie", Quantity: 48, BuyPrice: 2500, SellPrice: 3500, RackName: "R1", CoordinatorCode: "KOR-01"},
				}, nil
			},
		},
	)

	data, filename, err := svc.SessionPDF(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01.pdf" {
		t.Errorf("expected filename SESI-01.pdf, got %q", filename)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("expected PDF magic header")
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
	text := compressedPDFText(t, data)
	if !strings.Contains(text, "Pasaraya Ritelindo") || !strings.Contains(text, "Session Report") {
		t.Error("expected session report header")
	}
	for _, want := range []string{"PLU", "Harga Beli", "Harga Jual", "TOTAL", "Rp 2.500,00"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected session PDF to contain %q", want)
		}
	}
	if bytes.Contains(data, []byte("/Subtype /Image")) {
		t.Error("session PDF must not contain a logo image")
	}
}

func TestSessionPDFSesiError(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{
			findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
				return dto.SesiResponse{}, &model.NotFoundError{Resource: "Session", ID: id}
			},
		},
		&fakeCoordinatorService{},
		&fakeUploadService{},
		&fakeStockOpnameService{},
	)

	_, _, err := svc.SessionPDF(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestSessionPDFCoordinatorError(t *testing.T) {
	stockCalled := false
	svc := NewReportService(
		&fakeSesiService{
			findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
				return dto.SesiResponse{ID: 1, Code: "SESI-01"}, nil
			},
		},
		&fakeCoordinatorService{
			findAllSummaryFunc: func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
				return nil, errors.New("boom")
			},
		},
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
				stockCalled = true
				return nil, nil
			},
		},
	)

	_, _, err := svc.SessionPDF(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if stockCalled {
		t.Error("stock opname must not be fetched when coordinator lookup fails")
	}
}

func TestCoordinatorPDF(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{
			findByIdDetailFunc: func(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error) {
				if id != 2 {
					t.Errorf("expected id 2, got %d", id)
				}
				return &dto.CoordinatorDetailResponse{
					ID: 2, Code: "KOR-02", Status: "COMPLETED",
					SessionCode: "SESI-01", SessionLocation: "Gudang A",
					RackAssigned: 2, RackCompleted: 2,
				}, nil
			},
		},
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
				if coorId == nil || *coorId != 2 {
					t.Errorf("expected coorId 2, got %v", coorId)
				}
				if sesiId != nil {
					t.Errorf("expected nil sesiId, got %d", *sesiId)
				}
				return []dto.StockOpnameExportResponse{}, nil
			},
		},
	)

	data, filename, err := svc.CoordinatorPDF(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "KOR-02.pdf" {
		t.Errorf("expected filename KOR-02.pdf, got %q", filename)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("expected PDF magic header")
	}
	text := compressedPDFText(t, data)
	if !strings.Contains(text, "Pasaraya Ritelindo") || !strings.Contains(text, "Coordinator Report") {
		t.Error("expected coordinator report header")
	}
	if bytes.Contains(data, []byte("/Subtype /Image")) {
		t.Error("coordinator PDF must not contain a logo image")
	}
}

func TestCoordinatorPDFCoordinatorError(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{
			findByIdDetailFunc: func(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error) {
				return nil, &model.NotFoundError{Resource: "Coordinator", ID: id}
			},
		},
		&fakeUploadService{},
		&fakeStockOpnameService{},
	)

	_, _, err := svc.CoordinatorPDF(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestGroupByRack(t *testing.T) {
	rows := []dto.StockOpnameExportResponse{
		{CoordinatorCode: "KOR-01", RackName: "Rak 1", Quantity: 2.5, BuyPrice: 1000, SellPrice: 1500},
		{CoordinatorCode: "KOR-01", RackName: "Rak 1", Quantity: 1, BuyPrice: 1000, SellPrice: 1500},
		{CoordinatorCode: "KOR-01", RackName: "Rak 2", Quantity: 4, BuyPrice: 500, SellPrice: 750},
	}
	groups := groupByRack(rows)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Rack != "Rak 1" || groups[0].ItemCount != 2 || groups[0].TotalQty != 3.5 {
		t.Errorf("unexpected first group %+v", groups[0])
	}
	if groups[0].TotalBuy != 3500 || groups[0].TotalSell != 5250 {
		t.Errorf("unexpected first group totals %+v", groups[0])
	}
	if groups[1].TotalBuy != 2000 || groups[1].TotalSell != 3000 {
		t.Errorf("unexpected second group totals %+v", groups[1])
	}
}

func TestFormatMoneyID(t *testing.T) {
	cases := map[float64]string{
		0:           "Rp 0,00",
		1500:        "Rp 1.500,00",
		1234567.89:  "Rp 1.234.567,89",
		48.5 * 2500: "Rp 121.250,00",
	}
	for value, want := range cases {
		if got := formatMoneyID(value); got != want {
			t.Errorf("formatMoneyID(%v) = %q, want %q", value, got, want)
		}
	}
}

func TestFormatReportDate(t *testing.T) {
	if got := formatReportDate(""); got != "-" {
		t.Errorf("expected dash for empty, got %q", got)
	}
	if got := formatReportDate("not-a-date"); got != "-" {
		t.Errorf("expected dash for invalid, got %q", got)
	}
	if got := formatReportDate("0001-01-01T00:00:00Z"); got != "-" {
		t.Errorf("expected dash for zero time, got %q", got)
	}
	got := formatReportDate("2026-09-10T08:00:00+07:00")
	if got == "-" || got == "" {
		t.Errorf("expected formatted date, got %q", got)
	}
}

func compressedPDFText(t *testing.T, data []byte) string {
	t.Helper()
	rest := data
	var text strings.Builder
	for {
		start := bytes.Index(rest, []byte("stream"))
		if start < 0 {
			break
		}
		rest = rest[start+len("stream"):]
		rest = bytes.TrimLeft(rest, "\r\n")
		end := bytes.Index(rest, []byte("endstream"))
		if end < 0 {
			break
		}
		chunk := rest[:end]
		rest = rest[end+len("endstream"):]

		reader, err := zlib.NewReader(bytes.NewReader(chunk))
		if err != nil {
			continue
		}
		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, reader); err != nil {
			_ = reader.Close()
			continue
		}
		_ = reader.Close()
		text.Write(buffer.Bytes())
		text.WriteByte('\n')
	}
	return text.String()
}

func TestCoordinatorUploadSummaryPDF(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{
			findByIdDetailFunc: func(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error) {
				if id != 7 {
					t.Errorf("expected coordinator 7, got %d", id)
				}
				return &dto.CoordinatorDetailResponse{
					ID: 7, Code: "QA-C1", SessionCode: "SESI-01", SessionLocation: "Gudang A",
				}, nil
			},
		},
		&fakeUploadService{
			findSummaryFunc: func(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error) {
				return []*model.UploadReportRow{
					{CoordinatorCode: "QA-C1", UploadID: 1, UploadNumber: 1, Qty: 2.5, TotalBuy: 2500, TotalSell: 3750},
					{CoordinatorCode: "QA-C1", UploadID: 2, UploadNumber: 2, Qty: 1, TotalBuy: 2000, TotalSell: 2500},
				}, nil
			},
		},
		&fakeStockOpnameService{},
	)

	data, filename, err := svc.CoordinatorUploadSummaryPDF(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01_QA-C1_upload-summary.pdf" {
		t.Errorf("unexpected filename %q", filename)
	}
	text := compressedPDFText(t, data)
	for _, want := range []string{"Pasaraya Ritelindo", "Upload Summary Report", "Upload No", "Total HB", "Total HJ", "QA-C1", "3,50", "Rp 4.500,00", "Rp 6.250,00", "TOTAL"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected summary PDF to contain %q", want)
		}
	}
}

func TestUploadPDF(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{},
		&fakeUploadService{
			findDetailFunc: func(ctx context.Context, uploadID int) (dto.UploadDetailResponse, error) {
				if uploadID != 12 {
					t.Errorf("expected uploadID 12, got %d", uploadID)
				}
				return dto.UploadDetailResponse{
					UploadResponse:  dto.UploadResponse{UploadID: 12, RackCount: 2, ItemCount: 3, CreatedAt: "2026-10-26T10:14:00+07:00"},
					CoordinatorCode: "KOR-01",
					SessionCode:     "SESI-01",
					SessionLocation: "Gudang A",
					Racks: []dto.UploadDetailRackResponse{
						{RakID: 1, RakName: "Rak 1", ItemCount: 2, Items: []dto.UploadItemResponse{
							{PLU: "1001", Barcode: "111", Name: "Indomie", Qty: 5, BuyPrice: 2500, SellPrice: 3500},
							{PLU: "1002", Barcode: "222", Name: "Soto", Qty: 3, BuyPrice: 2000, SellPrice: 3000},
						}},
						{RakID: 2, RakName: "Rak 2", ItemCount: 1, Items: []dto.UploadItemResponse{
							{PLU: "1003", Barcode: "333", Name: "Teh", Qty: 1, BuyPrice: 1000, SellPrice: 1500},
						}},
					},
				}, nil
			},
		},
		&fakeStockOpnameService{},
	)

	data, filename, err := svc.UploadPDF(context.Background(), 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01_KOR-01_upload-12.pdf" {
		t.Errorf("unexpected filename %q", filename)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("expected PDF magic header")
	}
	text := compressedPDFText(t, data)
	for _, want := range []string{"Pasaraya Ritelindo", "Upload Report", "Harga Beli", "Harga Jual", "TOTAL", "1001", "Rp 2.500,00", "2 item"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected upload PDF to contain %q", want)
		}
	}
	if strings.Contains(text, "STOCK OPNAME UPLOAD") {
		t.Error("upload PDF must not print the old title")
	}
}
