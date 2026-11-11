package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestAuditFindAll(t *testing.T) {
	mock := mustMockPool(t)
	createdAt := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM stock_opname_audit").WithArgs((*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil), false, 10, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"audit_id", "batch_id", "stock_opname_id", "sesi_id",
			"source_rak_id", "target_rak_id", "source_coor_id", "target_coor_id",
			"action", "before_data", "after_data", "actor", "note", "created_at",
		}).AddRow(int64(1), "", 7, 3, nil, nil, nil, nil, "MOVE", []byte(`{"quantity":5}`), []byte(`{"quantity":7}`), "Niko", "salah rak", createdAt))
	mock.ExpectQuery("COUNT").WithArgs((*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil), false).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	repo := NewAuditRepository(mock)
	audits, total, err := repo.FindAll(context.Background(), AuditFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(audits) != 1 || audits[0].Action != "MOVE" || string(audits[0].BeforeData) != `{"quantity":5}` {
		t.Errorf("unexpected audits %+v total %d", audits, total)
	}
}

func TestAuditFindBatches(t *testing.T) {
	mock := mustMockPool(t)
	createdAt := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("GROUP BY a.batch_id").WithArgs((*int)(nil), (*int)(nil), (*int)(nil), 10, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"batch_id", "action", "actor", "note", "item_count", "created_at",
		}).AddRow("11111111-1111-4111-8111-111111111111", "UPDATE", "Niko", "koreksi", 3, createdAt))
	mock.ExpectQuery("COUNT\\(DISTINCT").WithArgs((*int)(nil), (*int)(nil), (*int)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	repo := NewAuditRepository(mock)
	batches, total, err := repo.FindBatches(context.Background(), AuditFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(batches) != 1 || batches[0].ItemCount != 3 || batches[0].Action != "UPDATE" {
		t.Errorf("unexpected batches %+v total %d", batches, total)
	}
}

func TestAuditFindTimeline(t *testing.T) {
	mock := mustMockPool(t)
	createdAt := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM \\(SELECT \\* FROM singles").WithArgs(
		(*int)(nil), (*int)(nil), (*int)(nil), []string{"UPDATE"}, 10, 0,
	).WillReturnRows(pgxmock.NewRows([]string{
		"kind", "audit_id", "batch_id", "action", "actor", "note",
		"created_at", "item_count", "before_data", "after_data",
		"factor_before", "factor_after", "source_rak_name", "target_rak_name", "merged_count",
	}).
		AddRow("single", int64(11), "11111111-1111-4111-8111-111111111111", "UPDATE", "Niko", "koreksi", createdAt, 1, []byte(`{"quantity":3}`), []byte(`{"quantity":3}`), nil, nil, "", "", 0).
		AddRow("batch", int64(0), "22222222-2222-4222-8222-222222222222", "UPDATE", "Niko", "massal", createdAt, 4, []byte("null"), []byte("null"), int64(1), int64(0), "", "", 0))
	mock.ExpectQuery("HAVING \\(\\$4::text\\[\\] IS NULL OR MIN\\(a.action\\)").
		WithArgs((*int)(nil), (*int)(nil), (*int)(nil), []string{"UPDATE"}).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

	repo := NewAuditRepository(mock)
	items, total, err := repo.FindTimeline(context.Background(), AuditFilter{Actions: []string{"UPDATE"}, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("unexpected timeline %+v total %d", items, total)
	}
	if items[0].Kind != "single" || items[0].AuditID == nil || *items[0].AuditID != 11 || items[0].BatchID == "" {
		t.Errorf("expected collapsed single with id and batch, got %+v", items[0])
	}
	if items[1].Kind != "batch" || items[1].AuditID != nil || items[1].ItemCount != 4 {
		t.Errorf("expected batch without audit id, got %+v", items[1])
	}
	if items[1].FactorAfter == nil || *items[1].FactorAfter != 0 {
		t.Errorf("expected batch factor after 0, got %+v", items[1].FactorAfter)
	}
	if items[1].FactorBefore == nil || *items[1].FactorBefore != 1 {
		t.Errorf("expected batch factor before 1, got %+v", items[1].FactorBefore)
	}
}
