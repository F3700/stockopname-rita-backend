package repository

import (
	"context"
	"errors"
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
func (r *RackRepositoryImpl) FindAll(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error) {
	const SQL = `
		SELECT
			r.rak_id AS RakID,
			r.rak_name AS RakName,
			r.rak_coor_id AS CoorID,
			u.upload_createdat AS UploadedAt
		FROM rak r
		JOIN coordinator c ON c.coor_id = r.rak_coor_id
		LEFT JOIN upload u ON u.upload_id = r.rak_upload_id
		WHERE ($1::int IS NULL OR c.coor_id = $1::int)
		AND ($2::int IS NULL OR c.coor_sesi_id = $2::int)
		ORDER BY regexp_replace(r.rak_name, '[0-9]+$', ''),
			(substring(r.rak_name from '[0-9]+$'))::int,
			r.rak_name;
	`

	rows, err := r.Pool.Query(ctx, SQL, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	racks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.RackState])
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

		JOIN coordinator c
			ON c.coor_id = r.rak_coor_id

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
		INSERT INTO rak (rak_name, rak_coor_id)
		VALUES ($1, $2)
		RETURNING rak_id
	`
	err := tx.QueryRow(ctx, SQL, rack.RackName, rack.CoorID).Scan(&rack.RackID)
	if err != nil {
		return model.MapPgError("Rack", err)
	}
	return nil
}

// ClaimForUpload attaches every rack of one batch to the upload. Returns how
// many rows were claimed; a smaller count means a rack was already uploaded,
// belonged to another coordinator, or does not exist.
func (r *RackRepositoryImpl) ClaimForUpload(ctx context.Context, tx pgx.Tx, uploadID int, rakIDs []int32, coorID int) (int64, error) {
	const SQL = `
		UPDATE rak
		SET rak_upload_id = $1
		WHERE rak_id = ANY($2)
		  AND rak_coor_id = $3
		  AND rak_upload_id IS NULL
	`
	res, err := tx.Exec(ctx, SQL, uploadID, rakIDs, coorID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// SetSnapshot stores the immutable item snapshot for a claimed rack.
func (r *RackRepositoryImpl) SetSnapshot(ctx context.Context, tx pgx.Tx, rakID int, snapshot string) error {
	const SQL = `UPDATE rak SET rak_upload_snapshot = $2 WHERE rak_id = $1`
	_, err := tx.Exec(ctx, SQL, rakID, snapshot)
	return err
}

const selectRackState = `
	SELECT
		r.rak_id AS RakID,
		r.rak_name AS RakName,
		r.rak_coor_id AS CoorID,
		u.upload_createdat AS UploadedAt
	FROM rak r
	LEFT JOIN upload u ON u.upload_id = r.rak_upload_id
	WHERE r.rak_id = $1
`

// FindById implements [RackRepository].
func (r *RackRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error) {
	var state model.RackState
	err := tx.QueryRow(ctx, selectRackState, rakID).
		Scan(&state.RakID, &state.RakName, &state.CoorID, &state.UploadedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &model.NotFoundError{Resource: "Rack", ID: rakID}
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// FindByIdForUpdate implements [RackRepository].
func (r *RackRepositoryImpl) FindByIdForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error) {
	var state model.RackState
	err := tx.QueryRow(ctx, selectRackState+" FOR UPDATE OF r", rakID).
		Scan(&state.RakID, &state.RakName, &state.CoorID, &state.UploadedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &model.NotFoundError{Resource: "Rack", ID: rakID}
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// CountByCoordinator implements [RackRepository].
func (r *RackRepositoryImpl) CountByCoordinator(ctx context.Context, tx pgx.Tx, coorID int) (int, error) {
	const SQL = `SELECT COUNT(*) FROM rak WHERE rak_coor_id = $1`
	var total int
	if err := tx.QueryRow(ctx, SQL, coorID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// Delete implements [RackRepository].
func (r *RackRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, rakID int) error {
	const SQL = `DELETE FROM rak WHERE rak_id = $1`
	_, err := tx.Exec(ctx, SQL, rakID)
	return err
}
