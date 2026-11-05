package service

import (
	"context"
	"testing"
	"time"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func TestAuditFindAllMapsRowsAndPagination(t *testing.T) {
	svc := NewAuditService(&fakeAuditRepository{
		findAllFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*model.StockOpnameAudit, int, error) {
			if filter.Limit != 10 || filter.Offset != 10 {
				t.Errorf("unexpected filter %+v", filter)
			}
			return []*model.StockOpnameAudit{
				{
					AuditID:    1,
					Action:     "MOVE",
					SesiID:     3,
					Actor:      "Niko",
					BeforeData: []byte(`{"quantity":5}`),
					AfterData:  []byte(`{"quantity":7}`),
					CreatedAt:  time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC),
				},
			}, 11, nil
		},
	})

	pagination := &dto.Pagination{Page: 2, Limit: 10}
	rows, err := svc.FindAll(context.Background(), repository.AuditFilter{}, pagination)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].Action != "MOVE" || string(rows[0].Before) != `{"quantity":5}` {
		t.Errorf("unexpected rows %+v", rows)
	}
	if pagination.TotalItems != 11 || pagination.TotalPages != 2 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}
