package model

import "time"

type Product struct {
	ProductID             int       `json:"product_id"`
	ProductPLU            string    `json:"product_plu"`
	ProductName           string    `json:"product_name"`
	ProductDepartmentCode string    `json:"department_code"`
	ProductBuyPrice       float64   `json:"product_buyprice"`
	ProductSellPrice      float64   `json:"product_sellprice"`
	ProductCreatedat      time.Time `json:"product_createdat"`
	ProductUpdatedat      time.Time `json:"product_updatedat"`

	ProductBarcodes []string `json:"barcodes"`
}

// PrimaryBarcode returns the first barcode (insertion order) or "".
func (p *Product) PrimaryBarcode() string {
	if len(p.ProductBarcodes) == 0 {
		return ""
	}
	return p.ProductBarcodes[0]
}
