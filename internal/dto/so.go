package dto

type StockOpnameResponse struct {
	Id                int     `json:"id"`
	Barcode           string  `json:"barcode"`
	PLU               string  `json:"plu"`
	Name              string  `json:"product_name"`
	Quantity          float64 `json:"quantity"`
	EffectiveQuantity float64 `json:"effective_quantity"`
	RackID            int     `json:"rak_id"`
	RackName          string  `json:"rak_name"`
	CoordinatorID     int     `json:"coordinator_id"`
	CoordinatorCode   string  `json:"coor_code"`
	QtyFactor         int     `json:"qty_factor"`
	UpdatedAt         string  `json:"updatedAt"`
}

// StockOpnameExportResponse is the denormalized row used for session file
// exports.
type StockOpnameExportResponse struct {
	Id                int     `json:"id"`
	Barcode           string  `json:"barcode"`
	PLU               string  `json:"plu"`
	Name              string  `json:"product_name"`
	DepartmentCode    string  `json:"department_code"`
	BuyPrice          float64 `json:"buy_price"`
	SellPrice         float64 `json:"sell_price"`
	QtyFactor         int     `json:"qty_factor"`
	PerCarton         int     `json:"per_carton"`
	Quantity          float64 `json:"quantity"`
	EffectiveQuantity float64 `json:"effective_quantity"`
	RackName          string  `json:"rak_name"`
	CoordinatorCode   string  `json:"coor_code"`
	UpdatedAt         string  `json:"updatedAt"`
}

// CreateStockOpnameRequest identifies the product by Barcode (scanned)
// or PLU (catalog pick). At least one of them is required.
type CreateStockOpnameRequest struct {
	// Pointer so a missing quantity is rejected while a real zero count is allowed.
	Quantity *float64 `json:"quantity" validate:"required,gte=0"`
	Barcode  string   `json:"barcode"`
	PLU      string   `json:"plu"`
	RakID    int      `json:"rak_id" validate:"required"`
	Note     string   `json:"note"`
	Actor    string   `json:"-"`
}

type UpdateStockOpnameRequest struct {
	Quantity *float64 `json:"quantity" validate:"required,gte=0"`
	Note     string   `json:"note"`
	Actor    string   `json:"-"`
}

// MoveStockOpnameRequest moves one rack or a subset of one rack.
// ResultIDs nil means "all results in the source rack".
type MoveStockOpnameRequest struct {
	SourceRakID int    `json:"source_rak_id" validate:"required"`
	TargetRakID int    `json:"target_rak_id" validate:"required"`
	ResultIDs   *[]int `json:"result_ids"`
	Note        string `json:"note"`
	Actor       string `json:"-"`
}

type MoveStockOpnameResponse struct {
	MovedCount     int  `json:"moved_count"`
	MergedCount    int  `json:"merged_count"`
	ReopenedRackID *int `json:"reopened_rack_id,omitempty"`
}

type QtyFactorFilterRequest struct {
	Search        string `json:"search"`
	CoordinatorID *int   `json:"coordinator_id"`
	RakID         *int   `json:"rak_id"`
}

type SetQtyFactorRequest struct {
	Scope     string                  `json:"scope" validate:"required,oneof=session coordinator rack item items filter"`
	ScopeID   *int                    `json:"scope_id"`
	ResultIDs []int                   `json:"result_ids"`
	Filter    *QtyFactorFilterRequest `json:"filter"`
	QtyFactor *int                    `json:"qty_factor" validate:"required,oneof=-1 0 1"`
	Note      string                  `json:"note"`
	Actor     string                  `json:"-"`
}

type SetQtyFactorResponse struct {
	UpdatedCount int    `json:"updated_count"`
	BatchID      string `json:"batch_id"`
}

type SellSummaryResponse struct {
	GroupBy string             `json:"group_by"`
	Groups  []SellSummaryGroup `json:"groups"`
}

type SellSummaryGroup struct {
	Key                string  `json:"key"`
	Label              string  `json:"label"`
	ItemCount          int     `json:"item_count"`
	TotalQty           float64 `json:"total_qty"`
	TotalSellOriginal  float64 `json:"total_sell_original"`
	TotalSellEffective float64 `json:"total_sell_effective"`
}
