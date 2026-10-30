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
	mock.ExpectQuery("FROM upload").WithArgs(5).WillReturnRows(
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

func TestUploadFindAllSummaryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM upload").WithArgs(5).WillReturnError(errors.New("boom"))

	repo := NewUploadRepository(mock)
	_, err := repo.FindAllSummary(context.Background(), 5)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
