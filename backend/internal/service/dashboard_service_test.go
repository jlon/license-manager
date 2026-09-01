package service

import (
	"testing"
	"time"

	"license-manager/internal/models"
)

func TestParseDashboardBusinessTrendRange(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	now := time.Date(2026, 9, 1, 16, 0, 0, 0, loc)
	tests := []struct {
		name      string
		req       models.DashboardBusinessTrendsRequest
		wantType  string
		wantStart string
		wantEnd   string
		wantCode  string
	}{
		{name: "default 30 days", req: models.DashboardBusinessTrendsRequest{Timezone: "Asia/Shanghai"}, wantType: "30d", wantStart: "2026-08-03", wantEnd: "2026-09-01"},
		{name: "seven days", req: models.DashboardBusinessTrendsRequest{Period: "7d", Timezone: "Asia/Shanghai"}, wantType: "7d", wantStart: "2026-08-26", wantEnd: "2026-09-01"},
		{name: "custom", req: models.DashboardBusinessTrendsRequest{Period: "custom", StartDate: "2026-08-01", EndDate: "2026-08-31", Timezone: "Asia/Shanghai"}, wantType: "custom", wantStart: "2026-08-01", wantEnd: "2026-08-31"},
		{name: "missing custom dates", req: models.DashboardBusinessTrendsRequest{Period: "custom"}, wantCode: "400002"},
		{name: "reversed dates", req: models.DashboardBusinessTrendsRequest{Period: "custom", StartDate: "2026-09-02", EndDate: "2026-09-01"}, wantCode: "400005"},
		{name: "over 365 days", req: models.DashboardBusinessTrendsRequest{Period: "custom", StartDate: "2025-09-01", EndDate: "2026-09-01"}, wantCode: "400004"},
		{name: "invalid timezone", req: models.DashboardBusinessTrendsRequest{Period: "30d", Timezone: "Invalid/Zone"}, wantCode: "400001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			periodType, startDate, endDate, rangeErr := parseDashboardBusinessTrendRange(&tt.req, now)
			if tt.wantCode != "" {
				if rangeErr == nil || rangeErr.code != tt.wantCode {
					t.Fatalf("got error %v, want code %s", rangeErr, tt.wantCode)
				}
				return
			}
			if rangeErr != nil {
				t.Fatalf("unexpected error: %v", rangeErr)
			}
			if periodType != tt.wantType || startDate.Format("2006-01-02") != tt.wantStart || endDate.Format("2006-01-02") != tt.wantEnd {
				t.Fatalf("got %s %s %s", periodType, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			}
		})
	}
}
