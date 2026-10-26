package dto

type RackResponse struct {
	RackID     int    `json:"id"`
	RackName   string `json:"name"`
	UploadedAt string `json:"uploaded_at"`
}

type RackProgressResponse struct {
	RackAssigned  int `json:"rackAssigned"`
	RackCompleted int `json:"rackCompleted"`
	TotalItems    int `json:"total_items"`
}
