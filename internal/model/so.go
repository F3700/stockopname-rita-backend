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
	SoQtyFactor   int     `json:"qty_factor"`
}

// QtyFactorFilter selects result rows for a filtered bulk factor update.
type QtyFactorFilter struct {
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
	Id                int       `json:"id"`
	Barcode           string    `json:"barcode"`
	ProductPLU        string    `json:"product_plu"`
	Name              string    `json:"product_name"`
	Quantity          float64   `json:"quantity"`
	EffectiveQuantity float64   `json:"effective_quantity"`
	RackID            int       `json:"rak_id"`
	RackName          string    `json:"rak_name"`
	CoordinatorID     int       `json:"coordinator_id"`
	CoordinatorCode   string    `json:"coor_code"`
	QtyFactor         int       `json:"qty_factor"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// StockOpnameExport is the denormalized row used for session exports.
type StockOpnameExport struct {
	Id                int       `json:"id"`
	Barcode           string    `json:"barcode"`
	ProductPLU        string    `json:"product_plu"`
	Name              string    `json:"product_name"`
	DepartmentCode    string    `json:"department_code"`
	BuyPrice          float64   `json:"buy_price"`
	SellPrice         float64   `json:"sell_price"`
	QtyFactor         int       `json:"qty_factor"`
	PerCarton         int       `json:"per_carton"`
	Quantity          float64   `json:"quantity"`
	EffectiveQuantity float64   `json:"effective_quantity"`
	RackName          string    `json:"rak_name"`
	CoordinatorCode   string    `json:"coor_code"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
