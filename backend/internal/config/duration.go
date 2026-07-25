package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// pureNumberPattern matches optional sign + integer or decimal with no unit suffix.
var pureNumberPattern = regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)

// ParseFlexibleDuration parses duration configuration values used across Coindistro.
//
// Accepted forms:
//   - pure numbers (int/float/string): interpreted as units of defaultUnit
//     (for JWT TTLs defaultUnit is time.Minute, so "15" → 15m)
//   - Go duration strings: "15m", "1h", "168h", "90s", etc.
//
// Invalid values return a clear error. Empty string returns an error
// (callers should apply defaults before parsing).
func ParseFlexibleDuration(raw interface{}, defaultUnit time.Duration) (time.Duration, error) {
	if defaultUnit <= 0 {
		defaultUnit = time.Minute
	}

	switch v := raw.(type) {
	case nil:
		return 0, fmt.Errorf("duration value is empty")
	case time.Duration:
		// Already a duration. Values that arrived as bare integers via mapstructure
		// are nanosecond counts (e.g. 15). Treat small magnitudes as unit counts.
		if isLikelyUnitCount(v) {
			return time.Duration(int64(v)) * defaultUnit, nil
		}
		return v, nil
	case int:
		return time.Duration(v) * defaultUnit, nil
	case int8:
		return time.Duration(v) * defaultUnit, nil
	case int16:
		return time.Duration(v) * defaultUnit, nil
	case int32:
		return time.Duration(v) * defaultUnit, nil
	case int64:
		return time.Duration(v) * defaultUnit, nil
	case uint:
		return time.Duration(v) * defaultUnit, nil
	case uint8:
		return time.Duration(v) * defaultUnit, nil
	case uint16:
		return time.Duration(v) * defaultUnit, nil
	case uint32:
		return time.Duration(v) * defaultUnit, nil
	case uint64:
		return time.Duration(v) * defaultUnit, nil
	case float32:
		return time.Duration(float64(v) * float64(defaultUnit)), nil
	case float64:
		return time.Duration(v * float64(defaultUnit)), nil
	case string:
		return parseFlexibleDurationString(v, defaultUnit)
	default:
		// Fall back to string form for uncommon numeric types.
		return parseFlexibleDurationString(fmt.Sprint(v), defaultUnit)
	}
}

// isLikelyUnitCount detects bare integer durations produced by mapstructure
// when YAML/env provides "15" without a unit (stored as 15 nanoseconds).
// Real intentional nanosecond durations are never used for auth TTLs.
func isLikelyUnitCount(d time.Duration) bool {
	if d == 0 {
		return true
	}
	// Any absolute value under 1ms that is a whole number of nanoseconds
	// representing a small integer count (e.g. 15, 10080) is treated as units.
	abs := d
	if abs < 0 {
		abs = -abs
	}
	return abs < time.Millisecond
}

func parseFlexibleDurationString(s string, defaultUnit time.Duration) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("duration value is empty")
	}

	// Pure numeric string → defaultUnit (minutes for JWT TTLs).
	if pureNumberPattern.MatchString(s) {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid numeric duration %q: %w", s, err)
		}
		if f < 0 {
			return 0, fmt.Errorf("duration must be non-negative, got %q", s)
		}
		return time.Duration(f * float64(defaultUnit)), nil
	}

	// Standard Go duration string (ns, us, µs, ms, s, m, h).
	// Note: Go does not support "d" for days — use "168h" for 7 days.
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid duration %q: use a number of minutes (e.g. 15 or 10080) or a Go duration (e.g. 15m, 1h, 168h): %w",
			s, err,
		)
	}
	if d < 0 {
		return 0, fmt.Errorf("duration must be non-negative, got %q", s)
	}
	return d, nil
}

// flexibleDurationDecodeHook is a mapstructure decode hook that converts
// strings/numbers into time.Duration for fields typed as time.Duration.
// Pure numbers are interpreted as minutes (matching Coindistro JWT TTL convention).
func flexibleDurationDecodeHook() func(from, to reflect.Type, data interface{}) (interface{}, error) {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if to != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}
		// Leave nil alone.
		if data == nil {
			return data, nil
		}
		d, err := ParseFlexibleDuration(data, time.Minute)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
}
