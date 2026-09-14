package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Sesi struct {
	SesiID        int                `json:"sesi_id"`
	SesiLocation  string             `json:"sesi_location"`
	SesiCode      string             `json:"sesi_code"`
	SesiStatus    string             `json:"sesi_status"`
	SesiStartedAt time.Time          `json:"sesi_startedat"`
	SesiEndedAt   pgtype.Timestamptz `json:"sesi_endedat"`
}
