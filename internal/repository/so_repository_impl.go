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
func (s *StockOpnameRepositoryImpl) FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			p.product_barcode AS barcode,
			p.product_name AS name,
			so.stock_opname_quantity AS quantity,
			r.rak_name AS rackName,
			i.inspector_code AS inspectorCode,
			c.coor_code AS coordinatorCode,
			so.stock_opname_updatedat AS "updatedAt"
		FROM stock_opname so
		JOIN product p
			ON so.stock_opname_product_id = p.product_id
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id
		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id
		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR c.coor_sesi_id = $2::int)
		AND (p.product_barcode ILIKE $3 OR p.product_name ILIKE $3)
		ORDER BY so.stock_opname_id DESC
		LIMIT $4 OFFSET $5;
	`

	rows, err := s.Pool.Query(ctx, SQL, coorId, sesiId, "%"+search+"%", limit, offset)
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
		JOIN product p
			ON so.stock_opname_product_id = p.product_id
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id
		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id
		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR c.coor_sesi_id = $2::int)
		AND (p.product_barcode ILIKE $3 OR p.product_name ILIKE $3)
	`
	if err := s.Pool.QueryRow(ctx, countSQL, coorId, sesiId, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return stockOpnames, total, nil
}

// FindById implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			p.product_barcode AS barcode,
			p.product_name,
			so.stock_opname_quantity AS quantity,
			r.rak_name,
			i.inspector_code,
			c.coor_code,
			so.stock_opname_updatedat AS "updatedAt"

		FROM stock_opname so

		JOIN product p
			ON so.stock_opname_product_id = p.product_id

		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id

		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id

		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id

		WHERE so.stock_opname_id = $1;
	`

	var stockOpname model.StockOpnameSummary
	err := tx.QueryRow(ctx, SQL, id).Scan(&stockOpname.Id, &stockOpname.Barcode, &stockOpname.Name, &stockOpname.Quantity, &stockOpname.RackName, &stockOpname.InspectorCode, &stockOpname.CoordinatorCode, &stockOpname.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Stock Opname", ID: id}
		}
		return nil, err
	}

	return &stockOpname, nil
}

// Save implements [StockOpnameRepository].
func (s *StockOpnameRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	const SQL = `
		INSERT INTO stock_opname (stock_opname_quantity, stock_opname_product_id, stock_opname_rak_id)
		VALUES ($1, $2, $3)
		RETURNING stock_opname_id
	`
	err := tx.QueryRow(ctx, SQL, stockOpname.StockOpnameQuantity, stockOpname.StockOpnameProductID, stockOpname.StockOpnameRakID).Scan(&stockOpname.StockOpnameID)
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
// It returns the full denormalized result set of one session, including
// product prices, for file exports. Unlike FindAll it needs no pagination.
func (s *StockOpnameRepositoryImpl) FindAllForExport(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
	const SQL = `
		SELECT
			so.stock_opname_id AS id,
			p.product_barcode AS barcode,
			p.product_name AS name,
			p.product_buyprice AS buyPrice,
			p.product_sellprice AS sellPrice,
			so.stock_opname_quantity AS quantity,
			r.rak_name AS rackName,
			i.inspector_code AS inspectorCode,
			c.coor_code AS coordinatorCode
		FROM stock_opname so
		JOIN product p
			ON so.stock_opname_product_id = p.product_id
		JOIN rak r
			ON so.stock_opname_rak_id = r.rak_id
		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id
		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id
		WHERE c.coor_sesi_id = $1
		ORDER BY so.stock_opname_id;
	`

	rows, err := s.Pool.Query(ctx, SQL, sesiId)
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
