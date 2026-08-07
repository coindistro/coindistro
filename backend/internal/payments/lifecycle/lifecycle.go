// Package lifecycle owns the valid states and transitions for gateway payments.
package lifecycle

import "fmt"

// State is persisted in each payment transaction's status column.
type State string

const (
	Pending    State = "pending"
	Processing State = "processing"
	Completed  State = "completed"
	Failed     State = "failed"
	Cancelled  State = "cancelled"
	Expired    State = "expired"
	Refunded   State = "refunded"
)

// InvalidTransitionError identifies a rejected lifecycle move without relying
// on fragile string matching at callers.
type InvalidTransitionError struct{ From, To State }

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("invalid payment lifecycle transition: %s -> %s", e.From, e.To)
}

// CanTransition reports whether a lifecycle transition is allowed.
func CanTransition(from, to State) bool {
	return (from == Pending && (to == Processing || to == Cancelled || to == Expired)) ||
		(from == Processing && (to == Completed || to == Failed)) ||
		(from == Completed && to == Refunded)
}

// ValidateTransition returns a typed error when the requested transition is invalid.
func ValidateTransition(from, to State) error {
	if !CanTransition(from, to) {
		return &InvalidTransitionError{From: from, To: to}
	}
	return nil
}

// TimestampColumn returns the immutable timestamp set for a state transition.
func TimestampColumn(to State) string {
	switch to {
	case Processing:
		return "processing_at"
	case Completed:
		return "completed_at"
	case Failed:
		return "failed_at"
	case Cancelled:
		return "cancelled_at"
	case Expired:
		return "expired_at"
	case Refunded:
		return "refunded_at"
	default:
		return ""
	}
}

func EventName(to State) string { return "payment." + string(to) }
