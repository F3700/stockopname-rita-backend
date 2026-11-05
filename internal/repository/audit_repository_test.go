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
	mock.ExpectQuery("FROM stock_opname_audit").WithArgs((*int)(nil), (*int)(nil), (*int)(nil), 10, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"audit_id", "batch_id", "stock_opname_id", "sesi_id",
			"source_rak_id", "target_rak_id", "source_coor_id", "target_coor_id",
			"action", "before_data", "after_data", "actor", "note", "created_at",
		}).AddRow(int64(1), "", 7, 3, nil, nil, nil, nil, "MOVE", []byte(`{"quantity":5}`), []byte(`{"quantity":7}`), "Niko", "salah rak", createdAt))
	mock.ExpectQuery("COUNT").WithArgs((*int)(nil), (*int)(nil), (*int)(nil)).
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
