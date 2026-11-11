package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

// These are real-database tests for the upload ownership model. They need a
// migrated database:
//
//	TEST_DATABASE_URL=... go test ./internal/service -run TestUploadOwnership
type ownershipFixture struct {
	pool          *pgxpool.Pool
	soService     StockOpnameService
	uploadService UploadService
	sesiID        int
	coorA         int
	coorB         int
	rakA1         int
	rakA2         int
	rakB1         int
}

func newOwnershipFixture(t *testing.T) *ownershipFixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	f := &ownershipFixture{pool: pool}
	code := fmt.Sprintf("OW-%012d", time.Now().UnixNano()%1_000_000_000_000)

	if err := pool.QueryRow(ctx,
		`INSERT INTO sesi (sesi_location, sesi_code, sesi_status) VALUES ('Gudang uji', $1, 'IN_PROGRESS') RETURNING sesi_id`,
		code).Scan(&f.sesiID); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM sesi WHERE sesi_id = $1`, f.sesiID)
	})

	f.coorA = f.insertCoordinator(t, "A")
	f.coorB = f.insertCoordinator(t, "B")
	f.rakA1 = f.insertRack(t, f.coorA, "Rak A1")
	f.rakA2 = f.insertRack(t, f.coorA, "Rak A2")
	f.rakB1 = f.insertRack(t, f.coorB, "Rak B1")

	for _, product := range []struct{ plu, name string }{
		{"100251", "Sari Roti"},
		{"100252", "Teh Botol"},
		{"100253", "Kopi Bubuk"},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO product (product_sesi_id, product_plu, product_name, product_department_code, product_buyprice, product_sellprice)
			VALUES ($1, $2, $3, '1138', 14500, 17200)
		`, f.sesiID, product.plu, product.name); err != nil {
			t.Fatalf("insert product %s: %v", product.plu, err)
		}
	}

	validate := validator.New()
	soRepo := repository.NewStockOpnameRepository(pool)
	f.soService = NewStockOpnameService(
		soRepo,
		repository.NewProductRepository(pool),
		repository.NewSesiRepository(pool),
		repository.NewAuditRepository(pool),
		repository.NewCoordinatorRepository(pool),
		pool,
		validate,
	)
	f.uploadService = NewUploadService(
		repository.NewUploadRepository(pool),
		repository.NewRackRepository(pool),
		repository.NewCoordinatorRepository(pool),
		repository.NewSesiRepository(pool),
		repository.NewProductRepository(pool),
		soRepo,
		pool,
		validate,
	)
	return f
}

func (f *ownershipFixture) insertCoordinator(t *testing.T, suffix string) int {
	t.Helper()
	var id int
	if err := f.pool.QueryRow(context.Background(),
		`INSERT INTO coordinator (coor_code, coor_sesi_id, coor_status) VALUES ($1, $2, 'IN_PROGRESS') RETURNING coor_id`,
		"KOR-"+suffix, f.sesiID).Scan(&id); err != nil {
		t.Fatalf("insert coordinator: %v", err)
	}
	return id
}

func (f *ownershipFixture) insertRack(t *testing.T, coorID int, name string) int {
	t.Helper()
	var id int
	if err := f.pool.QueryRow(context.Background(),
		`INSERT INTO rak (rak_name, rak_coor_id) VALUES ($1, $2) RETURNING rak_id`,
		name, coorID).Scan(&id); err != nil {
		t.Fatalf("insert rack: %v", err)
	}
	return id
}

func (f *ownershipFixture) upload(t *testing.T, coorID int, rackID int, plus ...string) dto.UploadResponse {
	t.Helper()
	items := make([]dto.UploadItemRequest, 0, len(plus))
	for _, plu := range plus {
		items = append(items, dto.UploadItemRequest{PLU: plu, Quantity: qtyPtr(5)})
	}
	res, err := f.uploadService.Create(context.Background(), dto.CreateUploadRequest{
		CoordinatorID: coorID,
		Racks:         []dto.UploadRackRequest{{RakID: rackID, Items: items}},
	})
	if err != nil {
		t.Fatalf("upload rack %d: %v", rackID, err)
	}
	return res
}

func (f *ownershipFixture) move(t *testing.T, sourceRakID int, targetRakID int, resultIDs *[]int) dto.MoveStockOpnameResponse {
	t.Helper()
	res, err := f.soService.Move(context.Background(), dto.MoveStockOpnameRequest{
		SourceRakID: sourceRakID,
		TargetRakID: targetRakID,
		ResultIDs:   resultIDs,
	})
	if err != nil {
		t.Fatalf("move rack %d -> %d: %v", sourceRakID, targetRakID, err)
	}
	return res
}

