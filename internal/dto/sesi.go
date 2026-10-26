package dto

type SesiResponse struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	StartDate string `json:"startedat"`
	EndDate   string `json:"endedat"`
}

type CoordinatorInput struct {
	Code      string `json:"code" validate:"required"`
	RackStart int    `json:"rack_start" validate:"required,gte=1"`
	RackEnd   int    `json:"rack_end" validate:"required,gtefield=RackStart"`
}

type CreateSesiRequest struct {
	Code         string             `json:"sesi_code" validate:"required"`
	Location     string             `json:"location" validate:"required"`
	Coordinators []CoordinatorInput `json:"coordinators" validate:"required,min=1,dive"`
}

type UpdateSesiRequest struct {
	Status string `json:"status"`
}
