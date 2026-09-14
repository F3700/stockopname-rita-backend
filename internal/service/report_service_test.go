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
		&fakeInspectorService{},
		&fakeStockOpnameService{
			findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
				if sesiId == nil || *sesiId != 1 {
					t.Errorf("expected sesiId 1, got %v", sesiId)
				}
				if coorId != nil {
					t.Errorf("expected nil coorId, got %d", *coorId)
				}
				return []dto.StockOpnameResponse{
					{Id: 1, Barcode: "8991234567890", Name: "Indomie", Quantity: 48, RackName: "R1", InspectorCode: "INSP-01"},
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
		&fakeInspectorService{},
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
		&fakeInspectorService{},
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
		&fakeInspectorService{
			findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
				if coorId == nil || *coorId != 2 {
					t.Errorf("expected coorId 2, got %v", coorId)
				}
				return []dto.InspectorResponse{{ID: 1, Code: "INSP-01", RackAssigned: 2, RackCompleted: 2, TotalItems: 10}}, nil
			},
		},
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
	inspectorCalled := false
	svc := NewReportService(
		&fakeSesiService{},
		&fakeCoordinatorService{
			findByIdSummaryFunc: func(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
				return nil, &model.NotFoundError{Resource: "Coordinator", ID: id}
			},
		},
		&fakeInspectorService{
			findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
				inspectorCalled = true
				return nil, nil
			},
		},
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
	if inspectorCalled {
		t.Error("inspector lookup must not run when coordinator lookup fails")
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
