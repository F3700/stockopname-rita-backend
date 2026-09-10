package dto

type InspectorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	RackAssigned  int    `json:"rackAssigned"`
	RackCompleted int    `json:"rackCompleted"`
	TotalItems    int    `json:"totalItems"`
}

type InspectorRequest struct {
	SesiCode      string   `json:"sesi_code" validate:"required"`
	CoorCode      string   `json:"coor_code" validate:"required"`
	InspectorCode string   `json:"inspector_code" validate:"required"`
	Rak           []string `json:"rak" validate:"required"`
}

type InspectorJoinResponse struct {
	InspectorID int            `json:"inspector_id"`
	Rak         []RackResponse `json:"rak"`
}
