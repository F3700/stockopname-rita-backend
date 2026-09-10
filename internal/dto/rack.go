package dto

type RackResponse struct {
	RackID   int    `json:"id"`
	RackName string `json:"name"`
}

type RackProgressResponse struct {
	RackAssigned  int `json:"rackAssigned"`
	RackCompleted int `json:"rackCompleted"`
	TotalItems    int `json:"total_items"`
}

type CreateRackRequest struct {
	InspectorID int    `json:"inspector_id"`
	RackName    string `json:"rak_name"`
}
