package repository

import (
	"context"
	"encoding/json"
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

// FindAllSummary implements [UploadRepository].
func (r *UploadRepositoryImpl) FindAllSummary(ctx context.Context, coordinatorId int) ([]*model.UploadSummary, error) {
	const SQL = `
		SELECT
			u.upload_id AS UploadID,
			ROW_NUMBER() OVER (ORDER BY u.upload_id) AS UploadNumber,
			u.upload_coor_id AS CoorID,
			COUNT(r.rak_id) AS RackCount,
			COALESCE(SUM(jsonb_array_length(r.rak_upload_snapshot)), 0) AS ItemCount,
			u.upload_createdat AS CreatedAt
		FROM upload u
		LEFT JOIN rak r ON r.rak_upload_id = u.upload_id
		WHERE u.upload_coor_id = $1
		GROUP BY u.upload_id, u.upload_coor_id, u.upload_createdat
		ORDER BY u.upload_createdat DESC, u.upload_id DESC;
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

// FindDetail implements [UploadRepository].
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
		SELECT rak_id, rak_name, rak_upload_snapshot
		FROM rak
		WHERE rak_upload_id = $1
		ORDER BY rak_name;
	`
	rows, err := r.Pool.Query(ctx, racksSQL, uploadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rack model.UploadRack
		var snapshot []byte
		if err := rows.Scan(&rack.RakID, &rack.RakName, &snapshot); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(snapshot, &rack.Items); err != nil {
			return nil, err
		}
		rack.ItemCount = len(rack.Items)
		detail.Racks = append(detail.Racks, rack)
		detail.RackCount++
		detail.ItemCount += rack.ItemCount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detail, nil
}
