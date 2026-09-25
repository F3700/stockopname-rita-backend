package model

import "time"

type DeletedProduct struct {
	ProductPLU string    `json:"product_plu"`
	DeletedAt  time.Time `json:"deleted_at"`
}
