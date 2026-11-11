package model

import "github.com/jackc/pgx/v5/pgtype"

type Rack struct {
	RackID   int    `json:"rak_id"`
	RackName string `json:"rak_name"`
	CoorID   int    `json:"rak_coor_id"`
}

// RackState is the row shape used by FindAll/FindById. Upload fields come from
// a LEFT JOIN so unuploaded and uploaded racks share one struct. UploadedAt is
// the host upload: the upload that last claimed the rack. CanUpload mirrors the
// claim guard (rak_upload_id IS NULL); ItemCount counts every row in the rack
// while UnownedCount counts only rows without an owner.
type RackState struct {
	RakID        int
	RakName      string
	CoorID       int
	UploadedAt   pgtype.Timestamptz
	ItemCount    int
	UnownedCount int
	CanUpload    bool
}

// UploadItem is the item shape written to rak_upload_snapshot at submit time.
// The snapshot is an archive only: upload reports read the current
// stock_opname rows, and the next upload of the rack overwrites it.
type UploadItem struct {
	PLU       string  `json:"plu"`
	Barcode   string  `json:"barcode"`
	Name      string  `json:"name"`
	Dept      string  `json:"dept"`
	Qty       float64 `json:"qty"`
	BuyPrice  float64 `json:"buy_price"`
	SellPrice float64 `json:"sell_price"`
	PerCarton int     `json:"per_carton,omitempty"`
}

type RackProgress struct {
	RackAssigned  int `json:"rack_assigned"`
	RackCompleted int `json:"rack_completed"`
	TotalItems    int `json:"total_items"`
}