// owner returns the upload that owns one (rack, plu) row, or nil when the row
// has no owner. It also reports the stored quantity.
func (f *ownershipFixture) owner(t *testing.T, rakID int, plu string) (*int, float64) {
	t.Helper()
	var owner *int
	var qty float64
	if err := f.pool.QueryRow(context.Background(),
		`SELECT so_upload_id, stock_opname_quantity FROM stock_opname WHERE stock_opname_rak_id = $1 AND so_product_plu = $2`,
		rakID, plu).Scan(&owner, &qty); err != nil {
		t.Fatalf("read owner of %s in rack %d: %v", plu, rakID, err)
	}
	return owner, qty
}

func (f *ownershipFixture) resultID(t *testing.T, rakID int, plu string) int {
	t.Helper()
	var id int
	if err := f.pool.QueryRow(context.Background(),
		`SELECT stock_opname_id FROM stock_opname WHERE stock_opname_rak_id = $1 AND so_product_plu = $2`,
		rakID, plu).Scan(&id); err != nil {
		t.Fatalf("read result id of %s in rack %d: %v", plu, rakID, err)
	}
	return id
}

func (f *ownershipFixture) host(t *testing.T, rakID int) *int {
	t.Helper()
	var host *int
	if err := f.pool.QueryRow(context.Background(),
		`SELECT rak_upload_id FROM rak WHERE rak_id = $1`, rakID).Scan(&host); err != nil {
		t.Fatalf("read host of rack %d: %v", rakID, err)
	}
	return host
}

func ownerIs(t *testing.T, owner *int, want int, label string) {
	t.Helper()
	if owner == nil {
		t.Fatalf("%s: expected owner %d, got none", label, want)
	}
	if *owner != want {
		t.Fatalf("%s: expected owner %d, got %d", label, want, *owner)
	}
}

// The main scenario: an item moved inside one coordinator keeps its upload, and
// the next upload of the destination rack does not steal it.
func TestUploadOwnershipMoveKeepsOwner(t *testing.T) {
	f := newOwnershipFixture(t)

	first := f.upload(t, f.coorA, f.rakA1, "100251", "100252")

	ids := []int{f.resultID(t, f.rakA1, "100251")}
	f.move(t, f.rakA1, f.rakA2, &ids)

	owner, _ := f.owner(t, f.rakA2, "100251")
	ownerIs(t, owner, first.UploadID, "moved row keeps its upload")

	second := f.upload(t, f.coorA, f.rakA2, "100253")

	owner, _ = f.owner(t, f.rakA2, "100251")
	ownerIs(t, owner, first.UploadID, "upload of the destination rack must not steal the row")
	owner, _ = f.owner(t, f.rakA2, "100253")
	ownerIs(t, owner, second.UploadID, "scanned row belongs to the new upload")

	firstDetail, err := f.uploadService.FindDetail(context.Background(), first.UploadID)
	if err != nil {
		t.Fatalf("detail of first upload: %v", err)
	}
	if firstDetail.RackCount != 2 || firstDetail.ItemCount != 2 {
		t.Fatalf("first upload should own 2 items in 2 racks, got %+v", firstDetail.UploadResponse)
	}
	if firstDetail.Racks[0].RakID != f.rakA1 || firstDetail.Racks[1].RakID != f.rakA2 {
		t.Fatalf("first upload should be grouped by current rack, got %+v", firstDetail.Racks)
	}

	secondDetail, err := f.uploadService.FindDetail(context.Background(), second.UploadID)
	if err != nil {
		t.Fatalf("detail of second upload: %v", err)
	}
	if secondDetail.ItemCount != 1 || secondDetail.Racks[0].Items[0].PLU != "100253" {
		t.Fatalf("second upload should own only its scan, got %+v", secondDetail.Racks)
	}
}

// Re-count: scanning a PLU that already belongs to another upload in that rack
// transfers the row and overwrites the quantity.
func TestUploadOwnershipRecountTransfersRow(t *testing.T) {
	f := newOwnershipFixture(t)

	first := f.upload(t, f.coorA, f.rakA1, "100251")
	ids := []int{f.resultID(t, f.rakA1, "100251")}
	f.move(t, f.rakA1, f.rakA2, &ids)

	// The destination rack has no host after adoption, so it can be claimed.
	if host := f.host(t, f.rakA2); host != nil {
		t.Fatalf("destination rack should have no host, got %v", *host)
	}

	second := f.upload(t, f.coorA, f.rakA2, "100251")

	owner, qty := f.owner(t, f.rakA2, "100251")
	ownerIs(t, owner, second.UploadID, "recount transfers the row")
	if qty != 5 {
		t.Fatalf("recount must overwrite the quantity, got %v", qty)
	}
	if owner != nil && *owner == first.UploadID {
		t.Fatal("the old upload must lose the row")
	}
}

