package dto

type RackResponse struct {
	RackID        int    `json:"id"`
	RackName      string `json:"name"`
	CoordinatorID int    `json:"coordinator_id"`
	UploadedAt    string `json:"uploaded_at"`
	ItemCount     int    `json:"item_count"`
	UnownedCount  int    `json:"unowned_count"`
	CanUpload     bool   `json:"can_upload"`
	IsEmpty       bool   `json:"is_empty"`
}

type RackProgressResponse struct {
	RackAssigned  int `json:"rackAssigned"`
	RackCompleted int `json:"rackCompleted"`
	TotalItems    int `json:"total_items"`
}
