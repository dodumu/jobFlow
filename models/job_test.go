package models

import (
	"testing"
	"time"
)

func TestJobDeadlinePassed(t *testing.T) {
	deadline := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	job := Job{Deadline: deadline}

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"well before deadline", time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC), false},
		{"during the deadline day", time.Date(2026, 11, 20, 23, 59, 0, 0, time.UTC), false},
		{"first moment after deadline day", time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC), true},
		{"long after", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := job.DeadlinePassed(tc.now); got != tc.want {
				t.Errorf("DeadlinePassed(%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}
}
