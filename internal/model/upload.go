package model

import "time"

type Upload struct {
	UploadID  int       `json:"upload_id"`
	CoorID    int       `json:"upload_coor_id"`
	CreatedAt time.Time `json:"upload_createdat"`
}

type UploadSummary struct {
	UploadID  int       `json:"upload_id"`
	CoorID    int       `json:"upload_coor_id"`
	RackCount int       `json:"rack_count"`
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"upload_createdat"`
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
