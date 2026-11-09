package model

import "time"

// StockOpnameAudit records one administrative change to a result row.
type StockOpnameAudit struct {
	AuditID       int64
	BatchID       string
	StockOpnameID int
	SesiID        int
	SourceRakID   *int
	TargetRakID   *int
	SourceCoorID  *int
	TargetCoorID  *int
	Action        string
	BeforeData    []byte
	AfterData     []byte
	Actor         string
	Note          string
	CreatedAt     time.Time
}

// StockOpnameAuditData is the JSON payload stored in before_data/after_data.
type StockOpnameAuditData struct {
	StockOpnameID     int        `json:"stock_opname_id"`
	RakID             int        `json:"rak_id"`
	RakName           string     `json:"rak_name"`
	CoordinatorID     int        `json:"coordinator_id"`
	CoordinatorCode   string     `json:"coordinator_code"`
	CoordinatorStatus string     `json:"coordinator_status,omitempty"`
	RakUploadID       *int       `json:"rak_upload_id,omitempty"`
	RakUploadedAt     *time.Time `json:"rak_uploaded_at,omitempty"`
	PLU               string     `json:"plu"`
	Barcode           string     `json:"barcode"`
	Name              string     `json:"name"`
	DepartmentCode    string     `json:"department_code"`
	Quantity          float64    `json:"quantity"`
	BuyPrice          float64    `json:"buy_price"`
	SellPrice         float64    `json:"sell_price"`
	SellFactor        int        `json:"sell_factor"`
}

// StockOpnameAuditBatch is one row of the audit batch summary.
type StockOpnameAuditBatch struct {
	BatchID   string
	Action    string
	Actor     string
	Note      string
	ItemCount int
	CreatedAt time.Time
}

// StockOpnameRackContext describes the rack and coordinator that own a result.
type StockOpnameRackContext struct {
	RakID      int
	RakName    string
	CoorID     int
	CoorCode   string
	CoorStatus string
	SesiID     int
	UploadID   *int
	UploadedAt *time.Time
}
