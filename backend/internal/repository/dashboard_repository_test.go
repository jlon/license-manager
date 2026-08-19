package repository

import (
	"testing"
	"time"
)

func TestBuildAuthorizationTrendDataFillsDatesAndUsesTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, loc)
	endExclusive := start.AddDate(0, 0, 3)
	createdTimes := []time.Time{
		time.Date(2026, 7, 31, 16, 30, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 10, 0, 0, 0, loc),
		time.Date(2026, 8, 2, 20, 0, 0, 0, loc),
	}
	expiredTimes := []time.Time{
		time.Date(2026, 8, 3, 1, 0, 0, 0, loc),
	}

	trend := buildAuthorizationTrendData(start, endExclusive, 4, createdTimes, expiredTimes)
	if len(trend) != 3 {
		t.Fatalf("expected 3 days, got %d", len(trend))
	}
	wantTotal := []int64{5, 7, 7}
	wantNew := []int64{1, 2, 0}
	wantExpired := []int64{0, 0, 1}
	for i := range trend {
		if trend[i].TotalAuthorizations != wantTotal[i] || trend[i].NewAuthorizations != wantNew[i] || trend[i].ExpiredAuthorizations != wantExpired[i] {
			t.Fatalf("day %d: unexpected trend point: %+v", i, trend[i])
		}
	}
}
