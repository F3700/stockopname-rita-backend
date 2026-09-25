package dto

type ProductCreateRequest struct {
	PLU            string   `json:"plu" validate:"required"`
	Name           string   `json:"name" validate:"required"`
	DepartmentCode string   `json:"department_code" validate:"required"`
	BuyPrice       *float64 `json:"buy_price" validate:"required,gte=0"`
	SellPrice      *float64 `json:"sell_price" validate:"required,gte=0"`
	Barcodes       []string `json:"barcodes"`
}

type ProductUpdateRequest struct {
	PLU            *string  `json:"plu"`
	Name           *string  `json:"name"`
	DepartmentCode *string  `json:"department_code"`
	BuyPrice       *float64 `json:"buy_price" validate:"omitempty,gte=0"`
	SellPrice      *float64 `json:"sell_price" validate:"omitempty,gte=0"`
	// Barcodes replaces the whole barcode set when non-nil.
	// An empty (non-nil) slice removes all barcodes.
	Barcodes *[]string `json:"barcodes"`
}

type ProductResponse struct {
	Id             int      `json:"id"`
	PLU            string   `json:"plu"`
	Barcode        string   `json:"barcode"`
	Barcodes       []string `json:"barcodes"`
	Name           string   `json:"name"`
	BuyPrice       float64  `json:"buy_price"`
	SellPrice      float64  `json:"sell_price"`
	DateCreated    string   `json:"date_created"`
	DateUpdated    string   `json:"date_updated"`
	DepartmentCode string   `json:"department_code"`
}

type ProductLastSessionResponse struct {
	Id              int    `json:"id"`
	Barcode         string `json:"barcode"`
	Name            string `json:"name"`
	LastSessionDate string `json:"last_session_date"`
	LastSessionCode string `json:"last_session_code"`
}
