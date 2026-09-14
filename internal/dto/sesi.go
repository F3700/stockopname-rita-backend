package dto

type SesiResponse struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	StartDate string `json:"startedat"`
	EndDate   string `json:"endedat"`
}

type CreateSesiRequest struct {
	Code        string   `json:"sesi_code" validate:"required"`
	Location    string   `json:"location" validate:"required"`
	Coordinator []string `json:"coor_code" validate:"required"`
}

type UpdateSesiRequest struct {
	Status string `json:"status"`
}
