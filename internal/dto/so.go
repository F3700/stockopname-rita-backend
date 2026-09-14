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

type CreateStockOpnameRequest struct {
	Quantity  int `json:"quantity" validate:"required"`
	ProductID int `json:"product_id" validate:"required"`
	RakID     int `json:"rak_id" validate:"required"`
}

type UpdateStockOpnameRequest struct {
	Quantity int `json:"quantity" validate:"required"`
}

type CreateStockOpnameByRackRequest struct {
	RackID int                        `json:"rak_id" validate:"required"`
	Items  []CreateStockOpnameRequest `json:"so_products" validate:"required"`
}
