package repository

import (
	"context"
	"errors"
	"fmt"
	"stockopname-rita-backend/internal/model"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type StockOpnameRepositoryImpl struct {
	Pool        DBPool
	exportLimit int
}

// DefaultExportLimit bounds how many rows a single export may load. Exports are
// rendered fully in memory, so an unbounded result set is an OOM waiting for a
// large enough session.
const DefaultExportLimit = 20000

type StockOpnameRepositoryOption func(*StockOpnameRepositoryImpl)

// WithExportLimit overrides the row limit used by FindAllForExport. A
// non-positive limit disables the check.
func WithExportLimit(limit int) StockOpnameRepositoryOption {
	return func(r *StockOpnameRepositoryImpl) {
		r.exportLimit = limit
	}
}

func NewStockOpnameRepository(pool DBPool, opts ...StockOpnameRepositoryOption) StockOpnameRepository {
	repo := &StockOpnameRepositoryImpl{
		Pool:        pool,
		exportLimit: DefaultExportLimit,
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

// Delete implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `
		DELETE FROM stock_opname
		WHERE stock_opname_id = $1
	`
	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Stock Opname", err)
	}

	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Stock Opname", ID: id}
	}
	return nil
}

// FindAll implements [StockOpnameRepository].
// Product display data comes from the snapshot columns, so history
// never depends on the product master table.
func (s *StockOpnameRepositoryImpl) FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			so.so_barcode AS barcode,
			so.so_product_plu AS productPLU,
			so.so_product_name AS name,
			so.stock_opname_quantity AS quantity,
			so.stock_opname_quantity * so.so_qty_factor AS effective_quantity,
			so.stock_opname_rak_id AS rackId,
			r.rak_name AS rackName,
			c.coor_id AS coordinatorId,
			c.coor_code AS coordinatorCode,
			so.so_qty_factor AS qty_factor,
			so.stock_opname_updatedat AS "updatedAt"
		FROM stock_opname so
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id
		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR c.coor_sesi_id = $2::int)
		AND ($3::int IS NULL OR so.stock_opname_rak_id = $3::int)
		AND (so.so_barcode ILIKE $4 OR so.so_product_name ILIKE $4 OR so.so_product_plu ILIKE $4)
		ORDER BY so.stock_opname_id DESC
		LIMIT $5 OFFSET $6;
	`

	rows, err := s.Pool.Query(ctx, SQL, coorId, sesiId, rakId, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	stockOpnames, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.StockOpnameSummary])
	if err != nil {
		return nil, 0, err
	}

	var total int
	countSQL := `
		SELECT
			COUNT(*)
		FROM stock_opname so
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id
		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR c.coor_sesi_id = $2::int)
		AND ($3::int IS NULL OR so.stock_opname_rak_id = $3::int)
		AND (so.so_barcode ILIKE $4 OR so.so_product_name ILIKE $4 OR so.so_product_plu ILIKE $4)
	`
	if err := s.Pool.QueryRow(ctx, countSQL, coorId, sesiId, rakId, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return stockOpnames, total, nil
}

// FindById implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			so.so_barcode AS barcode,
			so.so_product_plu AS product_plu,
			so.so_product_name AS product_name,
			so.stock_opname_quantity AS quantity,
			so.stock_opname_quantity * so.so_qty_factor AS effective_quantity,
			so.stock_opname_rak_id AS rak_id,
			r.rak_name AS rak_name,
			c.coor_id AS coordinator_id,
			c.coor_code AS coor_code,
			so.so_qty_factor AS qty_factor,
			so.stock_opname_updatedat AS "updatedAt"

		FROM stock_opname so

		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id

		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id

		WHERE so.stock_opname_id = $1;
	`

	var stockOpname model.StockOpnameSummary
	err := tx.QueryRow(ctx, SQL, id).Scan(&stockOpname.Id, &stockOpname.Barcode, &stockOpname.ProductPLU, &stockOpname.Name, &stockOpname.Quantity, &stockOpname.EffectiveQuantity, &stockOpname.RackID, &stockOpname.RackName, &stockOpname.CoordinatorID, &stockOpname.CoordinatorCode, &stockOpname.QtyFactor, &stockOpname.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Stock Opname", ID: id}
		}
		return nil, err
	}

	return &stockOpname, nil
}

// Create implements [StockOpnameRepository]. It is a strict insert: the
// (rack, plu) pair must not exist yet. Web correction uses PATCH instead.
func (s *StockOpnameRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	const SQL = `
		INSERT INTO stock_opname (stock_opname_quantity, stock_opname_rak_id, so_product_plu, so_product_name, so_product_dept, so_barcode, so_buyprice, so_sellprice, so_qty_factor)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::smallint)
		ON CONFLICT (stock_opname_rak_id, so_product_plu) DO NOTHING
		RETURNING stock_opname_id
	`
	err := tx.QueryRow(ctx, SQL, stockOpname.StockOpnameQuantity, stockOpname.StockOpnameRakID, stockOpname.SoProductPLU, stockOpname.SoProductName, stockOpname.SoProductDept, stockOpname.SoBarcode, stockOpname.SoBuyPrice, stockOpname.SoSellPrice, stockOpname.SoQtyFactor).Scan(&stockOpname.StockOpnameID)
	if errors.Is(err, pgx.ErrNoRows) {
		return &model.ConflictError{
			Resource: "Stock Opname",
			Detail:   "Produk sudah ada di rak ini; gunakan Edit untuk mengubah quantity",
		}
	}
	if err != nil {
		return model.MapPgError("Stock Opname", err)
	}
	return nil
}

// SaveMany implements [StockOpnameRepository]. It is an idempotent upsert on
// (rak_id, product_plu) for a whole upload batch in one round trip: pointers are
// unpacked into arrays and handed to a single INSERT ... SELECT FROM unnest.
func (s *StockOpnameRepositoryImpl) SaveMany(ctx context.Context, tx pgx.Tx, stockOpnames []*model.StockOpname) error {
	if len(stockOpnames) == 0 {
		return nil
	}

	const SQL = `
		INSERT INTO stock_opname (stock_opname_quantity, stock_opname_rak_id, so_product_plu, so_product_name, so_product_dept, so_barcode, so_buyprice, so_sellprice, so_qty_factor)
		SELECT * FROM unnest(
			$1::numeric[], $2::int[], $3::text[], $4::text[], $5::text[],
			$6::text[], $7::numeric[], $8::numeric[], $9::smallint[]
		)
		ON CONFLICT (stock_opname_rak_id, so_product_plu) DO UPDATE
		SET stock_opname_quantity = EXCLUDED.stock_opname_quantity,
		    so_barcode = EXCLUDED.so_barcode,
		    so_product_name = EXCLUDED.so_product_name,
		    so_product_dept = EXCLUDED.so_product_dept,
		    so_buyprice = EXCLUDED.so_buyprice,
		    so_sellprice = EXCLUDED.so_sellprice,
		    stock_opname_updatedat = NOW()
	`

	quantities := make([]float64, len(stockOpnames))
	rakIDs := make([]int32, len(stockOpnames))
	plus := make([]string, len(stockOpnames))
	names := make([]string, len(stockOpnames))
	depts := make([]string, len(stockOpnames))
	barcodes := make([]string, len(stockOpnames))
	buyPrices := make([]float64, len(stockOpnames))
	sellPrices := make([]float64, len(stockOpnames))
	factors := make([]int16, len(stockOpnames))

	for i, stockOpname := range stockOpnames {
		quantities[i] = stockOpname.StockOpnameQuantity
		rakIDs[i] = int32(stockOpname.StockOpnameRakID)
		plus[i] = stockOpname.SoProductPLU
		names[i] = stockOpname.SoProductName
		depts[i] = stockOpname.SoProductDept
		barcodes[i] = stockOpname.SoBarcode
		buyPrices[i] = stockOpname.SoBuyPrice
		sellPrices[i] = stockOpname.SoSellPrice
		factors[i] = int16(stockOpname.SoQtyFactor)
	}

	_, err := tx.Exec(ctx, SQL, quantities, rakIDs, plus, names, depts, barcodes, buyPrices, sellPrices, factors)
	if err != nil {
		return model.MapPgError("Stock Opname", err)
	}
	return nil
}

// Update implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	const SQL = `
		UPDATE stock_opname
		SET stock_opname_quantity = $1, stock_opname_updatedat = NOW()
		WHERE stock_opname_id = $2
	`
	res, err := tx.Exec(ctx, SQL, stockOpname.StockOpnameQuantity, stockOpname.StockOpnameID)
	if err != nil {
		return model.MapPgError("Stock Opname", err)
	}

	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Stock Opname", ID: stockOpname.StockOpnameID}
	}
	return nil
}

// FindAllForExport implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]*model.StockOpnameExport, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			so.so_barcode AS barcode,
			so.so_product_plu AS ProductPLU,
			so.so_product_name AS name,
			COALESCE(p.product_department_code, so.so_product_dept, '') AS departmentCode,
			so.so_buyprice AS buyPrice,
			so.so_sellprice AS sellPrice,
			so.so_qty_factor AS qty_factor,
			COALESCE(p.product_per_carton, 0) AS per_carton,
			so.stock_opname_quantity AS quantity,
			so.stock_opname_quantity * so.so_qty_factor AS effective_quantity,
			r.rak_name AS rackName,
			c.coor_code AS coordinatorCode,
			so.stock_opname_updatedat AS "updatedAt"
		FROM stock_opname so
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id
		LEFT JOIN product p
			ON p.product_sesi_id = c.coor_sesi_id
			AND p.product_plu = so.so_product_plu
		WHERE ($1::int IS NULL OR c.coor_sesi_id = $1::int)
		AND ($2::int IS NULL OR c.coor_id = $2::int)
		ORDER BY c.coor_code,
			regexp_replace(r.rak_name, '[0-9]+$', ''),
			(substring(r.rak_name from '[0-9]+$'))::int,
			r.rak_name,
			so.so_product_name
	`

	query := SQL
	args := []any{sesiId, coorId}
	if s.exportLimit > 0 {
		// Fetch one row past the limit so an oversized result set is detectable
		// without a second COUNT query.
		query += ` LIMIT $3`
		args = append(args, s.exportLimit+1)
	}

	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.StockOpnameExport])
	if err != nil {
		return nil, err
	}

	if s.exportLimit > 0 && len(exports) > s.exportLimit {
		return nil, &model.TooLargeError{Resource: "Laporan stock opname", Limit: s.exportLimit}
	}

	return exports, nil
}

