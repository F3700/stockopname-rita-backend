package dto

type StockOpnameResponse struct {
	Id              int     `json:"id"`
	Barcode         string  `json:"barcode"`
	PLU             string  `json:"plu"`
	Name            string  `json:"product_name"`
	Quantity        float64 `json:"quantity"`
	RackID          int     `json:"rak_id"`
	RackName        string  `json:"rak_name"`
	CoordinatorID   int     `json:"coordinator_id"`
	CoordinatorCode string  `json:"coor_code"`
	UpdatedAt       string  `json:"updatedAt"`
}

// StockOpnameExportResponse is the denormalized row used for session file
// exports. It is separate from StockOpnameResponse so the JSON API
// responses stay unchanged.
type StockOpnameExportResponse struct {
	Id              int     `json:"id"`
	Barcode         string  `json:"barcode"`
	PLU             string  `json:"plu"`
	Name            string  `json:"product_name"`
	DepartmentCode  string  `json:"department_code"`
	BuyPrice        float64 `json:"buy_price"`
	SellPrice       float64 `json:"sell_price"`
	Quantity        float64 `json:"quantity"`
	RackName        string  `json:"rak_name"`
	CoordinatorCode string  `json:"coor_code"`
	UpdatedAt       string  `json:"updatedAt"`
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
	MovedCount  int `json:"moved_count"`
	MergedCount int `json:"merged_count"`
}
