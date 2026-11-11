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
		AND (NULLIF($4::text, '') IS NULL OR a.batch_id = NULLIF($4::text, '')::uuid)
		AND (NOT $5::boolean OR a.batch_id IS NULL)
		ORDER BY a.created_at DESC, a.audit_id DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.Pool.Query(ctx, SQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.BatchID, filter.WithoutBatch, filter.Limit, filter.Offset)
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
		AND (NULLIF($4::text, '') IS NULL OR a.batch_id = NULLIF($4::text, '')::uuid)
		AND (NOT $5::boolean OR a.batch_id IS NULL)
	`
	var total int
	if err := r.Pool.QueryRow(ctx, countSQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.BatchID, filter.WithoutBatch).Scan(&total); err != nil {
		return nil, 0, err
	}

	return audits, total, nil
}

func (r *AuditRepositoryImpl) FindBatches(ctx context.Context, filter AuditFilter) ([]*model.StockOpnameAuditBatch, int, error) {
	const SQL = `
		SELECT
			a.batch_id::text AS batch_id,
			MIN(a.action) AS action,
			MIN(a.actor) AS actor,
			COALESCE(MIN(a.note), '') AS note,
			COUNT(*)::int AS item_count,
			MIN(a.created_at) AS created_at
		FROM stock_opname_audit a
		WHERE a.batch_id IS NOT NULL
		AND ($1::int IS NULL OR a.sesi_id = $1::int)
		AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
		AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
		GROUP BY a.batch_id
		ORDER BY MIN(a.created_at) DESC, a.batch_id
		LIMIT $4 OFFSET $5
	`
	rows, err := r.Pool.Query(ctx, SQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	batches := make([]*model.StockOpnameAuditBatch, 0)
	for rows.Next() {
		var batch model.StockOpnameAuditBatch
		if err := rows.Scan(
			&batch.BatchID,
			&batch.Action,
			&batch.Actor,
			&batch.Note,
			&batch.ItemCount,
			&batch.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		batches = append(batches, &batch)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const countSQL = `
		SELECT COUNT(DISTINCT a.batch_id)
		FROM stock_opname_audit a
		WHERE a.batch_id IS NOT NULL
		AND ($1::int IS NULL OR a.sesi_id = $1::int)
		AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
		AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
	`
	var total int
	if err := r.Pool.QueryRow(ctx, countSQL, filter.SesiID, filter.CoordinatorID, filter.ResultID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return batches, total, nil
}

// FindTimeline implements [AuditRepository]. It returns batches and single
// changes in one chronological list. A batch that holds a single row is
// returned as a single entry, so one-item work is never shown as a bulk action.
func (r *AuditRepositoryImpl) FindTimeline(ctx context.Context, filter AuditFilter) ([]*model.AuditTimelineItem, int, error) {
	const SQL = `
		WITH filtered AS (
			SELECT a.*
			FROM stock_opname_audit a
			WHERE ($1::int IS NULL OR a.sesi_id = $1::int)
			AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
			AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
		),
		singles AS (
			SELECT
				'single'::text AS kind,
				a.audit_id,
				NULL::uuid AS batch_id,
				a.action,
				a.actor,
				COALESCE(a.note, '') AS note,
				a.created_at,
				1 AS item_count,
				a.before_data,
				a.after_data
			FROM filtered a
			WHERE a.batch_id IS NULL
			AND ($4::text[] IS NULL OR a.action = ANY($4::text[]))
		),
		batched AS (
			SELECT
				CASE WHEN COUNT(*) = 1 THEN 'single' ELSE 'batch' END AS kind,
				CASE WHEN COUNT(*) = 1 THEN MIN(a.audit_id) ELSE NULL END AS audit_id,
				a.batch_id,
				MIN(a.action) AS action,
				MIN(a.actor) AS actor,
				COALESCE(MIN(a.note), '') AS note,
				MIN(a.created_at) AS created_at,
				COUNT(*)::int AS item_count,
				CASE WHEN COUNT(*) = 1 THEN (array_agg(a.before_data))[1] ELSE NULL END AS before_data,
				CASE WHEN COUNT(*) = 1 THEN (array_agg(a.after_data))[1] ELSE NULL END AS after_data
			FROM filtered a
			WHERE a.batch_id IS NOT NULL
			GROUP BY a.batch_id
			HAVING ($4::text[] IS NULL OR MIN(a.action) = ANY($4::text[]))
		)
		SELECT
			t.kind,
			COALESCE(t.audit_id, 0) AS audit_id,
			COALESCE(t.batch_id::text, '') AS batch_id,
			t.action,
			t.actor,
			t.note,
			t.created_at,
			t.item_count,
			COALESCE(t.before_data, 'null'::jsonb) AS before_data,
			COALESCE(t.after_data, 'null'::jsonb) AS after_data
		FROM (SELECT * FROM singles UNION ALL SELECT * FROM batched) t
		ORDER BY t.created_at DESC, COALESCE(t.audit_id, 0) DESC, t.batch_id DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.Pool.Query(ctx, SQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.Actions, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*model.AuditTimelineItem, 0)
	for rows.Next() {
		var item model.AuditTimelineItem
		var auditID int64
		var before, after json.RawMessage
		if err := rows.Scan(
			&item.Kind,
			&auditID,
			&item.BatchID,
			&item.Action,
			&item.Actor,
			&item.Note,
			&item.CreatedAt,
			&item.ItemCount,
			&before,
			&after,
		); err != nil {
			return nil, 0, err
		}
		if item.Kind == "single" {
			id := auditID
			item.AuditID = &id
		}
		item.BeforeData = []byte(before)
		item.AfterData = []byte(after)
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const countSQL = `
		SELECT
			(SELECT COUNT(*)
				FROM stock_opname_audit a
				WHERE a.batch_id IS NULL
				AND ($1::int IS NULL OR a.sesi_id = $1::int)
				AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
				AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
				AND ($4::text[] IS NULL OR a.action = ANY($4::text[])))
			+
			(SELECT COUNT(*)
				FROM (
					SELECT 1
					FROM stock_opname_audit a
					WHERE a.batch_id IS NOT NULL
					AND ($1::int IS NULL OR a.sesi_id = $1::int)
					AND ($2::int IS NULL OR a.source_coor_id = $2::int OR a.target_coor_id = $2::int)
					AND ($3::int IS NULL OR a.stock_opname_id = $3::int)
					GROUP BY a.batch_id
					HAVING ($4::text[] IS NULL OR MIN(a.action) = ANY($4::text[]))
				) b)
	`
	var total int
	if err := r.Pool.QueryRow(ctx, countSQL, filter.SesiID, filter.CoordinatorID, filter.ResultID, filter.Actions).Scan(&total); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
