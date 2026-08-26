package model

type Category struct {
	CategoryID          int    `json:"category_id"`
	CategoryName        string `json:"category_name"`
	CategoryDescription string `json:"category_desc"`
}
