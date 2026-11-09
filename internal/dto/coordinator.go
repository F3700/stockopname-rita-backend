package dto

type CoordinatorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	RackAssigned  int    `json:"rackAssigned"`
	RackCompleted int    `json:"rackCompleted"`
	Status        string `json:"status"`
}

// CoordinatorDetailResponse enriches CoordinatorResponse with parent session
// info for scan-QR preview. Only used by GET /stockopname/coordinators/:id.
type CoordinatorDetailResponse struct {
	ID              int    `json:"id"`
	Code            string `json:"code"`
	RackAssigned    int    `json:"rackAssigned"`
	RackCompleted   int    `json:"rackCompleted"`
	Status          string `json:"status"`
	SessionCode     string `json:"sessionCode"`
	SessionLocation string `json:"sessionLocation"`
}

type CreateCoordinatorRequest struct {
	SessionID int    `json:"session_id"`
	Code      string `json:"code"`
	RackStart int    `json:"rack_start"`
	RackEnd   int    `json:"rack_end"`
}

type UpdateCoordinatorRequest struct {
	Status string `json:"status" validate:"required"`
}
