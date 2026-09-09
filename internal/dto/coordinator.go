package dto

type CoordinatorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	Inspector     int    `json:"inspector"`
	RackAssigned  int    `json:"rackAssigned"`
	RackCompleted int    `json:"rackCompleted"`
	Status        string `json:"status"`
}

type CreateCoordinatorRequest struct {
	Code   string `json:"code" validate:"required"`
	SesiId int    `json:"sesiId" validate:"required"`
}
