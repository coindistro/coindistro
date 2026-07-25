package config

import (
	"testing"
	"time"
)

func TestParseFlexibleDuration_NumericMinutes(t *testing.T) {
	cases := []struct {
		name string
		raw  interface{}
		want time.Duration
	}{
		{"int", 15, 15 * time.Minute},
		{"int64", int64(10080), 10080 * time.Minute},
		{"float64", 15.0, 15 * time.Minute},
		{"string int", "15", 15 * time.Minute},
		{"string large", "10080", 10080 * time.Minute},
		{"zero", 0, 0},
		{"string zero", "0", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFlexibleDuration(tc.raw, time.Minute)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestParseFlexibleDuration_GoDurationStrings(t *testing.T) {
	cases := []struct {
		raw  string
		want time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"1h", time.Hour},
		{"168h", 168 * time.Hour},
		{"90s", 90 * time.Second},
		{"1h30m", time.Hour + 30*time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got, err := ParseFlexibleDuration(tc.raw, time.Minute)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestParseFlexibleDuration_Invalid(t *testing.T) {
	cases := []interface{}{
		"",
		"   ",
		"not-a-duration",
		"15d", // Go does not support days
		"-5",
		"-15m",
	}
	for _, raw := range cases {
		t.Run(fmtRaw(raw), func(t *testing.T) {
			_, err := ParseFlexibleDuration(raw, time.Minute)
			if err == nil {
				t.Fatalf("expected error for %v", raw)
			}
		})
	}
}

func TestParseFlexibleDuration_Nil(t *testing.T) {
	_, err := ParseFlexibleDuration(nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil")
	}
}

func TestParseFlexibleDuration_BareDurationCount(t *testing.T) {
	// mapstructure may hand us time.Duration(15) meaning "15 units"
	got, err := ParseFlexibleDuration(time.Duration(15), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if got != 15*time.Minute {
		t.Fatalf("got %v want 15m", got)
	}
}

func TestParseFlexibleDuration_RealDurationUnchanged(t *testing.T) {
	in := 15 * time.Minute
	got, err := ParseFlexibleDuration(in, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Fatalf("got %v want %v", got, in)
	}
}

func fmtRaw(v interface{}) string {
	if s, ok := v.(string); ok {
		if s == "" {
			return "empty"
		}
		return s
	}
	return "value"
}
