package repository

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestBarcodeSaveBatch(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "8991001010016").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewBarcodeRepository(mock)
	if err := repo.SaveBatch(context.Background(), tx, 11, []string{"1002515550011", "8991001010016"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarcodeReplace(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM barcode").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 2))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewBarcodeRepository(mock)
	if err := repo.Replace(context.Background(), tx, 11, []string{"1002515550011"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
