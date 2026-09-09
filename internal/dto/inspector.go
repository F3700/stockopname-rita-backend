package dto

type InspectorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	RackAssigned  int    `json:"rackAssigned"`
	RackCompleted int    `json:"rackCompleted"`
	TotalItems    int    `json:"totalItems"`
}
