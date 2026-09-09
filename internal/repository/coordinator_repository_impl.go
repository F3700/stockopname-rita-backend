package repository

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type CoordinatorRepositoryImpl struct {
}

func NewCoordinatorRepository() CoordinatorRepository {
	return &CoordinatorRepositoryImpl{}
}

func (c *CoordinatorRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	const SQL = `
		UPDATE coordinator
		SET coor_status = $1
		WHERE coor_id = $2
	`
	result, err := tx.Exec(ctx, SQL, coordinator.CoorStatus, coordinator.CoorID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("coordinator with id %d not found", coordinator.CoorID)
	}
	return err
}

// FindAll implements [CoordinatorRepository].
func (c *CoordinatorRepositoryImpl) FindAllSummary(ctx context.Context, tx pgx.Tx) ([]*model.CoordinatorSummary, error) {
	const SQL = `
		SELECT
			c.coor_id AS id,
			c.coor_code AS code,
			COUNT(DISTINCT i.inspector_id) AS inspector,
			COUNT(DISTINCT r.rak_id) AS rack_assigned,
			COUNT(DISTINCT r.rak_id) FILTER (
				WHERE so.stock_opname_id IS NOT NULL
			) AS rack_completed,
			c.coor_status AS status
		FROM coordinator c
		LEFT JOIN inspector i
			ON i.inspector_coor_id = c.coor_id
		LEFT JOIN rak r
			ON r.rak_inspector_id = i.inspector_id
		LEFT JOIN stock_opname so
			ON so.stock_opname_rak_id = r.rak_id
		GROUP BY
			c.coor_id,
			c.coor_code,
			c.coor_status
		ORDER BY c.coor_id;
	`
	rows, err := tx.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coordinatorsSummary []*model.CoordinatorSummary
	for rows.Next() {
		var coordinatorSummary model.CoordinatorSummary
		err := rows.Scan(&coordinatorSummary.ID, &coordinatorSummary.Code, &coordinatorSummary.Inspector, &coordinatorSummary.RackAssigned, &coordinatorSummary.RackCompleted, &coordinatorSummary.Status)
		if err != nil {
			return nil, err
		}
		coordinatorsSummary = append(coordinatorsSummary, &coordinatorSummary)
	}

	return coordinatorsSummary, nil
}

// FindById implements [CoordinatorRepository].
func (c *CoordinatorRepositoryImpl) FindByIdSummary(ctx context.Context, tx pgx.Tx, id int) (*model.CoordinatorSummary, error) {
	const SQL = `
		SELECT
			c.coor_id AS id,
			c.coor_code AS code,
			COUNT(DISTINCT i.inspector_id) AS inspector,
			COUNT(DISTINCT r.rak_id) AS rack_assigned,
			COUNT(DISTINCT r.rak_id) FILTER (
				WHERE so.stock_opname_id IS NOT NULL
			) AS rack_completed,
			c.coor_status AS status
		FROM coordinator c
		LEFT JOIN inspector i
			ON i.inspector_coor_id = c.coor_id
		LEFT JOIN rak r
			ON r.rak_inspector_id = i.inspector_id
		LEFT JOIN stock_opname so
			ON so.stock_opname_rak_id = r.rak_id
		WHERE c.coor_id = $1
		GROUP BY
			c.coor_id,
			c.coor_code,
			c.coor_status
		ORDER BY c.coor_id;
	`
	row := tx.QueryRow(ctx, SQL, id)
	var coordinatorSummary model.CoordinatorSummary
	err := row.Scan(&coordinatorSummary.ID, &coordinatorSummary.Code, &coordinatorSummary.Inspector, &coordinatorSummary.RackAssigned, &coordinatorSummary.RackCompleted, &coordinatorSummary.Status)
	if err != nil {
		return nil, err
	}
	return &coordinatorSummary, nil
}

// FindBySesiId implements [CoordinatorRepository].
func (c *CoordinatorRepositoryImpl) FindBySesiIdSummary(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.CoordinatorSummary, error) {
	const SQL = `
		SELECT
			c.coor_id AS id,
			c.coor_code AS code,
			COUNT(DISTINCT i.inspector_id) AS inspector,
			COUNT(DISTINCT r.rak_id) AS rack_assigned,
			COUNT(DISTINCT r.rak_id) FILTER (
				WHERE so.stock_opname_id IS NOT NULL
			) AS rack_completed,
			c.coor_status AS status
		FROM coordinator c
		LEFT JOIN inspector i
			ON i.inspector_coor_id = c.coor_id
		LEFT JOIN rak r
			ON r.rak_inspector_id = i.inspector_id
		LEFT JOIN stock_opname so
			ON so.stock_opname_rak_id = r.rak_id
		WHERE c.coor_sesi_id = $1
		GROUP BY
			c.coor_id,
			c.coor_code,
			c.coor_status
		ORDER BY c.coor_id;
	`
	rows, err := tx.Query(ctx, SQL, sesiId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coordinatorsSummary []*model.CoordinatorSummary
	for rows.Next() {
		var coordinatorSummary model.CoordinatorSummary
		err := rows.Scan(&coordinatorSummary.ID, &coordinatorSummary.Code, &coordinatorSummary.Inspector, &coordinatorSummary.RackAssigned, &coordinatorSummary.RackCompleted, &coordinatorSummary.Status)
		if err != nil {
			return nil, err
		}
		coordinatorsSummary = append(coordinatorsSummary, &coordinatorSummary)
	}

	return coordinatorsSummary, nil
}

// Save implements [CoordinatorRepository].
func (c *CoordinatorRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	const SQL = `
		INSERT INTO coordinator (coor_code, coor_sesi_id, coor_status)
		VALUES ($1, $2, $3)
	`
	_, err := tx.Exec(ctx, SQL, coordinator.CoorCode, coordinator.CoorSesiID, coordinator.CoorStatus)
	return err
}
