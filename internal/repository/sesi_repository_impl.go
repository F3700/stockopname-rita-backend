package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type SesiRepositoryImpl struct {
}

func NewSesiRepository() SesiRepository {
	return &SesiRepositoryImpl{}
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

func (s *SesiRepositoryImpl) FindAllInPageSearch(ctx context.Context, tx pgx.Tx, limit int, offset int, search string) ([]*model.Sesi, int, error) {
	const SQL = `
		SELECT sesi_id, sesi_location, sesi_code, sesi_status, sesi_startedat, sesi_endedat
		FROM sesi
		WHERE sesi_location ILIKE $3 OR sesi_code ILIKE $3
		ORDER BY sesi_id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := tx.Query(ctx, SQL, limit, offset, "%"+search+"%")
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sesiModels, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.Sesi])
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err = tx.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM sesi
        WHERE sesi_location ILIKE $1 OR sesi_code ILIKE $1
    `, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return sesiModels, total, nil
}

// FindById implements [SesiRepository].
func (s *SesiRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error) {
	const SQL = `
		SELECT sesi_id, sesi_location, sesi_code, sesi_status, sesi_startedat, sesi_endedat
		FROM sesi
		WHERE sesi_id = $1
	`
	var sesi model.Sesi
	err := tx.QueryRow(ctx, SQL, id).Scan(&sesi.SesiID, &sesi.SesiLocation, &sesi.SesiCode, &sesi.SesiStatus, &sesi.SesiStartedAt, &sesi.SesiEndedAt)

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
	const SQL = `
		SELECT update_sesi_status($1, $2)
	`
	result, err := tx.Exec(ctx, SQL, sesi.SesiID, sesi.SesiStatus)
	if err != nil {
		return model.MapPgError("Session", err)
	}
	if result.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Session", ID: sesi.SesiID}
	}
	return nil
}
