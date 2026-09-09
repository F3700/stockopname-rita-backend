package dto

type InspectorResponse struct {
	ID            int    `json:"id"`
	Code          string `json:"code"`
	RackAssigned  string `json:"rackAssigned"`
	RackCompleted string `json:"rackCompleted"`
	TotalItems    int    `json:"totalItems"`
}
