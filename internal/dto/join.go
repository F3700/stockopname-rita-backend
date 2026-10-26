package dto

type JoinByCoordinatorQRRequest struct {
	CoordinatorQR string `json:"coordinator_qr" validate:"required"`
}

// JoinRackResponse carries upload status so a second device joining the same
// coordinator can skip racks that are already done.
type JoinRackResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UploadedAt string `json:"uploaded_at"`
}

type CoordinatorJoinResponse struct {
	CoordinatorID int                `json:"coordinator_id"`
	SesiID        int                `json:"sesi_id"`
	Rak           []JoinRackResponse `json:"rak"`
}
