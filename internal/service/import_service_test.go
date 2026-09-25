package service

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/valentin-kaiser/go-dbase/dbase"
	"golang.org/x/text/encoding/charmap"

	"stockopname-rita-backend/internal/repository"
)

type fixtureProdukRow struct {
	PLU     string  `dbase:"PLU"`
	NAMABRG string  `dbase:"NAMABRG"`
	KODEDPT string  `dbase:"KODE_DPT"`
	HRGJUAL float64 `dbase:"HRG_JUAL"`
	RPBELI  float64 `dbase:"RP_BELI"`
}

type fixtureBarcodeRow struct {
	PLU     string `dbase:"PLU"`
	BARCODE string `dbase:"BARCODE"`
}

// buildDBF creates a real dBase III file in memory using the same library
// the import service reads with.
func buildDBF(t *testing.T, filename string, columns []*dbase.Column, rows []interface{}) []byte {
	t.Helper()
	mem := newMemReadWriteSeeker()
	table, err := dbase.NewTable(
		dbase.FoxBasePlus,
		&dbase.Config{Filename: filename, Converter: dbase.NewDefaultConverter(charmap.Windows1252)},
		columns,
		0,
		dbase.GenericIO{Handle: mem},
	)
	if err != nil {
		t.Fatalf("create fixture table: %v", err)
	}
	for _, r := range rows {
		row, err := table.RowFromStruct(r)
		if err != nil {
			t.Fatalf("row from struct: %v", err)
		}
		if err := row.Add(); err != nil {
			t.Fatalf("add row: %v", err)
		}
	}
	if err := table.Close(); err != nil {
		t.Fatalf("close fixture table: %v", err)
	}
	data, err := compactDBFFile(mem.Data(), len(columns))
	if err != nil {
		t.Fatalf("compact fixture: %v", err)
	}
	return data
}

func mustColumn(t *testing.T, name string, dataType dbase.DataType, length uint8) *dbase.Column {
	t.Helper()
	column, err := dbase.NewColumn(name, dataType, length, 0, false)
	if err != nil {
		t.Fatalf("fixture column %s: %v", name, err)
	}
	return column
}

func produkFixture(t *testing.T, rows []interface{}) []byte {
	t.Helper()
	return buildDBF(t, "PRODUK.DBF", []*dbase.Column{
		mustColumn(t, "PLU", dbase.Character, 6),
		mustColumn(t, "NAMABRG", dbase.Character, 43),
		mustColumn(t, "KODE_DPT", dbase.Character, 5),
		mustColumn(t, "HRG_JUAL", dbase.Numeric, 8),
		mustColumn(t, "RP_BELI", dbase.Numeric, 8),
	}, rows)
}

func barcodeFixture(t *testing.T, rows []interface{}) []byte {
	t.Helper()
	return buildDBF(t, "BARCODE.DBF", []*dbase.Column{
		mustColumn(t, "PLU", dbase.Character, 6),
		mustColumn(t, "BARCODE", dbase.Character, 15),
	}, rows)
}

func toAny[T any](rows []T) []interface{} {
	out := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, r)
	}
	return out
}

func newImportServiceWithMock(mock pgxmock.PgxPoolIface) ImportService {
	return NewImportService(repository.NewProductRepository(mock), repository.NewBarcodeRepository(mock), mock)
}

func existingProductRows() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
		AddRow(11, "100251", "Sari Roti", "1138", 14500.0, 17200.0, ts, ts)
}

func TestImportInsertNew(t *testing.T) {
	produk := produkFixture(t, toAny([]fixtureProdukRow{
		{PLU: "100251", NAMABRG: "Sari Roti", KODEDPT: "1138", HRGJUAL: 14500, RPBELI: 17200},
	}))
	barcodes := barcodeFixture(t, toAny([]fixtureBarcodeRow{
		{PLU: "100251", BARCODE: "1002515550011"},
		{PLU: "100251", BARCODE: "8991001010016"},
	}))

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
	mock.ExpectQuery("FROM product p").WillReturnRows(pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}))
	mock.ExpectQuery("INSERT INTO product").WithArgs("100251", "Sari Roti", "1138", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"product_id"}).AddRow(11),
	)
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "8991001010016").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newImportServiceWithMock(mock)
	result, err := svc.Import(context.Background(), produk, barcodes, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProductsInserted != 1 || result.BarcodesAdded != 2 {
		t.Errorf("unexpected result %+v", result)
	}
	if len(result.InvalidRows) != 0 || len(result.OrphanBarcodes) != 0 {
		t.Errorf("expected no invalid/orphan, got %+v", result)
	}
}

