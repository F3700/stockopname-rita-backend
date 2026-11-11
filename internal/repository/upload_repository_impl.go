package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type UploadRepositoryImpl struct {
	Pool DBPool
}

func NewUploadRepository(pool DBPool) UploadRepository {
	return &UploadRepositoryImpl{Pool: pool}
}

// Save implements [UploadRepository].
func (r *UploadRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, upload *model.Upload) error {
	const SQL = `
		INSERT INTO upload (upload_coor_id)
		VALUES ($1)
		RETURNING upload_id, upload_createdat
	`
	return tx.QueryRow(ctx, SQL, upload.CoorID).Scan(&upload.UploadID, &upload.CreatedAt)
}

// FindAllSummary implements [UploadRepository]. Counts come from row
// ownership, not from the racks the upload once claimed, so an item moved to
// another rack still counts here. Uploads without any owned row are hidden, but
// their row is kept for stable upload numbers.
func (r *UploadRepositoryImpl) FindAllSummary(ctx context.Context, coordinatorId int) ([]*model.UploadSummary, error) {
	const SQL = `
		WITH numbered AS (
			SELECT
				u.upload_id,
				ROW_NUMBER() OVER (ORDER BY u.upload_id) AS upload_number,
				u.upload_coor_id,
				u.upload_createdat
			FROM upload u
			WHERE u.upload_coor_id = $1
		),
		stats AS (
			SELECT
				so.so_upload_id AS upload_id,
				COUNT(DISTINCT so.stock_opname_rak_id) AS rack_count,
				COUNT(*) AS item_count
			FROM stock_opname so
			WHERE so.so_upload_id IN (SELECT upload_id FROM numbered)
			GROUP BY so.so_upload_id
		)
		SELECT
			n.upload_id AS UploadID,
			n.upload_number AS UploadNumber,
			n.upload_coor_id AS CoorID,
			s.rack_count AS RackCount,
			s.item_count AS ItemCount,
			n.upload_createdat AS CreatedAt
		FROM numbered n
		JOIN stats s ON s.upload_id = n.upload_id
		ORDER BY n.upload_createdat DESC, n.upload_id DESC;
	`
	rows, err := r.Pool.Query(ctx, SQL, coordinatorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	uploads, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.UploadSummary])
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

// FindSummaryByCoordinator returns one row per upload with totals from row
// ownership. Rows without an owner are grouped under UploadID 0 ("Belum
// upload"), so the coordinator total always reconciles.
func (r *UploadRepositoryImpl) FindSummaryByCoordinator(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error) {
	const SQL = `
		WITH numbered AS (
			SELECT
				u.upload_id,
				ROW_NUMBER() OVER (
					PARTITION BY u.upload_coor_id
					ORDER BY u.upload_id
				) AS upload_number
			FROM upload u
			WHERE u.upload_coor_id = $1
		)
		SELECT
			c.coor_code AS CoordinatorCode,
			COALESCE(so.so_upload_id, 0) AS UploadID,
			COALESCE(n.upload_number, 0) AS UploadNumber,
			COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor), 0) AS Qty,
			COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor * so.so_buyprice), 0) AS TotalBuy,
			COALESCE(SUM(so.stock_opname_quantity * so.so_qty_factor * so.so_sellprice), 0) AS TotalSell
		FROM stock_opname so
		JOIN rak r ON r.rak_id = so.stock_opname_rak_id
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		LEFT JOIN numbered n ON n.upload_id = so.so_upload_id
		WHERE c.coor_id = $1
		GROUP BY c.coor_code, COALESCE(so.so_upload_id, 0), COALESCE(n.upload_number, 0)
		ORDER BY (COALESCE(so.so_upload_id, 0) = 0), COALESCE(so.so_upload_id, 0);
	`
	rows, err := r.Pool.Query(ctx, SQL, coordinatorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.UploadReportRow])
}

// FindDetail implements [UploadRepository]. Items are grouped by the rack they
// currently sit in, and racks with no item of this upload simply do not appear.
// An item moved to another rack still shows up here, under its new rack.
func (r *UploadRepositoryImpl) FindDetail(ctx context.Context, uploadID int) (*model.UploadDetail, error) {
	const headerSQL = `
		SELECT
			u.upload_id,
			u.upload_coor_id,
			u.upload_createdat,
			c.coor_code,
			s.sesi_code,
			s.sesi_location
		FROM upload u
		JOIN coordinator c ON c.coor_id = u.upload_coor_id
		JOIN sesi s ON s.sesi_id = c.coor_sesi_id
		WHERE u.upload_id = $1
	`
	var detail model.UploadDetail
	err := r.Pool.QueryRow(ctx, headerSQL, uploadID).Scan(
		&detail.UploadID,
		&detail.CoorID,
		&detail.CreatedAt,
		&detail.CoordinatorCode,
		&detail.SessionCode,
		&detail.SessionLocation,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &model.NotFoundError{Resource: "Upload", ID: uploadID}
	}
	if err != nil {
		return nil, err
	}

	const racksSQL = `
		SELECT
			r.rak_id,
			r.rak_name,
			so.so_barcode AS barcode,
			so.so_product_plu AS plu,
			so.so_product_name AS name,
			so.so_product_dept AS dept,
			so.stock_opname_quantity * so.so_qty_factor AS qty,
			so.so_buyprice AS buy_price,
			so.so_sellprice AS sell_price,
			COALESCE(p.product_per_carton, 0) AS per_carton
		FROM stock_opname so
		JOIN rak r ON r.rak_id = so.stock_opname_rak_id
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		LEFT JOIN product p
			ON p.product_sesi_id = c.coor_sesi_id
			AND p.product_plu = so.so_product_plu
		WHERE so.so_upload_id = $1
		ORDER BY r.rak_name, so.so_product_name;
	`
	rows, err := r.Pool.Query(ctx, racksSQL, uploadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var current *model.UploadRack
	flush := func() {
		if current == nil {
			return
		}
		current.ItemCount = len(current.Items)
		detail.Racks = append(detail.Racks, *current)
		detail.RackCount++
		detail.ItemCount += current.ItemCount
		current = nil
	}

	for rows.Next() {
		var rakID int
		var rakName string
		var item model.UploadItem
		if err := rows.Scan(
			&rakID,
			&rakName,
			&item.Barcode,
			&item.PLU,
			&item.Name,
			&item.Dept,
			&item.Qty,
			&item.BuyPrice,
			&item.SellPrice,
			&item.PerCarton,
		); err != nil {
			return nil, err
		}
		if current == nil || current.RakID != rakID {
			flush()
			current = &model.UploadRack{
				RakID:   rakID,
				RakName: rakName,
				Items:   make([]model.UploadItem, 0),
			}
		}
		current.Items = append(current.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	flush()

	return &detail, nil
}
