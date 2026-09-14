package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

func newInspectorServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) InspectorService {
	t.Helper()
	return NewInspectorService(
		repository.NewInspectorRepository(mock),
		repository.NewCoordinatorRepository(mock),
		repository.NewRackRepository(mock),
		mock,
	)
}

func TestInspectorCreateTx(t *testing.T) {
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
	mock.ExpectQuery("s.sesi_code").WithArgs("SESI-01", "KOR-01").WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(1, "KOR-01", 1, "IN_PROGRESS"),
	)
	mock.ExpectQuery("INSERT INTO inspector").WithArgs("INSP-01", 1).WillReturnRows(
		pgxmock.NewRows([]string{"inspector_id"}).AddRow(9),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R1", 9).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(21),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R2", 9).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(22),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	res, err := svc.CreateInspector(context.Background(), dto.InspectorRequest{SesiCode: "SESI-01", CoorCode: "KOR-01", InspectorCode: "INSP-01", Rak: []string{"R1", "R2"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.InspectorID != 9 {
		t.Errorf("expected inspector id 9, got %d", res.InspectorID)
	}
	if len(res.Rak) != 2 || res.Rak[0].RackID != 21 || res.Rak[1].RackName != "R2" {
		t.Errorf("unexpected racks %+v", res.Rak)
	}
}

func TestInspectorCreateTxCoordinatorNotFound(t *testing.T) {
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
	mock.ExpectQuery("s.sesi_code").WithArgs("SESI-01", "NOPE").WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspector(context.Background(), dto.InspectorRequest{SesiCode: "SESI-01", CoorCode: "NOPE", InspectorCode: "INSP-01", Rak: []string{"R1"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInspectorCreateTxRackError(t *testing.T) {
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
	mock.ExpectQuery("s.sesi_code").WithArgs("SESI-01", "KOR-01").WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(1, "KOR-01", 1, "IN_PROGRESS"),
	)
	mock.ExpectQuery("INSERT INTO inspector").WithArgs("INSP-01", 1).WillReturnRows(
		pgxmock.NewRows([]string{"inspector_id"}).AddRow(9),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R1", 9).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspector(context.Background(), dto.InspectorRequest{SesiCode: "SESI-01", CoorCode: "KOR-01", InspectorCode: "INSP-01", Rak: []string{"R1"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
