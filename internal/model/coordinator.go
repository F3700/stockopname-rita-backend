package model

type Coordinator struct {
	CoorID     int    `json:"coor_id"`
	CoorCode   string `json:"coor_code"`
	CoorSesiID int    `json:"coor_sesi_id"`
	CoorStatus string `json:"coor_status"`
}

type CoordinatorSummary struct {
	ID            int
	Code          string
	Inspector     int
	RackAssigned  int
	RackCompleted int
	Status        string
}

type CoordinatorReport struct {
	CoordinatorSummary
	SessionCode string
}
