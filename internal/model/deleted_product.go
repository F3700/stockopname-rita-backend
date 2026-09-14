package model

import "time"

type DeletedProduct struct {
	ProductID int       `json:"product_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
