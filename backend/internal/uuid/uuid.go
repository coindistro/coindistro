package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// UUIDv7 represents a UUID version 7 value.
// UUIDv7 is a time-ordered UUID with millisecond precision,
// making it ideal for database primary keys and sorting.
type UUIDv7 [16]byte

// String returns the standard UUID string representation.
func (u UUIDv7) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// Bytes returns the raw 16-byte slice.
func (u UUIDv7) Bytes() []byte {
	return u[:]
}

// Time returns the embedded timestamp.
func (u UUIDv7) Time() time.Time {
	ms := binary.BigEndian.Uint64(u[0:8]) >> 16
	return time.UnixMilli(int64(ms))
}

var (
	lastTimestamp int64
	sequence      uint16
	sequenceMu    sync.Mutex
)

// New generates a new UUIDv7.
// Format: 48-bit timestamp (ms) | 74-bit random | 2-bit variant (10)
func New() UUIDv7 {
	var u UUIDv7

	// 48-bit timestamp (milliseconds since Unix epoch)
	now := time.Now().UnixMilli()

	// Handle sequence counter for same-millisecond collisions
	sequenceMu.Lock()
	if now == lastTimestamp {
		sequence++
	} else {
		sequence = 0
		lastTimestamp = now
	}
	seq := sequence
	sequenceMu.Unlock()

	// Timestamp: bytes 0-5
	binary.BigEndian.PutUint64(u[0:8], uint64(now)<<16|uint64(seq))

	// Version: byte 6, high nibble = 7
	u[6] = (u[6] & 0x0f) | 0x70

	// Variant: byte 8, high bits = 10
	u[8] = (u[8] & 0x3f) | 0x80

	// Random: bytes 7, 9-15
	_, _ = rand.Read(u[7:8])
	_, _ = rand.Read(u[9:16])

	return u
}

// Parse parses a UUID string into a UUIDv7.
func Parse(s string) (UUIDv7, error) {
	if len(s) != 36 {
		return UUIDv7{}, fmt.Errorf("invalid UUID length: %d", len(s))
	}

	var u UUIDv7
	// Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := []struct {
		start int
		end   int
		dest  []byte
	}{
		{0, 8, u[0:4]},
		{9, 13, u[4:6]},
		{14, 18, u[6:8]},
		{19, 23, u[8:10]},
		{24, 36, u[10:16]},
	}

	for _, p := range parts {
		_, err := fmt.Sscanf(s[p.start:p.end], "%x", &p.dest)
		if err != nil {
			return UUIDv7{}, fmt.Errorf("invalid UUID format: %w", err)
		}
	}

	return u, nil
}

// MustParse parses a UUID string and panics on error.
func MustParse(s string) UUIDv7 {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// Nil returns the zero UUID.
func Nil() UUIDv7 {
	return UUIDv7{}
}

// IsNil checks if the UUID is nil.
func (u UUIDv7) IsNil() bool {
	return u == UUIDv7{}
}

// NewString generates a new UUIDv7 and returns its string representation.
func NewString() string {
	return New().String()
}
