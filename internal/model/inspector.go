package model

type Inspector struct {
	InspectorID   int    `json:"inspector_id"`
	InspectorCode string `json:"inspector_code"`
	StartedAt     string `json:"inspector_startedat"`
	EndedAt       string `json:"inspector_endedat"`
	CoordinatorID int    `json:"inspector_coor_id"`
}

type InspectorSummary struct {
	InspectorID   int    `json:"inspector_id"`
	InspectorCode string `json:"inspector_code"`
	RackAssigned  int    `json:"rack_assigned"`
	RackCompleted int    `json:"rack_completed"`
	TotalItems    int    `json:"total_items"`
}
