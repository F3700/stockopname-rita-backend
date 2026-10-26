package service

import (
	"bytes"
	"context"
	"errors"
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
			findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
				if sesiId == nil || *sesiId != 1 {
					t.Errorf("expected sesiId 1, got %v", sesiId)
				}
				if coorId != nil {
					t.Errorf("expected nil coorId, got %d", *coorId)
				}
				return []dto.StockOpnameResponse{
					{Id: 1, Barcode: "8991234567890", Name: "Indomie", Quantity: 48, RackName: "R1"},
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
			findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
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
			findByIdSummaryFunc: func(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
				if id != 2 {
					t.Errorf("expected id 2, got %d", id)
				}
				return &dto.CoordinatorResponse{ID: 2, Code: "KOR-02", Status: "COMPLETED"}, nil
			},
			findByIdReportFunc: func(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error) {
				if id != 2 {
					t.Errorf("expected id 2, got %d", id)
				}
				return &dto.CoordinatorReportResponse{
					Coordinator: dto.CoordinatorResponse{ID: 2, Code: "KOR-02", Status: "COMPLETED"},
					SessionCode: "SESI-01",
				}, nil
			},
		},
		&fakeUploadService{},
		&fakeStockOpnameService{
			findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
				if coorId == nil || *coorId != 2 {
					t.Errorf("expected coorId 2, got %v", coorId)
				}
				if sesiId != nil {
					t.Errorf("expected nil sesiId, got %d", *sesiId)
				}
				return []dto.StockOpnameResponse{}, nil
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
}

func TestCoordinatorPDFCoordinatorError(t *testing.T) {
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{
			findByIdSummaryFunc: func(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
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
							{Barcode: "111", Name: "Indomie", Qty: 5},
							{Barcode: "222", Name: "Soto", Qty: 3},
						}},
						{RakID: 2, RakName: "Rak 2", ItemCount: 1, Items: []dto.UploadItemResponse{
							{Barcode: "333", Name: "Teh", Qty: 1},
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
}
