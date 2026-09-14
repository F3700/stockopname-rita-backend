package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

func newRackServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) RackService {
	t.Helper()
	return NewRackService(repository.NewRackRepository(mock), mock)
}

func TestRackCreateTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R9", 3).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(31),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	res, err := svc.CreateRack(context.Background(), dto.CreateRackRequest{InspectorID: 3, RackName: "R9"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RackID != 31 || res.RackName != "R9" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestRackCreateTxError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R9", 3).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	_, err = svc.CreateRack(context.Background(), dto.CreateRackRequest{InspectorID: 3, RackName: "R9"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
