package helper

import (
	"fmt"
	"strconv"
	"strings"

	"stockopname-rita-backend/internal/model"
)

// CoordinatorQRPrefix is the fixed prefix embedded in coordinator QR codes.
// QR payload format: RITA-COOR-<coordinator_id>, e.g. RITA-COOR-12.
const CoordinatorQRPrefix = "RITA-COOR-"

// FormatCoordinatorQR builds the QR payload string for a coordinator ID.
// Used by web/desktop clients to render QR codes.
func FormatCoordinatorQR(coordinatorID int) string {
	return fmt.Sprintf("%s%d", CoordinatorQRPrefix, coordinatorID)
}

// ParseCoordinatorQR parses a scanned QR payload into a coordinator ID.
// Accepts surrounding whitespace and case-insensitive prefix.
// Returns *model.ValidationError on malformed input so handlers map it to 400.
func ParseCoordinatorQR(qr string) (int, error) {
	trimmed := strings.TrimSpace(qr)
	if trimmed == "" {
		return 0, &model.ValidationError{Detail: fmt.Sprintf("invalid coordinator QR: expected format %s<id>", CoordinatorQRPrefix)}
	}
	upper := strings.ToUpper(trimmed)
	if !strings.HasPrefix(upper, CoordinatorQRPrefix) {
		return 0, &model.ValidationError{Detail: fmt.Sprintf("invalid coordinator QR %q: expected format %s<id>", qr, CoordinatorQRPrefix)}
	}
	numPart := strings.TrimSpace(upper[len(CoordinatorQRPrefix):])
	id, err := strconv.Atoi(numPart)
	if err != nil || id <= 0 {
		return 0, &model.ValidationError{Detail: fmt.Sprintf("invalid coordinator QR %q: expected format %s<id>", qr, CoordinatorQRPrefix)}
	}
	return id, nil
}
