package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SesiRepositoryImpl struct {
	Pool *pgxpool.Pool
}

func NewSesiRepository(pool *pgxpool.Pool) SesiRepository {
	return &SesiRepositoryImpl{
		Pool: pool,
	}
}

// Delete implements [SesiRepository].
func (s *SesiRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `
		DELETE FROM sesi
		WHERE sesi_id = $1
	`
	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Session", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Session", ID: id}
	}
	return nil
}

func (s *SesiRepositoryImpl) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error) {
	const SQL = `
		SELECT sesi_id, sesi_location, sesi_code, sesi_status, sesi_startedat, sesi_endedat
		FROM sesi
		WHERE sesi_location ILIKE $3 OR sesi_code ILIKE $3
		ORDER BY sesi_id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.Pool.Query(ctx, SQL, limit, offset, "%"+search+"%")
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sesiModels, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.Sesi])
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err = s.Pool.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM sesi
        WHERE sesi_location ILIKE $1 OR sesi_code ILIKE $1
    `, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return sesiModels, total, nil
}

// FindById implements [SesiRepository].
func (s *SesiRepositoryImpl) FindById(ctx context.Context, id int) (*model.Sesi, error) {
	const SQL = `
		SELECT sesi_id, sesi_location, sesi_code, sesi_status, sesi_startedat, sesi_endedat
		FROM sesi
		WHERE sesi_id = $1
	`
	var sesi model.Sesi
	err := s.Pool.QueryRow(ctx, SQL, id).Scan(&sesi.SesiID, &sesi.SesiLocation, &sesi.SesiCode, &sesi.SesiStatus, &sesi.SesiStartedAt, &sesi.SesiEndedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Session", ID: id}
		}
		return nil, err
	}
	return &sesi, nil
}

// Save implements [SesiRepository].
func (s *SesiRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error {
	const SQL = `
		INSERT INTO sesi (sesi_location, sesi_code, sesi_status)
		VALUES ($1, $2, $3)
		RETURNING sesi_id, sesi_startedat, sesi_endedat
	`
	row := tx.QueryRow(ctx, SQL, sesi.SesiLocation, sesi.SesiCode, sesi.SesiStatus)

	err := row.Scan(&sesi.SesiID, &sesi.SesiStartedAt, &sesi.SesiEndedAt)
	if err != nil {
		return model.MapPgError("Session", err)
	}

	return nil
}

// Update implements [SesiRepository].
func (s *SesiRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sesi WHERE sesi_id = $1)`, sesi.SesiID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return &model.NotFoundError{Resource: "Session", ID: sesi.SesiID}
	}

	const SQL = `
		SELECT update_sesi_status($1, $2)
	`
	if _, err := tx.Exec(ctx, SQL, sesi.SesiID, sesi.SesiStatus); err != nil {
		return model.MapPgError("Session", err)
	}
	return nil
}
