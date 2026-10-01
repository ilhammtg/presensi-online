package app

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilham/presensi-online/backend/internal/domain"
)

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name        string
		lat1, lon1  float64
		lat2, lon2  float64
		expectedM   float64
		toleranceM  float64
	}{
		{
			name:       "same coordinates should yield 0 distance",
			lat1:       -6.200000,
			lon1:       106.816666,
			lat2:       -6.200000,
			lon2:       106.816666,
			expectedM:  0,
			toleranceM: 0.1,
		},
		{
			name:       "approx 100 meters offset",
			lat1:       -6.200000,
			lon1:       106.816666,
			lat2:       -6.200900,
			lon2:       106.816666,
			expectedM:  100.0,
			toleranceM: 5.0,
		},
		{
			name:       "Monas to Istiqlal Mosque Jakarta (~780m)",
			lat1:       -6.175392,
			lon1:       106.827153,
			lat2:       -6.170170,
			lon2:       106.831713,
			expectedM:  768.0,
			toleranceM: 20.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HaversineDistance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
			diff := math.Abs(got - tc.expectedM)
			if diff > tc.toleranceM {
				t.Errorf("HaversineDistance() = %v meters, want %v (+/-%v), diff = %v",
					got, tc.expectedM, tc.toleranceM, diff)
			}
		})
	}
}

func TestEvaluateAttendanceStatus(t *testing.T) {
	schedule := &domain.ClassSchedule{
		StartTime: "08:00:00",
	}

	today := time.Now()
	session := &domain.ClassSession{
		SessionDate: today,
	}

	loc := today.Location()
	// Case 1: Scanned at 08:05 (within 15 min tolerance) -> Hadir
	onTime := time.Date(today.Year(), today.Month(), today.Day(), 8, 5, 0, 0, loc)
	status := evaluateAttendanceStatus(session, schedule, onTime)
	if status != domain.StatusHadir {
		t.Errorf("expected StatusHadir, got %v", status)
	}

	// Case 2: Scanned exactly at 08:15:00 -> Hadir
	exactDeadline := time.Date(today.Year(), today.Month(), today.Day(), 8, 15, 0, 0, loc)
	status = evaluateAttendanceStatus(session, schedule, exactDeadline)
	if status != domain.StatusHadir {
		t.Errorf("expected StatusHadir at boundary, got %v", status)
	}

	// Case 3: Scanned at 08:16 (after 15 min tolerance) -> Terlambat
	late := time.Date(today.Year(), today.Month(), today.Day(), 8, 16, 0, 0, loc)
	status = evaluateAttendanceStatus(session, schedule, late)
	if status != domain.StatusTerlambat {
		t.Errorf("expected StatusTerlambat, got %v", status)
	}
}

func TestGenerateQRSeed(t *testing.T) {
	seed1, err := generateQRSeed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seed1) != 64 {
		t.Errorf("expected 64 hex characters (32 bytes), got %d", len(seed1))
	}

	seed2, err := generateQRSeed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seed1 == seed2 {
		t.Errorf("consecutive seeds should be randomly distinct")
	}
}

func TestSessionRequestValidation(t *testing.T) {
	schedID := uuid.New()
	lecturerID := uuid.New()

	req := OpenSessionRequest{
		ScheduleID: schedID,
		LecturerID: lecturerID,
		MeetingNo:  1,
	}

	if req.MeetingNo < 1 || req.MeetingNo > 16 {
		t.Errorf("meeting number should be between 1 and 16")
	}
}
