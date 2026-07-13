package errors

import (
	"net/http"

	apperrors "github.com/coindistro/backend/internal/errors"
)

var (
	ErrProductNotFound       = apperrors.New("EARN_PRODUCT_NOT_FOUND", "Earn product not found", http.StatusNotFound)
	ErrProductNotActive      = apperrors.New("EARN_PRODUCT_NOT_ACTIVE", "Earn product is not active", http.StatusBadRequest)
	ErrCategoryDisabled      = apperrors.New("EARN_CATEGORY_DISABLED", "This earn category is disabled", http.StatusServiceUnavailable)
	ErrEarnDisabled          = apperrors.New("EARN_DISABLED", "Earn module is disabled", http.StatusServiceUnavailable)
	ErrParticipationNotFound = apperrors.New("EARN_PARTICIPATION_NOT_FOUND", "Participation not found", http.StatusNotFound)
	ErrInvalidAllocation     = apperrors.New("EARN_INVALID_ALLOCATION", "Allocation amount is invalid for this product", http.StatusBadRequest)
	ErrCapacityExceeded      = apperrors.New("EARN_CAPACITY_EXCEEDED", "Product capacity exceeded", http.StatusConflict)
	ErrExitNotAllowed        = apperrors.New("EARN_EXIT_NOT_ALLOWED", "Exit is not allowed for this participation", http.StatusBadRequest)
	ErrAlreadyParticipating  = apperrors.New("EARN_ALREADY_PARTICIPATING", "Already participating in this product", http.StatusConflict)
	ErrInvalidAsset          = apperrors.New("EARN_INVALID_ASSET", "Asset is not supported by this product", http.StatusBadRequest)
	ErrInvalidDuration       = apperrors.New("EARN_INVALID_DURATION", "Invalid fixed earn duration", http.StatusBadRequest)
	ErrInvalidStrategy       = apperrors.New("EARN_INVALID_STRATEGY", "Invalid strategy profile", http.StatusBadRequest)
	ErrCampaignNotFound      = apperrors.New("EARN_CAMPAIGN_NOT_FOUND", "Campaign not found", http.StatusNotFound)
	ErrAlreadyCompleted      = apperrors.New("EARN_ALREADY_COMPLETED", "Learning campaign already completed", http.StatusConflict)
	ErrSlugExists            = apperrors.New("EARN_SLUG_EXISTS", "Product slug already exists", http.StatusConflict)
)
