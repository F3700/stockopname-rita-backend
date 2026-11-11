package dto

import "encoding/json"

// AuditResponse is one row from GET /stockopname/audits.
type AuditResponse struct {
	ID          int64           `json:"id"`
	BatchID     string          `json:"batch_id,omitempty"`
	Action      string          `json:"action"`
	SesiID      int             `json:"sesi_id"`
	SourceRakID *int            `json:"source_rak_id,omitempty"`
	TargetRakID *int            `json:"target_rak_id,omitempty"`
	Actor       string          `json:"actor"`
	Note        string          `json:"note"`
	Before      json.RawMessage `json:"before"`
	After       json.RawMessage `json:"after"`
	CreatedAt   string          `json:"created_at"`
}

// AuditBatchSummaryResponse is one batch group in the audit timeline.
type AuditBatchSummaryResponse struct {
	BatchID   string `json:"batch_id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Note      string `json:"note"`
	ItemCount int    `json:"item_count"`
	CreatedAt string `json:"created_at"`
}

// AuditTimelineItemResponse is one row of GET /stockopname/audits/timeline.
// kind is "single" for one change and "batch" for a grouped change. A batch
// that contains a single row is returned as "single".
type AuditTimelineItemResponse struct {
	Kind          string          `json:"kind"`
	ID            *int64          `json:"id,omitempty"`
	BatchID       string          `json:"batch_id,omitempty"`
	Action        string          `json:"action"`
	Actor         string          `json:"actor"`
	Note          string          `json:"note"`
	ItemCount     int             `json:"item_count,omitempty"`
	CreatedAt     string          `json:"created_at"`
	Before        json.RawMessage `json:"before,omitempty"`
	After         json.RawMessage `json:"after,omitempty"`
	FactorBefore  *int            `json:"factor_before,omitempty"`
	FactorAfter   *int            `json:"factor_after,omitempty"`
	SourceRakName string          `json:"source_rak_name,omitempty"`
	TargetRakName string          `json:"target_rak_name,omitempty"`
	MergedCount   int             `json:"merged_count,omitempty"`
}
