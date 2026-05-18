package usecase

import (
	"math"
	"testing"
	"time"
)

func TestCalculateTransferFee(t *testing.T) {
	uc := &TransactionUsecase{}

	cases := []struct {
		name   string
		amount float64
		typeIn string
		want   float64
	}{
		{"internal", 100, "INTERNAL", 0},
		{"external", 100, "EXTERNAL", 1},
		{"default", 100, "OTHER", 0.5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := uc.CalculateTransferFee(nil, c.amount, c.typeIn)
			if math.Abs(got-c.want) > 0.0001 {
				t.Fatalf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestParseStatementRangeInvalidDate(t *testing.T) {
	_, _, err := parseStatementRange("bad", "2026-05-10")
	if err == nil {
		t.Fatal("expected error for invalid start date")
	}
}

func TestParseStatementRangeExplicit(t *testing.T) {
	start, end, err := parseStatementRange("2026-05-01", "2026-05-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 5, 10, 23, 59, 59, 0, time.UTC)

	if !start.Equal(wantStart) {
		t.Fatalf("expected start %v, got %v", wantStart, start)
	}
	if !end.Equal(wantEnd) {
		t.Fatalf("expected end %v, got %v", wantEnd, end)
	}
}

func TestParseStatementRangeDefault(t *testing.T) {
	start, end, err := parseStatementRange("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !end.After(start) {
		t.Fatal("expected end after start")
	}

	dur := end.Sub(start)
	min := 29 * 24 * time.Hour
	max := 31 * 24 * time.Hour
	if dur < min || dur > max {
		t.Fatalf("unexpected range duration: %v", dur)
	}
}
