package model

import "time"

type StockOpname struct {
	StockOpnameID        int       `json:"stockopname_id"`
	StockOpnameQuantity  float64   `json:"stockopname_quantity"`
	StockOpnameUpdatedAt time.Time `json:"stockopname_updatedat"`
	StockOpnameRakID     int       `json:"stockopname_rak_id"`

	// Product snapshot: plain columns, NOT a foreign key to product,
	// so history stays readable after the master changes or is cleared.
	SoProductPLU  string  `json:"product_plu"`
	SoProductName string  `json:"product_name"`
	SoProductDept string  `json:"department_code"`
	SoBarcode     string  `json:"barcode"`
	SoBuyPrice    float64 `json:"buy_price"`
	SoSellPrice   float64 `json:"sell_price"`
	SoSellFactor  int     `json:"sell_factor"`
}

// SellFactorFilter selects result rows for a filtered bulk factor update.
type SellFactorFilter struct {
	Search        string
	CoordinatorID *int
	RakID         *int
}

// SellSummaryGroup is one grouping row for the sell summary report.
type SellSummaryGroup struct {
	Key                string
	Label              string
	ItemCount          int
	TotalQty           float64
	TotalSellOriginal  float64
	TotalSellEffective float64
}

type StockOpnameSummary struct {
	Id              int       `json:"id"`
	Barcode         string    `json:"barcode"`
	ProductPLU      string    `json:"product_plu"`
	Name            string    `json:"product_name"`
	Quantity        float64   `json:"quantity"`
	RackID          int       `json:"rak_id"`
	RackName        string    `json:"rak_name"`
	CoordinatorID   int       `json:"coordinator_id"`
	CoordinatorCode string    `json:"coor_code"`
	SellFactor      int       `json:"sell_factor"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// StockOpnameExport is the denormalized row used for session exports.
// It intentionally mirrors FindAll plus product prices and is separate from
// StockOpnameSummary so the JSON API responses stay unchanged.
type StockOpnameExport struct {
	Id              int       `json:"id"`
	Barcode         string    `json:"barcode"`
	ProductPLU      string    `json:"product_plu"`
	Name            string    `json:"product_name"`
	DepartmentCode  string    `json:"department_code"`
	BuyPrice        float64   `json:"buy_price"`
	SellPrice       float64   `json:"sell_price"`
	SellFactor      int       `json:"sell_factor"`
	PerCarton       int       `json:"per_carton"`
	Quantity        float64   `json:"quantity"`
	RackName        string    `json:"rak_name"`
	CoordinatorCode string    `json:"coor_code"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
