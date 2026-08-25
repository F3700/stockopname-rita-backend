package dto

type ProductCreateRequest struct {
	Barcode      string  `json:"barcode" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	BuyPrice     float64 `json:"buy_price" validate:"required"`
	SellPrice    float64 `json:"sell_price" validate:"required"`
	CategoryID   int     `json:"category_id" validate:"required"`
	DepartmentID int     `json:"department_id" validate:"required"`
}

type ProductUpdateRequest struct {
	Barcode      *string  `json:"barcode"`
	Name         *string  `json:"name"`
	BuyPrice     *float64 `json:"buy_price"`
	SellPrice    *float64 `json:"sell_price"`
	CategoryID   *int     `json:"category_id"`
	DepartmentID *int     `json:"department_id"`
}

type ProductResponse struct {
	Id             int     `json:"id"`
	Barcode        string  `json:"barcode"`
	Name           string  `json:"name"`
	BuyPrice       float64 `json:"buy_price"`
	SellPrice      float64 `json:"sell_price"`
	DateCreated    string  `json:"date_created"`
	DateUpdated    string  `json:"date_updated"`
	CategoryName   string  `json:"category_name"`
	DepartmentCode string  `json:"department_code"`
}