func TestImportIdempotentDryRun(t *testing.T) {
	produk := produkFixture(t, toAny([]fixtureProdukRow{
		{PLU: "100251", NAMABRG: "Sari Roti", KODEDPT: "1138", HRGJUAL: 14500, RPBELI: 17200},
	}))
	barcodes := barcodeFixture(t, toAny([]fixtureBarcodeRow{
		{PLU: "100251", BARCODE: "1002515550011"},
		{PLU: "100251", BARCODE: "8991001010016"},
	}))

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
	mock.ExpectQuery("FROM product p").WillReturnRows(existingProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).
			AddRow(11, "1002515550011").
			AddRow(11, "8991001010016"),
	)
	mock.ExpectRollback()

	svc := newImportServiceWithMock(mock)
	result, err := svc.Import(context.Background(), produk, barcodes, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DryRun || result.ProductsUnchanged != 1 {
		t.Errorf("unexpected result %+v", result)
	}
	if result.ProductsInserted != 0 || result.ProductsUpdated != 0 || result.BarcodesAdded != 0 || result.BarcodesRemoved != 0 {
		t.Errorf("idempotent re-import must change nothing: %+v", result)
	}
}

func TestImportUpdateChanged(t *testing.T) {
	// Existing price 14000 -> fixture 14500, and one extra barcode.
	produk := produkFixture(t, toAny([]fixtureProdukRow{
		{PLU: "100251", NAMABRG: "Sari Roti", KODEDPT: "1138", HRGJUAL: 14500, RPBELI: 17200},
	}))
	barcodes := barcodeFixture(t, toAny([]fixtureBarcodeRow{
		{PLU: "100251", BARCODE: "1002515550011"},
		{PLU: "100251", BARCODE: "8991001010016"},
	}))

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("FROM product p").WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
			AddRow(11, "100251", "Sari Roti", "1138", 14000.0, 17200.0, ts, ts),
	)
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).AddRow(11, "1002515550011"),
	)
	mock.ExpectExec("UPDATE product").WithArgs("100251", "Sari Roti", "1138", 14500.0, 17200.0, 11).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("DELETE FROM barcode").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "8991001010016").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newImportServiceWithMock(mock)
	result, err := svc.Import(context.Background(), produk, barcodes, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProductsUpdated != 1 || result.BarcodesAdded != 1 || result.BarcodesRemoved != 0 {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestImportOrphanAndInvalid(t *testing.T) {
	produk := produkFixture(t, toAny([]fixtureProdukRow{
		{PLU: "100251", NAMABRG: "Sari Roti", KODEDPT: "1138", HRGJUAL: 14500, RPBELI: 17200},
		{PLU: "100252", NAMABRG: "", KODEDPT: "1138", HRGJUAL: 1000, RPBELI: 900},
	}))
	barcodes := barcodeFixture(t, toAny([]fixtureBarcodeRow{
		{PLU: "100251", BARCODE: "1002515550011"},
		{PLU: "999999", BARCODE: "9999995550011"},
		{PLU: "100251", BARCODE: ""},
	}))

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
	mock.ExpectQuery("FROM product p").WillReturnRows(pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}))
	mock.ExpectQuery("INSERT INTO product").WithArgs("100251", "Sari Roti", "1138", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"product_id"}).AddRow(11),
	)
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newImportServiceWithMock(mock)
	result, err := svc.Import(context.Background(), produk, barcodes, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProductsInserted != 1 {
		t.Errorf("unexpected inserted %+v", result)
	}
	if len(result.OrphanBarcodes) != 1 || result.OrphanBarcodes[0] != "999999|9999995550011" {
		t.Errorf("unexpected orphans %+v", result.OrphanBarcodes)
	}
	if len(result.InvalidRows) != 2 {
		t.Errorf("expected 2 invalid rows (empty name, empty barcode), got %+v", result.InvalidRows)
	}
}

func TestImportNoValidProdukRows(t *testing.T) {
	produk := produkFixture(t, toAny([]fixtureProdukRow{
		{PLU: "", NAMABRG: "Tanpa PLU", KODEDPT: "1138", HRGJUAL: 1000, RPBELI: 900},
	}))
	barcodes := barcodeFixture(t, toAny([]fixtureBarcodeRow{
		{PLU: "100251", BARCODE: "1002515550011"},
	}))

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	svc := newImportServiceWithMock(mock)
	_, err = svc.Import(context.Background(), produk, barcodes, false)
	if err == nil {
		t.Fatal("expected error for empty valid produk rows, got nil")
	}
}
