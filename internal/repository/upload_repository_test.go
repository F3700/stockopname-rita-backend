package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestUploadFindAllSummary(t *testing.T) {
	mock := mustMockPool(t)
	createdAt := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("rack_count > 0").WithArgs(5).WillReturnRows(
		pgxmock.NewRows([]string{"UploadID", "UploadNumber", "CoorID", "RackCount", "ItemCount", "CreatedAt"}).
			AddRow(7, 2, 5, 1, 2, createdAt).
			AddRow(6, 1, 5, 2, 3, createdAt.Add(-time.Hour)),
	)

	repo := NewUploadRepository(mock)
	results, err := repo.FindAllSummary(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].UploadNumber != 2 || results[1].UploadID != 6 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestUploadFindSummaryByCoordinator(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("current_rows").WithArgs(7).WillReturnRows(
		pgxmock.NewRows([]string{"CoordinatorCode", "UploadID", "UploadNumber", "Qty", "TotalBuy", "TotalSell"}).
			AddRow("QA-C1", 1, 1, 2.5, 2500.0, 3750.0).
			AddRow("QA-C1", 2, 2, 1.0, 2000.0, 2500.0).
			AddRow("QA-C1", 0, 0, 3.0, 3000.0, 4500.0),
	)

	repo := NewUploadRepository(mock)
	rows, err := repo.FindSummaryByCoordinator(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 || rows[1].UploadNumber != 2 || rows[0].TotalSell != 3750 {
		t.Errorf("unexpected rows %+v", rows)
	}
	if rows[2].UploadID != 0 || rows[2].UploadNumber != 0 {
		t.Errorf("expected unuploaded bucket, got %+v", rows[2])
	}
}

func TestUploadFindDetailUsesCurrentResults(t *testing.T) {
	mock := mustMockPool(t)
	createdAt := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM upload u").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"upload_id", "upload_coor_id", "upload_createdat", "coor_code", "sesi_code", "sesi_location"}).
			AddRow(12, 7, createdAt, "QA-C1", "SESI-01", "Gudang A"),
	)
	mock.ExpectQuery("LEFT JOIN stock_opname").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id", "rak_name", "barcode", "plu", "name", "dept", "qty", "buy_price", "sell_price", "per_carton", "has_item"}).
			AddRow(1, "Rak 1", "899", "1001", "Produk A", "1138", 10.0, 2500.0, 3500.0, 12, true).
			AddRow(1, "Rak 1", "898", "1002", "Produk B", "1138", -2.0, 2000.0, 3000.0, 6, true).
			AddRow(2, "Rak 2", "", "", "", "", 0.0, 0.0, 0.0, 0, false),
	)

	repo := NewUploadRepository(mock)
	detail, err := repo.FindDetail(context.Background(), 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.RackCount != 2 || detail.ItemCount != 2 {
		t.Fatalf("unexpected counts: %+v", detail)
	}
	if len(detail.Racks) != 2 || len(detail.Racks[0].Items) != 2 || len(detail.Racks[1].Items) != 0 {
		t.Fatalf("unexpected racks: %+v", detail.Racks)
	}
	if detail.Racks[0].Items[0].Qty != 10 || detail.Racks[0].Items[1].Qty != -2 {
		t.Errorf("expected effective quantities, got %+v", detail.Racks[0].Items)
	}
	if detail.Racks[0].Items[0].PerCarton != 12 || detail.Racks[0].Items[1].PerCarton != 6 {
		t.Errorf("expected per carton values, got %+v", detail.Racks[0].Items)
	}
	if detail.Racks[1].ItemCount != 0 {
		t.Errorf("expected empty rack, got %+v", detail.Racks[1])
	}
}

func TestUploadFindAllSummaryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM upload").WithArgs(5).WillReturnError(errors.New("boom"))

	repo := NewUploadRepository(mock)
	_, err := repo.FindAllSummary(context.Background(), 5)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
