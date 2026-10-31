package model

import "time"

type Upload struct {
	UploadID  int       `json:"upload_id"`
	CoorID    int       `json:"upload_coor_id"`
	CreatedAt time.Time `json:"upload_createdat"`
}

type UploadSummary struct {
	UploadID     int       `json:"upload_id"`
	UploadNumber int       `json:"upload_number"`
	CoorID       int       `json:"upload_coor_id"`
	RackCount    int       `json:"rack_count"`
	ItemCount    int       `json:"item_count"`
	CreatedAt    time.Time `json:"upload_createdat"`
}

type UploadRack struct {
	RakID     int          `json:"rak_id"`
	RakName   string       `json:"rak_name"`
	ItemCount int          `json:"item_count"`
	Items     []UploadItem `json:"items"`
}

type UploadDetail struct {
	UploadSummary
	CoordinatorCode string       `json:"coor_code"`
	SessionCode     string       `json:"sesi_code"`
	SessionLocation string       `json:"sesi_location"`
	Racks           []UploadRack `json:"racks"`
}

// UploadReportRow is one upload line in the per-coordinator summary PDF.
type UploadReportRow struct {
	CoordinatorCode string
	UploadID        int
	UploadNumber    int
	Qty             float64
	TotalBuy        float64
	TotalSell       float64
}
