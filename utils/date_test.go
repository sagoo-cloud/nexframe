package utils

import (
	"testing"
	"time"
)

func TestGetWeekDay(t *testing.T) {
	start, end := GetWeekDay()

	// Parse the returned strings to time.Time
	startTime, _ := time.Parse("2006-01-02 15:04:05", start)
	endTime, _ := time.Parse("2006-01-02 15:04:05", end)

	// Check if start is a Monday
	if startTime.Weekday() != time.Monday {
		t.Errorf("Start day is not Monday, got %v", startTime.Weekday())
	}

	// Check if end is a Sunday
	if endTime.Weekday() != time.Sunday {
		t.Errorf("End day is not Sunday, got %v", endTime.Weekday())
	}

	// Check if the difference between start and end is 6 days, 23 hours, 59 minutes and 59 seconds
	expectedDiff := 7*24*time.Hour - time.Second
	actualDiff := endTime.Sub(startTime)
	if actualDiff != expectedDiff {
		t.Errorf("Difference between start and end is not correct. Expected %v, got %v", expectedDiff, actualDiff)
	}
}
func TestGetBetweenDates(t *testing.T) {
	start := "2023-01-01"
	end := "2023-01-05"
	dates := GetBetweenDates(start, end)

	expected := []string{"2023-01-01", "2023-01-02", "2023-01-03", "2023-01-04", "2023-01-05"}

	if len(dates) != len(expected) {
		t.Errorf("Expected %d dates, got %d", len(expected), len(dates))
	}

	for i, date := range dates {
		if date != expected[i] {
			t.Errorf("Expected date %s, got %s", expected[i], date)
		}
	}
}

func TestGetHourDiffer(t *testing.T) {
	start := "2023-01-01 10:00:00"
	end := "2023-01-01 15:30:00"

	diff := GetHourDiffer(start, end)
	expected := 5.5

	if diff != expected {
		t.Errorf("Expected difference of %f hours, got %f", expected, diff)
	}
}

func TestIsSameDay(t *testing.T) {
	t1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC).Unix()
	t2 := time.Date(2023, 1, 1, 15, 30, 0, 0, time.UTC).Unix()
	t3 := time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC).Unix()

	if !IsSameDay(t1, t2) {
		t.Errorf("Expected t1 and t2 to be the same day")
	}

	if IsSameDay(t1, t3) {
		t.Errorf("Expected t1 and t3 to be different days")
	}
}

func TestGetTimeTagGroup(t *testing.T) {
	result := GetTimeTagGroup()

	now := time.Now()
	expected := now.Format("2006:01:")

	if result != expected {
		t.Errorf("Expected time tag %s, got %s", expected, result)
	}
}
