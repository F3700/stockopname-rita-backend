package dto

type UploadItemRequest struct {
	Quantity *float64 `json:"quantity" validate:"required,gte=0"`
	Barcode  string   `json:"barcode"`
	PLU      string   `json:"plu"`
}

type UploadRackRequest struct {
	RakID int                 `json:"rak_id" validate:"required"`
	Items []UploadItemRequest `json:"so_products" validate:"required,min=1,dive"`
}

type CreateUploadRequest struct {
	CoordinatorID int                 `json:"coordinator_id" validate:"required"`
	Racks         []UploadRackRequest `json:"racks" validate:"required,min=1,dive"`
}

type UploadResponse struct {
	UploadID     int    `json:"upload_id"`
	UploadNumber int    `json:"upload_number,omitempty"`
	RackCount    int    `json:"rack_count"`
	ItemCount    int    `json:"item_count"`
	CreatedAt    string `json:"created_at"`
}

type UploadDetailRackResponse struct {
	RakID     int                  `json:"rak_id"`
	RakName   string               `json:"rak_name"`
	ItemCount int                  `json:"item_count"`
	Items     []UploadItemResponse `json:"items"`
}

type UploadDetailResponse struct {
	UploadResponse
	CoordinatorCode string                     `json:"coor_code"`
	SessionCode     string                     `json:"sesi_code"`
	SessionLocation string                     `json:"sesi_location"`
	Racks           []UploadDetailRackResponse `json:"racks"`
}

type UploadItemResponse struct {
	PLU       string  `json:"plu"`
	Barcode   string  `json:"barcode"`
	Name      string  `json:"name"`
	Qty       float64 `json:"qty"`
	BuyPrice  float64 `json:"buy_price"`
	SellPrice float64 `json:"sell_price"`
}
