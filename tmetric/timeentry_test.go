package tmetric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getIso8601Duration(t *testing.T) {
	type args struct {
		startTime string
		endTime   string
	}
	tests := []struct {
		name              string
		timeEntry         TimeEntry
		wantISO8601Period string
		wantSpendOn       string
		wantErr           bool
		wantErrMessage    string
	}{
		{
			name: "Valid time range",
			timeEntry: TimeEntry{
				StartTime: "2024-01-31T10:10:20", EndTime: "2024-01-31T10:30:00",
			},
			wantISO8601Period: "P0DT0H19M40S",
			wantSpendOn:       "2024-01-31",
			wantErr:           false,
		},
		{
			name: "more than 24h",
			timeEntry: TimeEntry{
				StartTime: "2022-01-01T08:00:00", EndTime: "2022-01-02T10:31:00"},
			wantISO8601Period: "P1DT2H31M0S",
			wantSpendOn:       "2022-01-01",
			wantErr:           false,
		},
		{
			name: "end of year",
			timeEntry: TimeEntry{
				StartTime: "2022-12-31T08:00:00", EndTime: "2022-12-31T10:31:00"},
			wantISO8601Period: "P0DT2H31M0S",
			wantSpendOn:       "2022-12-31",
			wantErr:           false,
		},
		{
			name: "not leap year",
			timeEntry: TimeEntry{
				StartTime: "2023-02-28T08:00:00", EndTime: "2023-02-29T10:31:00"},
			wantISO8601Period: "",
			wantSpendOn:       "",
			wantErr:           true,
			wantErrMessage:    "failed to parse endTime: parsing time \"2023-02-29T10:31:00\": day out of range",
		},
		{
			name: "Invalid start time format",
			timeEntry: TimeEntry{
				StartTime: "invalid", EndTime: "2000-01-31T10:30:00"},
			wantISO8601Period: "",
			wantSpendOn:       "",
			wantErr:           true,
			wantErrMessage:    "failed to parse startTime: parsing time \"invalid\" as \"2006-01-02T15:04:05\": cannot parse \"invalid\" as \"2006\"",
		},
		{
			name: "Invalid end time format",
			timeEntry: TimeEntry{
				StartTime: "2000-01-31T08:00:00", EndTime: "invalid"},
			wantISO8601Period: "",
			wantSpendOn:       "",
			wantErr:           true,
			wantErrMessage:    "failed to parse endTime: parsing time \"invalid\" as \"2006-01-02T15:04:05\": cannot parse \"invalid\" as \"2006\"",
		},
		{
			name: "End time before start time",
			timeEntry: TimeEntry{
				StartTime: "2000-01-31T10:30:00", EndTime: "2000-01-31T08:00:00"},
			wantISO8601Period: "",
			wantSpendOn:       "",
			wantErr:           true,
			wantErrMessage:    "end time is before start time",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ISO8601Period, spendOn, err := tt.timeEntry.getIso8601Duration()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.EqualErrorf(t, err, tt.wantErrMessage, "convertTime() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ISO8601Period != tt.wantISO8601Period {
				t.Errorf("got ISO8601Period = %v, want %v", ISO8601Period, tt.wantISO8601Period)
			}
			if spendOn != tt.wantSpendOn {
				t.Errorf("got SpendOn = %v, want %v", spendOn, tt.wantSpendOn)
			}
		})
	}
}