// A claimed rack cannot be uploaded again while it still hosts an upload.
func TestUploadOwnershipRejectsClaimedRack(t *testing.T) {
	f := newOwnershipFixture(t)

	f.upload(t, f.coorA, f.rakA1, "100251")

	_, err := f.uploadService.Create(context.Background(), dto.CreateUploadRequest{
		CoordinatorID: f.coorA,
		Racks: []dto.UploadRackRequest{{
			RakID: f.rakA1,
			Items: []dto.UploadItemRequest{{PLU: "100252", Quantity: qtyPtr(1)}},
		}},
	})
	if err == nil {
		t.Fatal("expected the second upload of a claimed rack to fail")
	}
}

// Crossing coordinators always re-owns the row: to the target host when the rack
// has one, otherwise to nobody until someone uploads that rack.
func TestUploadOwnershipCrossCoordinatorMove(t *testing.T) {
	f := newOwnershipFixture(t)

	first := f.upload(t, f.coorA, f.rakA1, "100251")
	ids := []int{f.resultID(t, f.rakA1, "100251")}
	f.move(t, f.rakA1, f.rakB1, &ids)

	owner, _ := f.owner(t, f.rakB1, "100251")
	if owner != nil {
		t.Fatalf("row moved into a hostless rack must be unowned, got %d", *owner)
	}

	// The rack host is released once the source upload owns nothing there.
	if host := f.host(t, f.rakA1); host != nil {
		t.Fatalf("source rack host should be released, got %d", *host)
	}
	if first.UploadID == 0 {
		t.Fatal("unexpected empty upload id")
	}

	second := f.upload(t, f.coorB, f.rakB1, "100252")

	owner, _ = f.owner(t, f.rakB1, "100251")
	ownerIs(t, owner, second.UploadID, "the new host adopts the unowned row")
}

// Moving half of a rack keeps the source host; moving the rest releases it.
func TestUploadOwnershipReleasesHostOnlyWhenEmpty(t *testing.T) {
	f := newOwnershipFixture(t)

	first := f.upload(t, f.coorA, f.rakA1, "100251", "100252")

	ids := []int{f.resultID(t, f.rakA1, "100251")}
	f.move(t, f.rakA1, f.rakA2, &ids)
	if host := f.host(t, f.rakA1); host == nil || *host != first.UploadID {
		t.Fatalf("source host must stay while it still owns a row, got %v", host)
	}

	ids = []int{f.resultID(t, f.rakA1, "100252")}
	f.move(t, f.rakA1, f.rakA2, &ids)
	if host := f.host(t, f.rakA1); host != nil {
		t.Fatalf("source host must be released once it owns nothing, got %d", *host)
	}
}

// Two batches that claim the same racks at the same time must not deadlock:
// the claim locks the racks in id order first, so one batch wins and the other
// sees a conflict instead of a 40P01.
func TestUploadOwnershipConcurrentClaimsDoNotDeadlock(t *testing.T) {
	f := newOwnershipFixture(t)

	requests := []dto.CreateUploadRequest{
		{
			CoordinatorID: f.coorA,
			Racks: []dto.UploadRackRequest{
				{RakID: f.rakA1, Items: []dto.UploadItemRequest{{PLU: "100251", Quantity: qtyPtr(5)}}},
				{RakID: f.rakA2, Items: []dto.UploadItemRequest{{PLU: "100252", Quantity: qtyPtr(5)}}},
			},
		},
		{
			CoordinatorID: f.coorA,
			Racks: []dto.UploadRackRequest{
				{RakID: f.rakA2, Items: []dto.UploadItemRequest{{PLU: "100253", Quantity: qtyPtr(5)}}},
				{RakID: f.rakA1, Items: []dto.UploadItemRequest{{PLU: "100251", Quantity: qtyPtr(5)}}},
			},
		},
	}

	type outcome struct {
		err error
	}
	results := make(chan outcome, len(requests))
	start := make(chan struct{})
	for _, req := range requests {
		go func(req dto.CreateUploadRequest) {
			<-start
			_, err := f.uploadService.Create(context.Background(), req)
			results <- outcome{err: err}
		}(req)
	}
	close(start)

	succeeded, conflicted := 0, 0
	for range requests {
		got := <-results
		if got.err == nil {
			succeeded++
			continue
		}
		var conflict *model.ConflictError
		if !errors.As(got.err, &conflict) {
			t.Fatalf("concurrent upload failed with a non-conflict error: %v", got.err)
		}
		conflicted++
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("expected one winner and one conflict, got %d and %d", succeeded, conflicted)
	}
}
