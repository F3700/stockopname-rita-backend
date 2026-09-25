package dto

type StockOpnameResponse struct {
	Id              int    `json:"id"`
	Barcode         string `json:"barcode"`
	Name            string `json:"product_name"`
	Quantity        int    `json:"quantity"`
	RackName        string `json:"rak_name"`
	InspectorCode   string `json:"inspector_code"`
	CoordinatorCode string `json:"coor_code"`
	UpdatedAt       string `json:"updatedAt"`
}

// StockOpnameExportResponse is the denormalized row used for session file
// exports. It is separate from StockOpnameResponse so the JSON API
// responses stay unchanged.
type StockOpnameExportResponse struct {
	Id              int     `json:"id"`
	Barcode         string  `json:"barcode"`
	Name            string  `json:"product_name"`
	BuyPrice        float64 `json:"buy_price"`
	SellPrice       float64 `json:"sell_price"`
	Quantity        int     `json:"quantity"`
	RackName        string  `json:"rak_name"`
	InspectorCode   string  `json:"inspector_code"`
	CoordinatorCode string  `json:"coor_code"`
	UpdatedAt       string  `json:"updatedAt"`
}

// CreateStockOpnameRequest identifies the product by Barcode (scanned)
// or PLU (catalog pick). At least one of them is required.
type CreateStockOpnameRequest struct {
	Quantity int    `json:"quantity" validate:"required"`
	Barcode  string `json:"barcode"`
	PLU      string `json:"plu"`
	RakID    int    `json:"rak_id" validate:"required"`
}

type UpdateStockOpnameRequest struct {
	Quantity int `json:"quantity" validate:"required"`
}

type CreateStockOpnameByRackRequest struct {
	RackID int                        `json:"rak_id" validate:"required"`
	Items  []CreateStockOpnameRequest `json:"so_products" validate:"required"`
}
