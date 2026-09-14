package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/xuri/excelize/v2"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func excelTestReportService(t *testing.T, sesiErr, exportErr error) ReportService {
	t.Helper()
	return NewReportService(
		&fakeSesiService{
			findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
				if sesiErr != nil {
					return dto.SesiResponse{}, sesiErr
				}
				return dto.SesiResponse{ID: 1, Code: "SESI-01", Location: "Gudang A", Status: "COMPLETED"}, nil
			},
		},
		&fakeCoordinatorService{},
		&fakeInspectorService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error) {
				if exportErr != nil {
					return nil, exportErr
				}
				if sesiId != 1 {
					t.Errorf("expected sesiId 1, got %d", sesiId)
				}
				return []dto.StockOpnameExportResponse{
					{Id: 1, Barcode: "8991234567890", Name: "Indomie", BuyPrice: 2500, SellPrice: 3500, Quantity: 48, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01"},
					{Id: 2, Barcode: "8991234567891", Name: "Soto", BuyPrice: 2500, SellPrice: 3500, Quantity: 36, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01"},
				}, nil
			},
		},
	)
}

func TestSessionExcel(t *testing.T) {
	svc := excelTestReportService(t, nil, nil)

	data, filename, err := svc.SessionExcel(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01.xlsx" {
		t.Errorf("expected filename SESI-01.xlsx, got %q", filename)
	}
	if !bytes.HasPrefix(data, []byte("PK")) {
		t.Fatal("expected xlsx zip magic header")
	}

	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("generated file is not a valid xlsx: %v", err)
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows("Results")
	if err != nil {
		t.Fatalf("expected sheet Results: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected header + 2 data rows, got %d rows", len(rows))
	}

	wantHeader := []string{"Barcode", "Product", "Buy Price", "Sell Price", "Quantity", "Rak", "Inspector", "Coordinator", "Session Code"}
	if len(rows[0]) != len(wantHeader) {
		t.Fatalf("expected %d header columns, got %d", len(wantHeader), len(rows[0]))
	}
	for i, want := range wantHeader {
		if rows[0][i] != want {
			t.Errorf("expected header[%d] %q, got %q", i, want, rows[0][i])
		}
	}

	first := rows[1]
	if len(first) != len(wantHeader) {
		t.Fatalf("expected %d columns in data row, got %d", len(wantHeader), len(first))
	}
	wantFirst := []string{"8991234567890", "Indomie", "2500", "3500", "48", "R1", "INSP-01", "KOR-01", "SESI-01"}
	for i, want := range wantFirst {
		if first[i] != want {
			t.Errorf("expected cell[%d] %q, got %q", i, want, first[i])
		}
	}
}

func TestSessionExcelSesiError(t *testing.T) {
	svc := excelTestReportService(t, &model.NotFoundError{Resource: "Session", ID: 99}, nil)

	_, _, err := svc.SessionExcel(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestSessionExcelExportError(t *testing.T) {
	svc := excelTestReportService(t, nil, errors.New("boom"))

	_, _, err := svc.SessionExcel(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
