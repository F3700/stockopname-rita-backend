package model

import "time"

type Barcode struct {
	BarcodeID        int       `json:"barcode_id"`
	BarcodeSesiID    int       `json:"barcode_sesi_id"`
	BarcodeProductID int       `json:"barcode_product_id"`
	BarcodeCode      string    `json:"barcode_code"`
	BarcodeCreatedat time.Time `json:"barcode_createdat"`
}
