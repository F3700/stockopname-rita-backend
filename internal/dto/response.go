package dto

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponsePagination struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}
