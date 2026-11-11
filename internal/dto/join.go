package dto

type JoinByCoordinatorQRRequest struct {
	CoordinatorQR string `json:"coordinator_qr" validate:"required"`
}

// JoinRackResponse carries upload status so a second device joining the same
// coordinator can skip racks that are already done. CanUpload mirrors the claim
// guard, so clients should gate the upload button on it instead of is_empty.
type JoinRackResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	UploadedAt   string `json:"uploaded_at"`
	ItemCount    int    `json:"item_count"`
	UnownedCount int    `json:"unowned_count"`
	CanUpload    bool   `json:"can_upload"`
	IsEmpty      bool   `json:"is_empty"`
}

type CoordinatorJoinResponse struct {
	CoordinatorID int                `json:"coordinator_id"`
	SesiID        int                `json:"sesi_id"`
	Rak           []JoinRackResponse `json:"rak"`
}
