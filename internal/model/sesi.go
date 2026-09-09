package model

import "time"

type Sesi struct {
	SesiID        string    `json:"sesi_id"`
	SesiLocation  string    `json:"sesi_location"`
	SesiCode      string    `json:"sesi_code"`
	SesiStatus    string    `json:"sesi_status"`
	SesiStartedAt time.Time `json:"sesi_startedat"`
	SesiEndedAt   time.Time `json:"sesi_endedat"`
}
