package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type StockOpnameRepositoryImpl struct {
	Pool DBPool
}

func NewStockOpnameRepository(pool DBPool) StockOpnameRepository {
	return &StockOpnameRepositoryImpl{
		Pool: pool,
	}
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
			so.so_product_name AS name,
			so.stock_opname_quantity AS quantity,
			r.rak_name AS rackName,
			c.coor_code AS coordinatorCode,
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
			so.so_product_name,
			so.stock_opname_quantity AS quantity,
			r.rak_name,
			c.coor_code,
			so.stock_opname_updatedat AS "updatedAt"

		FROM stock_opname so

		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id

		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id

		WHERE so.stock_opname_id = $1;
	`

	var stockOpname model.StockOpnameSummary
	err := tx.QueryRow(ctx, SQL, id).Scan(&stockOpname.Id, &stockOpname.Barcode, &stockOpname.Name, &stockOpname.Quantity, &stockOpname.RackName, &stockOpname.CoordinatorCode, &stockOpname.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Stock Opname", ID: id}
		}
		return nil, err
	}

	return &stockOpname, nil
}

// Save implements [StockOpnameRepository].
// It is an idempotent upsert on (rak_id, product_plu): correcting an item
// from the web refreshes the snapshot instead of failing with a
// unique-constraint conflict.
func (s *StockOpnameRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	const SQL = `
		INSERT INTO stock_opname (stock_opname_quantity, stock_opname_rak_id, so_product_plu, so_product_name, so_barcode, so_buyprice, so_sellprice)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (stock_opname_rak_id, so_product_plu) DO UPDATE
		SET stock_opname_quantity = EXCLUDED.stock_opname_quantity,
		    so_barcode = EXCLUDED.so_barcode,
		    so_product_name = EXCLUDED.so_product_name,
		    so_buyprice = EXCLUDED.so_buyprice,
		    so_sellprice = EXCLUDED.so_sellprice,
		    stock_opname_updatedat = NOW()
		RETURNING stock_opname_id
	`
	err := tx.QueryRow(ctx, SQL, stockOpname.StockOpnameQuantity, stockOpname.StockOpnameRakID, stockOpname.SoProductPLU, stockOpname.SoProductName, stockOpname.SoBarcode, stockOpname.SoBuyPrice, stockOpname.SoSellPrice).Scan(&stockOpname.StockOpnameID)
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
			so.so_buyprice AS buyPrice,
			so.so_sellprice AS sellPrice,
			so.stock_opname_quantity AS quantity,
			r.rak_name AS rackName,
			c.coor_code AS coordinatorCode,
			so.stock_opname_updatedat AS "updatedAt"
		FROM stock_opname so
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN coordinator c
			ON r.rak_coor_id = c.coor_id
		WHERE ($1::int IS NULL OR c.coor_sesi_id = $1::int)
		AND ($2::int IS NULL OR c.coor_id = $2::int)
		ORDER BY c.coor_code,
			regexp_replace(r.rak_name, '[0-9]+$', ''),
			(substring(r.rak_name from '[0-9]+$'))::int,
			r.rak_name,
			so.so_product_name;
	`

	rows, err := s.Pool.Query(ctx, SQL, sesiId, coorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.StockOpnameExport])
	if err != nil {
		return nil, err
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

// CountBySesi implements [StockOpnameRepository].
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
