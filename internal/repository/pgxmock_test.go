package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func mustMockPool(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
	return mock
}

func mustBeginTx(t *testing.T, mock pgxmock.PgxPoolIface) pgx.Tx {
	t.Helper()
	mock.ExpectBegin()
	tx, err := mock.Begin(context.Background())
	if err != nil {
		t.Fatalf("failed to begin mock tx: %v", err)
	}
	return tx
}
