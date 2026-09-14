package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type InspectorRepositoryImpl struct {
	Pool DBPool
}

func NewInspectorRepository(pool DBPool) InspectorRepository {
	return &InspectorRepositoryImpl{
		Pool: pool,
	}
}

// FindAll implements [InspectorRepository].
func (i *InspectorRepositoryImpl) FindAll(ctx context.Context) ([]*model.InspectorSummary, error) {
	const SQL = `
		SELECT
			i.inspector_id,
			i.inspector_code,
			COUNT(DISTINCT r.rak_id) AS rack_assigned,
			COUNT(DISTINCT r.rak_id)
        		FILTER (WHERE so.stock_opname_rak_id IS NOT NULL) AS rack_completed,
			COUNT(so.stock_opname_id) AS total_items
		FROM inspector i
		LEFT JOIN rak r ON i.inspector_id = r.rak_inspector_id
		LEFT JOIN stock_opname so ON r.rak_id = so.stock_opname_rak_id
		GROUP BY i.inspector_id, i.inspector_code
	`
	rows, err := i.Pool.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inspectors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.InspectorSummary])
	if err != nil {
		return nil, err
	}

	return inspectors, nil
}

// FindByCoorId implements [InspectorRepository].
func (i *InspectorRepositoryImpl) FindByCoorId(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error) {
	const SQL = `
		SELECT
			i.inspector_id,
			i.inspector_code,
			COUNT(DISTINCT r.rak_id) AS rack_assigned,
			COUNT(DISTINCT r.rak_id)
        		FILTER (WHERE so.stock_opname_rak_id IS NOT NULL) AS rack_completed,
			COUNT(so.stock_opname_id) AS total_items
		FROM inspector i
		LEFT JOIN rak r ON i.inspector_id = r.rak_inspector_id
		LEFT JOIN stock_opname so ON r.rak_id = so.stock_opname_rak_id
		WHERE i.inspector_coor_id = $1
		GROUP BY i.inspector_id, i.inspector_code
	`
	rows, err := i.Pool.Query(ctx, SQL, coordinatorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inspectors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.InspectorSummary])
	if err != nil {
		return nil, err
	}

	return inspectors, nil
}

func (i *InspectorRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, inspector *model.Inspector) error {
	const SQL = `
		INSERT INTO inspector (inspector_code, inspector_coor_id)
		VALUES ($1, $2)
		RETURNING inspector_id
	`
	err := tx.QueryRow(ctx, SQL, inspector.InspectorCode, inspector.CoordinatorID).Scan(&inspector.InspectorID)
	if err != nil {
		return model.MapPgError("Inspector", err)
	}
	return nil
}
