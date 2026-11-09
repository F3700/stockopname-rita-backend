package dto

type ProductCreateRequest struct {
	SesiID         int      `json:"sesi_id" validate:"required,gte=1"`
	PLU            string   `json:"plu" validate:"required,max=10"`
	Name           string   `json:"name" validate:"required,max=150"`
	DepartmentCode string   `json:"department_code" validate:"required,max=10"`
	BuyPrice       *float64 `json:"buy_price" validate:"required,gte=0"`
	SellPrice      *float64 `json:"sell_price" validate:"required,gte=0"`
	Barcodes       []string `json:"barcodes" validate:"omitempty,dive,max=15"`
}

type ProductUpdateRequest struct {
	PLU            *string  `json:"plu" validate:"omitempty,max=10"`
	Name           *string  `json:"name" validate:"omitempty,max=150"`
	DepartmentCode *string  `json:"department_code" validate:"omitempty,max=10"`
	BuyPrice       *float64 `json:"buy_price" validate:"omitempty,gte=0"`
	SellPrice      *float64 `json:"sell_price" validate:"omitempty,gte=0"`
	// Barcodes replaces the whole barcode set when non-nil.
	// An empty (non-nil) slice removes all barcodes.
	Barcodes *[]string `json:"barcodes" validate:"omitempty,dive,max=15"`
}

type ProductResponse struct {
	Id             int      `json:"id"`
	PLU            string   `json:"plu"`
	Barcodes       []string `json:"barcodes"`
	Name           string   `json:"name"`
	BuyPrice       float64  `json:"buy_price"`
	SellPrice      float64  `json:"sell_price"`
	DateCreated    string   `json:"date_created"`
	DateUpdated    string   `json:"date_updated"`
	DepartmentCode string   `json:"department_code"`
}
