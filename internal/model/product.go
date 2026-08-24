package model

import "time"

type Product struct {
	ProductID           int       `json:"product_id"`
	ProductBarcode      string    `json:"product_barcode"`
	ProductName         string    `json:"product_name"`
	ProductBuyPrice     float64   `json:"product_buyprice"`
	ProductSellPrice    float64   `json:"product_sellprice"`
	ProductCreatedat    time.Time `json:"product_createdat"`
	ProductUpdatedat    time.Time `json:"product_updatedat"`
	ProductCategoryID   int       `json:"product_category_id"`
	ProductDepartmentID int       `json:"product_department_id"`

	ProductCategoryName   string `json:"category_name"`
	ProductDepartmentCode string `json:"department_code"`
}
