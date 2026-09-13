package model

import "time"

type StockOpname struct {
	StockOpnameID        int       `json:"stockopname_id"`
	StockOpnameQuantity  int       `json:"stockopname_quantity"`
	StockOpnameUpdatedAt time.Time `json:"stockopname_updatedat"`
	StockOpnameProductID int       `json:"stockopname_product_id"`
	StockOpnameRakID     int       `json:"stockopname_rak_id"`
}

type StockOpnameSummary struct {
	Id              int       `json:"id"`
	Barcode         string    `json:"barcode"`
	Name            string    `json:"product_name"`
	Quantity        int       `json:"quantity"`
	RackName        string    `json:"rak_name"`
	InspectorCode   string    `json:"inspector_code"`
	CoordinatorCode string    `json:"coor_code"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ProductLastSession struct {
	Id              int        `json:"id"`
	Barcode         string     `json:"barcode"`
	Name            string     `json:"name"`
	LastSessionDate *time.Time `json:"last_session_date"`
	LastSessionCode string     `json:"last_session_code"`
}
