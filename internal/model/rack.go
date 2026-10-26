package model

import "github.com/jackc/pgx/v5/pgtype"

type Rack struct {
	RackID   int    `json:"rak_id"`
	RackName string `json:"rak_name"`
	CoorID   int    `json:"rak_coor_id"`
}

// RackState is the row shape used by FindAll/FindById. Upload fields come from
// a LEFT JOIN so unuploaded and uploaded racks share one struct.
type RackState struct {
	RakID      int
	RakName    string
	CoorID     int
	UploadedAt pgtype.Timestamptz
}

// UploadItem is the immutable snapshot stored on the rack at submit time.
type UploadItem struct {
	PLU     string `json:"plu"`
	Barcode string `json:"barcode"`
	Name    string `json:"name"`
	Qty     int    `json:"qty"`
}

type RackProgress struct {
	RackAssigned  int `json:"rack_assigned"`
	RackCompleted int `json:"rack_completed"`
	TotalItems    int `json:"total_items"`
}
