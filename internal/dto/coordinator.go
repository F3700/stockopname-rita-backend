package dto

type CoordinatorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	Inspector     int    `json:"inspector"`
	RackAssigned  int    `json:"rackAssigned"`
	RackCompleted int    `json:"rackCompleted"`
	Status        string `json:"status"`
}

// CoordinatorDetailResponse enriches CoordinatorResponse with parent session
// info for scan-QR preview. Only used by GET /stockopname/coordinators/:id.
type CoordinatorDetailResponse struct {
	ID              int    `json:"id"`
	Code            string `json:"code"`
	Inspector       int    `json:"inspector"`
	RackAssigned    int    `json:"rackAssigned"`
	RackCompleted   int    `json:"rackCompleted"`
	Status          string `json:"status"`
	SessionCode     string `json:"sessionCode"`
	SessionLocation string `json:"sessionLocation"`
}

type CreateCoordinatorRequest struct {
	Code   string `json:"code" validate:"required"`
	SesiId int    `json:"sesiId" validate:"required"`
}

type UpdateCoordinatorRequest struct {
	Status string `json:"status" validate:"required"`
}