// FindSesiIdByRak implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) FindSesiIdByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error) {
	const SQL = `
		SELECT c.coor_sesi_id
		FROM rak r
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		WHERE r.rak_id = $1
	`
	var sesiId int
	if err := tx.QueryRow(ctx, SQL, rakID).Scan(&sesiId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &model.NotFoundError{Resource: "Rack", ID: rakID}
		}
		return 0, err
	}
	return sesiId, nil
}

// FindSesiIdByStockOpnameID implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) FindSesiIdByStockOpnameID(ctx context.Context, tx pgx.Tx, stockOpnameID int) (int, error) {
	const SQL = `
		SELECT c.coor_sesi_id
		FROM stock_opname so
		JOIN rak r ON r.rak_id = so.stock_opname_rak_id
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		WHERE so.stock_opname_id = $1
	`
	var sesiId int
	if err := tx.QueryRow(ctx, SQL, stockOpnameID).Scan(&sesiId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &model.NotFoundError{Resource: "Stock opname", ID: stockOpnameID}
		}
		return 0, err
	}
	return sesiId, nil
}

// FindRackContextForUpdate locks one rack and returns its coordinator/upload
// context. Used by create and move.
func (s *StockOpnameRepositoryImpl) FindRackContextForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.StockOpnameRackContext, error) {
	const SQL = `
		SELECT
			r.rak_id,
			r.rak_name,
			c.coor_id,
			c.coor_code,
			c.coor_status,
			c.coor_sesi_id,
			r.rak_upload_id,
			u.upload_createdat
		FROM rak r
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		LEFT JOIN upload u ON u.upload_id = r.rak_upload_id
		WHERE r.rak_id = $1
		FOR UPDATE OF r
	`
	var rack model.StockOpnameRackContext
	var uploadID pgtype.Int4
	var uploadedAt pgtype.Timestamptz
	err := tx.QueryRow(ctx, SQL, rakID).Scan(
		&rack.RakID,
		&rack.RakName,
		&rack.CoorID,
		&rack.CoorCode,
		&rack.CoorStatus,
		&rack.SesiID,
		&uploadID,
		&uploadedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &model.NotFoundError{Resource: "Rack", ID: rakID}
	}
	if err != nil {
		return nil, err
	}
	rack.UploadID = int4Ptr(uploadID)
	if uploadedAt.Valid {
		value := uploadedAt.Time
		rack.UploadedAt = &value
	}
	return &rack, nil
}

