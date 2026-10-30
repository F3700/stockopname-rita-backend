package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/valentin-kaiser/go-dbase/dbase"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func dbfTestReportService(t *testing.T, sesiErr, exportErr error) ReportService {
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
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
				if exportErr != nil {
					return nil, exportErr
				}
				if sesiId == nil || *sesiId != 1 {
					t.Errorf("expected sesiId 1, got %v", sesiId)
				}
				if coorId != nil {
					t.Errorf("expected nil coorId, got %d", *coorId)
				}
				return []dto.StockOpnameExportResponse{
					{Id: 1, Barcode: "8991234567890", Name: "Indomie", BuyPrice: 2500, SellPrice: 3500, Quantity: 48.5, RackName: "R1", CoordinatorCode: "KOR-01", UpdatedAt: "2026-09-10T08:00:00Z"},
					{Id: 2, Barcode: "8991234567891", Name: "Soto", BuyPrice: 2500, SellPrice: 3500, Quantity: 36, RackName: "R1", CoordinatorCode: "KOR-01", UpdatedAt: "2026-09-10T08:00:00Z"},
				}, nil
			},
		},
	)
}

func openTestDBF(t *testing.T, data []byte) *dbase.File {
	t.Helper()
	// Untested bypasses go-dbase's own version gate: it only supports
	// reading back FoxPro variants, while we intentionally emit dBASE III.
	table, err := dbase.OpenTable(&dbase.Config{Data: data, Untested: true})
	if err != nil {
		t.Fatalf("generated file is not a valid DBF: %v", err)
	}
	t.Cleanup(func() { _ = table.Close() })
	return table
}

func TestSessionDBF(t *testing.T) {
	svc := dbfTestReportService(t, nil, nil)

	data, filename, err := svc.SessionDBF(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01.dbf" {
		t.Errorf("expected filename SESI-01.dbf, got %q", filename)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty DBF")
	}
	if data[0] != 0x03 {
		t.Errorf("expected dBASE III version byte 0x03, got 0x%02x", data[0])
	}
	if count := binary.LittleEndian.Uint32(data[4:8]); count != 2 {
		t.Errorf("expected 2 records in header, got %d", count)
	}
	// Strict dBase III layout: header(32) + 9 descriptors(288) + terminator(1).
	if first := binary.LittleEndian.Uint16(data[8:10]); first != 321 {
		t.Errorf("expected first-row offset 321, got %d", first)
	}
	if data[320] != 0x0D {
		t.Errorf("expected header terminator 0x0D at offset 320, got 0x%02x", data[320])
	}
	if want := 321 + 2*291 + 1; len(data) != want {
		t.Errorf("expected exact file size %d, got %d", want, len(data))
	}
	if data[len(data)-1] != 0x1A {
		t.Errorf("expected EOF marker 0x1A, got 0x%02x", data[len(data)-1])
	}

	table := openTestDBF(t, data)
	if got := table.Header().RecordsCount(); got != 2 {
		t.Errorf("expected 2 records, got %d", got)
	}

	row, err := table.Row()
	if err != nil {
		t.Fatalf("failed to read first row: %v", err)
	}
	// DBF character fields are space-padded to their full width.
	if got := strings.TrimSpace(fmt.Sprintf("%v", row.FieldByName("BARCODE").GetValue())); got != "8991234567890" {
		t.Errorf("expected barcode %q, got %q", "8991234567890", got)
	}
	if got := strings.TrimSpace(fmt.Sprintf("%v", row.FieldByName("QTY").GetValue())); got != "48.5" {
		t.Errorf("expected qty 48.5, got %q", got)
	}
	if !bytes.Contains(data, []byte("48.50")) {
		t.Error("expected DBF record to store QTY with 2 decimals")
	}
	if got := strings.TrimSpace(fmt.Sprintf("%v", row.FieldByName("BUYPRICE").GetValue())); got != "2500" && got != "2500.00" {
		t.Errorf("expected buy price 2500, got %q", got)
	}
	if got := strings.TrimSpace(fmt.Sprintf("%v", row.FieldByName("SESI_CODE").GetValue())); got != "SESI-01" {
		t.Errorf("expected sesi code SESI-01, got %q", got)
	}
	if got := strings.TrimSpace(fmt.Sprintf("%v", row.FieldByName("UPDATED").GetValue())); got != "2026-09-10T08:00:00Z" {
		t.Errorf("expected updated timestamp, got %q", got)
	}
}

func TestSessionDBFEmpty(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{
			findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
				return dto.SesiResponse{ID: 1, Code: "SESI-01"}, nil
			},
		},
		&fakeCoordinatorService{},
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
				return nil, nil
			},
		},
	)

	data, filename, err := svc.SessionDBF(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "SESI-01.dbf" {
		t.Errorf("expected filename SESI-01.dbf, got %q", filename)
	}

	table := openTestDBF(t, data)
	if got := table.Header().RecordsCount(); got != 0 {
		t.Errorf("expected 0 records, got %d", got)
	}
}

func TestSessionDBFSesiError(t *testing.T) {
	svc := dbfTestReportService(t, &model.NotFoundError{Resource: "Session", ID: 99}, nil)

	_, _, err := svc.SessionDBF(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestSessionDBFExportError(t *testing.T) {
	svc := dbfTestReportService(t, nil, errors.New("boom"))

	_, _, err := svc.SessionDBF(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSessionDBFColumns(t *testing.T) {
	columns, err := sessionDBFColumns()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(columns) != 9 {
		t.Fatalf("expected 9 columns, got %d", len(columns))
	}
	for _, column := range columns {
		if len(column.Name()) > 10 {
			t.Errorf("column name %q exceeds 10 characters", column.Name())
		}
	}
}
