package repository

import (
	"context"
	"fmt"
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
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("sesi with id %d not found", id)
	}
	return err
}

// FindAll implements [SesiRepository].
func (s *SesiRepositoryImpl) FindAll(ctx context.Context, tx pgx.Tx) ([]*model.Sesi, error) {
	panic("unimplemented")
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
		return nil, fmt.Errorf("failed to find sesi with id %d: %w", id, err)
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
		return err
	}

	return nil
}

// Update implements [SesiRepository].
func (s *SesiRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error {
	const SQL = `
		UPDATE sesi
		SET sesi_status = $1
		WHERE sesi_id = $2
	`
	result, err := tx.Exec(ctx, SQL, sesi.SesiStatus, sesi.SesiID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("sesi with id %d not found", sesi.SesiID)
	}
	return nil
}
