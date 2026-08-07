package lifecycle

import "testing"

func TestLifecycleAllowsOnlyDocumentedTransitions(t *testing.T) {
	cases := []struct {
		from, to State
		ok       bool
	}{
		{Pending, Processing, true}, {Pending, Cancelled, true}, {Pending, Expired, true},
		{Processing, Completed, true}, {Processing, Failed, true}, {Completed, Refunded, true},
		{Pending, Completed, false}, {Completed, Failed, false}, {Failed, Completed, false},
	}
	for _, tc := range cases {
		if got := CanTransition(tc.from, tc.to); got != tc.ok {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestInvalidTransitionIsTyped(t *testing.T) {
	err := ValidateTransition(Pending, Completed)
	if _, ok := err.(*InvalidTransitionError); !ok {
		t.Fatalf("error type = %T, want *InvalidTransitionError", err)
	}
}
