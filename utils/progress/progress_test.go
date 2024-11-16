package progress

import (
	"testing"
)

func TestNewProgressBarInitializesCorrectly(t *testing.T) {
	bar := NewProgressBar(0, 100)
	if bar.cur != 0 {
		t.Errorf("expected cur to be 0, got %d", bar.cur)
	}
	if bar.total != 100 {
		t.Errorf("expected total to be 100, got %d", bar.total)
	}
	if bar.percent != 0 {
		t.Errorf("expected percent to be 0, got %d", bar.percent)
	}
	if bar.rate != "" {
		t.Errorf("expected rate to be empty, got %s", bar.rate)
	}
}

func TestGetPercentCalculatesCorrectly(t *testing.T) {
	bar := NewProgressBar(0, 100)
	bar.cur = 50
	expected := 50
	if bar.getPercent() != expected {
		t.Errorf("expected percent to be %d, got %d", expected, bar.getPercent())
	}
}

func TestPlayUpdatesProgressCorrectly(t *testing.T) {
	bar := NewProgressBar(0, 100)
	bar.Play(50)
	if bar.cur != 50 {
		t.Errorf("expected cur to be 50, got %d", bar.cur)
	}
	if bar.percent != 50 {
		t.Errorf("expected percent to be 50, got %d", bar.percent)
	}
}

func TestFinishCompletesProgress(t *testing.T) {
	bar := NewProgressBar(0, 100)
	bar.Finish()
	if bar.cur != 100 {
		t.Errorf("expected cur to be 100, got %d", bar.cur)
	}
	if bar.percent != 100 {
		t.Errorf("expected percent to be 100, got %d", bar.percent)
	}
}
