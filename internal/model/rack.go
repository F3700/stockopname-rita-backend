package model

type Rack struct {
	RackID      int    `json:"rak_id"`
	RackName    string `json:"rak_name"`
	InspectorID int    `json:"rak_inspector_id"`
}

type RackProgress struct {
	RackAssigned  int `json:"rack_assigned"`
	RackCompleted int `json:"rack_completed"`
	TotalItems    int `json:"total_items"`
}
