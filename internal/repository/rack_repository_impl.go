package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type RackRepositoryImpl struct {
	Pool DBPool
}

func NewRackRepository(pool DBPool) RackRepository {
	return &RackRepositoryImpl{
		Pool: pool,
	}
}

// FindAll implements [RackRepository].
func (r *RackRepositoryImpl) FindAll(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error) {
	const SQL = `
		SELECT
			r.rak_id AS RackID,
			r.rak_name AS RackName,
			r.rak_inspector_id AS InspectorID

		FROM rak r

		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id

		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id

		JOIN sesi s
			ON c.coor_sesi_id = s.sesi_id

		WHERE ($1::int IS NULL OR i.inspector_id = $1::int)
		AND ($2::int IS NULL OR c.coor_id = $2::int)
		AND ($3::int IS NULL OR s.sesi_id = $3::int);
	`

	rows, err := r.Pool.Query(ctx, SQL, inspectorId, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	racks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.Rack])
	if err != nil {
		return nil, err
	}

	return racks, nil
}

// FindProgress implements [RackRepository].
func (r *RackRepositoryImpl) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error) {
	const SQL = `
		SELECT
			COUNT(DISTINCT r.rak_id) AS rack_assigned,

			COUNT(DISTINCT r.rak_id) FILTER (
				WHERE so.stock_opname_id IS NOT NULL
			) AS rack_completed,

			COUNT(DISTINCT so.stock_opname_id) AS total_items

		FROM rak r

		JOIN inspector i
			ON r.rak_inspector_id = i.inspector_id

		JOIN coordinator c
			ON i.inspector_coor_id = c.coor_id

		JOIN sesi s
			ON c.coor_sesi_id = s.sesi_id

		LEFT JOIN stock_opname so
			ON so.stock_opname_rak_id = r.rak_id

		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR s.sesi_id = $2::int);
	`
	rows, err := r.Pool.Query(ctx, SQL, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progress, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.RackProgress])
	if err != nil {
		return nil, err
	}

	return progress, nil
}

func (r *RackRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, rack *model.Rack) error {
	const SQL = `
		INSERT INTO rak (rak_name, rak_inspector_id)
		VALUES ($1, $2)
		RETURNING rak_id
	`
	err := tx.QueryRow(ctx, SQL, rack.RackName, rack.InspectorID).Scan(&rack.RackID)
	if err != nil {
		return model.MapPgError("Rack", err)
	}
	return nil
}
