package model

import "time"

type Product struct {
	ProductID             int       `json:"product_id"`
	ProductSesiID         int       `json:"product_sesi_id"`
	ProductPLU            string    `json:"product_plu"`
	ProductName           string    `json:"product_name"`
	ProductDepartmentCode string    `json:"department_code"`
	ProductBuyPrice       float64   `json:"product_buyprice"`
	ProductSellPrice      float64   `json:"product_sellprice"`
	ProductCreatedat      time.Time `json:"product_createdat"`
	ProductUpdatedat      time.Time `json:"product_updatedat"`

	ProductBarcodes []string `json:"barcodes"`
}

// FirstBarcode returns the first barcode (insertion order) or "".
// It is only a deterministic fallback for snapshots when a product is
// selected by PLU; barcodes are otherwise a set with equal standing.
func (p *Product) FirstBarcode() string {
	if len(p.ProductBarcodes) == 0 {
		return ""
	}
	return p.ProductBarcodes[0]
}