const auditDataSelect = `
	SELECT
		so.stock_opname_id,
		so.stock_opname_quantity,
		so.so_product_plu,
		so.so_product_name,
		so.so_product_dept,
		so.so_barcode,
		so.so_buyprice,
		so.so_sellprice,
		so.so_qty_factor,
		r.rak_id,
		r.rak_name,
		c.coor_id,
		c.coor_code,
		c.coor_status,
		r.rak_upload_id,
		u.upload_createdat
	FROM stock_opname so
	JOIN rak r ON r.rak_id = so.stock_opname_rak_id
	JOIN coordinator c ON c.coor_id = r.rak_coor_id
	LEFT JOIN upload u ON u.upload_id = r.rak_upload_id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAuditData(row rowScanner) (*model.StockOpnameAuditData, error) {
	var data model.StockOpnameAuditData
	var uploadID pgtype.Int4
	var uploadedAt pgtype.Timestamptz
	if err := row.Scan(
		&data.StockOpnameID,
		&data.Quantity,
		&data.PLU,
		&data.Name,
		&data.DepartmentCode,
		&data.Barcode,
		&data.BuyPrice,
		&data.SellPrice,
		&data.QtyFactor,
		&data.RakID,
		&data.RakName,
		&data.CoordinatorID,
		&data.CoordinatorCode,
		&data.CoordinatorStatus,
		&uploadID,
		&uploadedAt,
	); err != nil {
		return nil, err
	}
	data.RakUploadID = int4Ptr(uploadID)
	if uploadedAt.Valid {
		value := uploadedAt.Time
		data.RakUploadedAt = &value
	}
	data.EffectiveQuantity = data.Quantity * float64(data.QtyFactor)
	return &data, nil
}

// FindByIdForUpdate locks one result row and returns its full snapshot plus
// the rack/coordinator context used by the audit log.
func (s *StockOpnameRepositoryImpl) FindByIdForUpdate(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameAuditData, error) {
	data, err := scanAuditData(tx.QueryRow(ctx, auditDataSelect+` WHERE so.stock_opname_id = $1 FOR UPDATE OF so`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &model.NotFoundError{Resource: "Stock Opname", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindForMove locks the source rows that will be moved.
func (s *StockOpnameRepositoryImpl) FindForMove(ctx context.Context, tx pgx.Tx, rakID int, ids []int) ([]*model.StockOpnameAuditData, error) {
	SQL := auditDataSelect + ` WHERE so.stock_opname_rak_id = $1`
	args := []any{rakID}
	if ids != nil {
		ids32 := make([]int32, len(ids))
		for i, id := range ids {
			ids32[i] = int32(id)
		}
		SQL += ` AND so.stock_opname_id = ANY($2)`
		args = append(args, ids32)
	}
	SQL += ` ORDER BY so.stock_opname_id FOR UPDATE OF so`

	rows, err := tx.Query(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.StockOpnameAuditData, 0)
	for rows.Next() {
		data, err := scanAuditData(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// FindByRakAndPLUForUpdate returns the target row when the same PLU already
// exists in the destination rack.
func (s *StockOpnameRepositoryImpl) FindByRakAndPLUForUpdate(ctx context.Context, tx pgx.Tx, rakID int, plu string) (*model.StockOpnameAuditData, error) {
	data, err := scanAuditData(tx.QueryRow(ctx, auditDataSelect+` WHERE so.stock_opname_rak_id = $1 AND so.so_product_plu = $2 FOR UPDATE OF so`, rakID, plu))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UpsertForMove inserts the moved row or adds its quantity to the existing
// (rack, plu) row. It returns the resulting target row fields.
func (s *StockOpnameRepositoryImpl) UpsertForMove(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) (*model.StockOpnameAuditData, error) {
	const SQL = `
		INSERT INTO stock_opname (stock_opname_quantity, stock_opname_rak_id, so_product_plu, so_product_name, so_product_dept, so_barcode, so_buyprice, so_sellprice, so_qty_factor)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::smallint)
		ON CONFLICT (stock_opname_rak_id, so_product_plu) DO UPDATE
		SET stock_opname_quantity = stock_opname.stock_opname_quantity + EXCLUDED.stock_opname_quantity,
		    so_product_dept = COALESCE(NULLIF(stock_opname.so_product_dept, ''), EXCLUDED.so_product_dept),
		    stock_opname_updatedat = NOW()
		RETURNING stock_opname_id, stock_opname_quantity, so_product_plu, so_product_name, so_product_dept, so_barcode, so_buyprice, so_sellprice, so_qty_factor
	`
	var data model.StockOpnameAuditData
	err := tx.QueryRow(ctx, SQL,
		stockOpname.StockOpnameQuantity,
		stockOpname.StockOpnameRakID,
		stockOpname.SoProductPLU,
		stockOpname.SoProductName,
		stockOpname.SoProductDept,
		stockOpname.SoBarcode,
		stockOpname.SoBuyPrice,
		stockOpname.SoSellPrice,
		stockOpname.SoQtyFactor,
	).Scan(
		&data.StockOpnameID,
		&data.Quantity,
		&data.PLU,
		&data.Name,
		&data.DepartmentCode,
		&data.Barcode,
		&data.BuyPrice,
		&data.SellPrice,
		&data.QtyFactor,
	)
	if err != nil {
		return nil, model.MapPgError("Stock Opname", err)
	}
	data.EffectiveQuantity = data.Quantity * float64(data.QtyFactor)
	return &data, nil
}

func (s *StockOpnameRepositoryImpl) CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error) {
	const SQL = `
		SELECT COUNT(*)
		FROM stock_opname so
		JOIN rak r ON r.rak_id = so.stock_opname_rak_id
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		WHERE c.coor_sesi_id = $1
	`
	var total int
	if err := tx.QueryRow(ctx, SQL, sesiId).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// CountByRak implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) CountByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error) {
	const SQL = `SELECT COUNT(*) FROM stock_opname WHERE stock_opname_rak_id = $1`
	var total int
	if err := tx.QueryRow(ctx, SQL, rakID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ReopenUploadIfEmpty clears the current upload pointer only when the rack is
// uploaded and has no results. The snapshot is intentionally left untouched:
// upload reports read current results, and the next upload overwrites it.
func (s *StockOpnameRepositoryImpl) ReopenUploadIfEmpty(ctx context.Context, tx pgx.Tx, rakID int) (bool, error) {
	const SQL = `
		UPDATE rak
		SET rak_upload_id = NULL
		WHERE rak_id = $1
		  AND rak_upload_id IS NOT NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM stock_opname so
		      WHERE so.stock_opname_rak_id = rak.rak_id
		  )
		RETURNING rak_id
	`
	var id int
	err := tx.QueryRow(ctx, SQL, rakID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

const qtyFactorFrom = `
	FROM stock_opname so
	JOIN rak r ON r.rak_id = so.stock_opname_rak_id
	JOIN coordinator c ON c.coor_id = r.rak_coor_id
	LEFT JOIN upload u ON u.upload_id = r.rak_upload_id
