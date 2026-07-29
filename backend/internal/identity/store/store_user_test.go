package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type stubRow struct {
	err error
}

func (s stubRow) Scan(dest ...interface{}) error {
	return s.err
}

func TestScanUserFromRow_NoRowsIsUserNotFound(t *testing.T) {
	s := &Store{}
	user, err := s.scanUserFromRow(stubRow{err: pgx.ErrNoRows})
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	// Must never look like a bare internal wrap that callers fail to classify.
	if err != nil && err.Error() == "failed to scan user: no rows in result set" {
		t.Fatal("ErrNoRows must not be rewrapped without errors.Is support")
	}
}

func TestScanUserFromRow_UnexpectedErrorWrapped(t *testing.T) {
	s := &Store{}
	base := errors.New("connection reset")
	user, err := s.scanUserFromRow(stubRow{err: base})
	if user != nil {
		t.Fatalf("expected nil user")
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, base) {
		t.Fatalf("expected wrapped base error, got %v", err)
	}
	if errors.Is(err, ErrUserNotFound) {
		t.Fatal("unexpected DB errors must not map to ErrUserNotFound")
	}
}
