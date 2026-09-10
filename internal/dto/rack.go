package dto

type RackResponse struct {
	RackID   int    `json:"rak_id"`
	RackName string `json:"rak_name"`
}

type RackProgressResponse struct {
	RackAssigned  int `json:"rackAssigned"`
	RackCompleted int `json:"rackCompleted"`
	TotalItems    int `json:"total_items"`
}
