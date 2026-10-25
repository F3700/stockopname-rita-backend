package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newRackServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) RackService {
	t.Helper()
	return NewRackService(
		repository.NewRackRepository(mock),
		repository.NewInspectorRepository(mock),
		repository.NewSesiRepository(mock),
		mock,
	)
}

func expectRackSession(mock pgxmock.PgxPoolIface, inspectorId int, sesiId int, status string) {
	mock.ExpectQuery("FROM inspector").WithArgs(inspectorId).WillReturnRows(
		pgxmock.NewRows([]string{"coor_sesi_id"}).AddRow(sesiId),
	)
	mock.ExpectQuery("FOR SHARE").WithArgs(sesiId).WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
			AddRow(sesiId, "Gudang A", "SESI-01", status, time.Now(), nil),
	)
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
	expectRackSession(mock, 3, 7, "IN_PROGRESS")
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

func TestRackCreateSessionFinal(t *testing.T) {
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
	expectRackSession(mock, 3, 7, "COMPLETED")
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	_, err = svc.CreateRack(context.Background(), dto.CreateRackRequest{InspectorID: 3, RackName: "R9"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestRackCreateInspectorNotFound(t *testing.T) {
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
	mock.ExpectQuery("FROM inspector").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"coor_sesi_id"}),
	)
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	_, err = svc.CreateRack(context.Background(), dto.CreateRackRequest{InspectorID: 99, RackName: "R9"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
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
	expectRackSession(mock, 3, 7, "IN_PROGRESS")
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R9", 3).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	_, err = svc.CreateRack(context.Background(), dto.CreateRackRequest{InspectorID: 3, RackName: "R9"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