`

const qtyFactorWhere = `
	WHERE c.coor_sesi_id = $1
	AND (
		$2 = 'session'
		OR ($2 = 'coordinator' AND c.coor_id = $3::int)
		OR ($2 = 'rack' AND r.rak_id = $3::int)
		OR ($2 = 'item' AND so.stock_opname_id = $3::int)
		OR ($2 = 'items' AND so.stock_opname_id = ANY($4::int[]))
		OR (
			$2 = 'filter'
			AND ($5::text IS NULL OR so.so_barcode ILIKE $5 OR so.so_product_name ILIKE $5 OR so.so_product_plu ILIKE $5)
			AND ($6::int IS NULL OR c.coor_id = $6::int)
			AND ($7::int IS NULL OR r.rak_id = $7::int)
		)
	)
`

func qtyFactorArgs(sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) []any {
	var scopeIDArg any
	if scopeID != nil {
		scopeIDArg = *scopeID
	}

	var resultIDsArg any
	if len(resultIDs) > 0 {
		ids := make([]int32, len(resultIDs))
		for i, id := range resultIDs {
			ids[i] = int32(id)
		}
		resultIDsArg = ids
	}

	var searchArg any
	var coordinatorArg any
	var rakArg any
	if filter != nil {
		if search := strings.TrimSpace(filter.Search); search != "" {
			searchArg = "%" + search + "%"
		}
		if filter.CoordinatorID != nil {
			coordinatorArg = *filter.CoordinatorID
		}
		if filter.RakID != nil {
			rakArg = *filter.RakID
		}
	}

	return []any{sessionID, scope, scopeIDArg, resultIDsArg, searchArg, coordinatorArg, rakArg}
}

func (s *StockOpnameRepositoryImpl) CountForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) (int, error) {
	var total int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) `+qtyFactorFrom+qtyFactorWhere, qtyFactorArgs(sessionID, scope, scopeID, resultIDs, filter)...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (s *StockOpnameRepositoryImpl) FindForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) ([]*model.StockOpnameAuditData, error) {
	rows, err := tx.Query(ctx, auditDataSelect+qtyFactorWhere+` ORDER BY so.stock_opname_id FOR UPDATE OF so`, qtyFactorArgs(sessionID, scope, scopeID, resultIDs, filter)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.StockOpnameAuditData, 0)
	for rows.Next() {
		data, err := scanAuditData(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *StockOpnameRepositoryImpl) UpdateQtyFactor(ctx context.Context, tx pgx.Tx, id int, factor int) error {
	const SQL = `
		UPDATE stock_opname
		SET so_qty_factor = $1::smallint, stock_opname_updatedat = NOW()
		WHERE stock_opname_id = $2
	`
	res, err := tx.Exec(ctx, SQL, factor, id)
	if err != nil {
		return model.MapPgError("Stock Opname", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Stock Opname", ID: id}
	}
	return nil
}

func (s *StockOpnameRepositoryImpl) FindSellSummary(ctx context.Context, sessionID int, groupBy string) ([]*model.SellSummaryGroup, error) {
	var SQL string
	switch groupBy {
	case "session":
		SQL = `
			SELECT
				s.sesi_id::text AS key,
				s.sesi_code AS label,
				COUNT(so.stock_opname_id)::int AS item_count,
				COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor), 0)::float8 AS total_qty,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice), 0)::float8 AS total_sell_original,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice * so.so_qty_factor), 0)::float8 AS total_sell_effective
			FROM sesi s
			LEFT JOIN coordinator c ON c.coor_sesi_id = s.sesi_id
			LEFT JOIN rak r ON r.rak_coor_id = c.coor_id
			LEFT JOIN stock_opname so ON so.stock_opname_rak_id = r.rak_id
			WHERE s.sesi_id = $1
			GROUP BY s.sesi_id, s.sesi_code
		`
	case "coordinator":
		SQL = `
			SELECT
				c.coor_id::text AS key,
				c.coor_code AS label,
				COUNT(*)::int AS item_count,
				COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor), 0)::float8 AS total_qty,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice), 0)::float8 AS total_sell_original,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice * so.so_qty_factor), 0)::float8 AS total_sell_effective
			FROM stock_opname so
			JOIN rak r ON r.rak_id = so.stock_opname_rak_id
			JOIN coordinator c ON c.coor_id = r.rak_coor_id
			WHERE c.coor_sesi_id = $1
			GROUP BY c.coor_id, c.coor_code
			ORDER BY c.coor_code
		`
	case "rack":
		SQL = `
			SELECT
				r.rak_id::text AS key,
				r.rak_name || ' · ' || c.coor_code AS label,
				COUNT(*)::int AS item_count,
				COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor), 0)::float8 AS total_qty,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice), 0)::float8 AS total_sell_original,
				COALESCE(SUM(so.stock_opname_quantity * so.so_sellprice * so.so_qty_factor), 0)::float8 AS total_sell_effective
			FROM stock_opname so
			JOIN rak r ON r.rak_id = so.stock_opname_rak_id
			JOIN coordinator c ON c.coor_id = r.rak_coor_id
			WHERE c.coor_sesi_id = $1
			GROUP BY r.rak_id, r.rak_name, c.coor_id, c.coor_code
			ORDER BY c.coor_code,
				regexp_replace(r.rak_name, '[0-9]+$', ''),
				(substring(r.rak_name from '[0-9]+$'))::int,
				r.rak_name
		`
	default:
		return nil, fmt.Errorf("invalid groupBy: %s", groupBy)
	}

	rows, err := s.Pool.Query(ctx, SQL, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.SellSummaryGroup])
}
