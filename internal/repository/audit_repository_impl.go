package repository

import (
	"context"
	"encoding/json"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditRepositoryImpl struct {
	Pool DBPool
}

func NewAuditRepository(pool DBPool) AuditRepository {
	return &AuditRepositoryImpl{Pool: pool}
}

// Save implements [AuditRepository].
func (r *AuditRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, audit *model.StockOpnameAudit) error {
	const SQL = `
		INSERT INTO stock_opname_audit (
			batch_id,
			stock_opname_id,
			sesi_id,
			source_rak_id,
			target_rak_id,
			source_coor_id,
			target_coor_id,
			action,
			before_data,
			after_data,
			actor,
			note
		)
		VALUES (
			NULLIF($1, '')::uuid,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9::jsonb,
			$10::jsonb,
			$11,
			NULLIF($12, '')
		)
	`
	_, err := tx.Exec(ctx, SQL,
		audit.BatchID,
		audit.StockOpnameID,
		audit.SesiID,
		audit.SourceRakID,
		audit.TargetRakID,
		audit.SourceCoorID,
		audit.TargetCoorID,
		audit.Action,
		jsonArg(audit.BeforeData),
		jsonArg(audit.AfterData),
		audit.Actor,
		audit.Note,
	)
	return err
}

func jsonArg(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}

func int4Ptr(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	n := int(value.Int32)
	return &n
}

// FindAll implements [AuditRepository].
func (r *AuditRepositoryImpl) FindAll(ctx context.Context, filter AuditFilter) ([]*model.StockOpnameAudit, int, error) {
	const SQL = `
		SELECT
			a.audit_id,
			COALESCE(a.batch_id::text, '') AS batch_id,
			COALESCE(a.stock_opname_id, 0) AS stock_opname_id,
			a.sesi_id,
			a.source_rak_id,
			a.target_rak_id,
			a.source_coor_id,
			a.target_coor_id,
			a.action,
			COALESCE(a.before_data, 'null'::jsonb) AS before_data,
			COALESCE(a.after_data, 'null'::jsonb) AS after_data,
			a.actor,
			COALESCE(a.note, '') AS note,
			a.created_at
		FROM stock_opname_audit a
		WHERE ($1::int IS NULL OR a.sesi_id = $1::int)
		AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
		AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
		ORDER BY a.created_at DESC, a.audit_id DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.Pool.Query(ctx, SQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	audits := make([]*model.StockOpnameAudit, 0)
	for rows.Next() {
		var audit model.StockOpnameAudit
		var before, after json.RawMessage
		var sourceRakID, targetRakID, sourceCoorID, targetCoorID pgtype.Int4
		if err := rows.Scan(
			&audit.AuditID,
			&audit.BatchID,
			&audit.StockOpnameID,
			&audit.SesiID,
			&sourceRakID,
			&targetRakID,
			&sourceCoorID,
			&targetCoorID,
			&audit.Action,
			&before,
			&after,
			&audit.Actor,
			&audit.Note,
			&audit.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		audit.SourceRakID = int4Ptr(sourceRakID)
		audit.TargetRakID = int4Ptr(targetRakID)
		audit.SourceCoorID = int4Ptr(sourceCoorID)
		audit.TargetCoorID = int4Ptr(targetCoorID)
		audit.BeforeData = []byte(before)
		audit.AfterData = []byte(after)
		audits = append(audits, &audit)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const countSQL = `
		SELECT COUNT(*)
		FROM stock_opname_audit a
		WHERE ($1::int IS NULL OR a.sesi_id = $1::int)
		AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
		AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
	`
	var total int
	if err := r.Pool.QueryRow(ctx, countSQL, filter.SesiID, filter.CoordinatorID, filter.ResultID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return audits, total, nil
}
