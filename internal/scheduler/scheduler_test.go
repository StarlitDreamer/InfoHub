package scheduler

import (
	"testing"
	"time"
)

func TestIntervalScheduleNext(t *testing.T) {
	schedule := intervalSchedule{interval: 30 * time.Minute}
	after := time.Date(2026, 5, 4, 10, 7, 12, 0, time.UTC)
	next := schedule.Next(after)
	expected := time.Date(2026, 5, 4, 10, 37, 12, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, next)
	}
}

func TestParseCronNextForQuarterHour(t *testing.T) {
	spec, err := ParseCron("*/15 * * * *")
	if err != nil {
		t.Fatalf("parse cron failed: %v", err)
	}

	after := time.Date(2026, 5, 4, 10, 7, 12, 0, time.UTC)
	next := spec.Next(after)
	expected := time.Date(2026, 5, 4, 10, 15, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, next)
	}
}

func TestParseCronNextForWeekdayMorning(t *testing.T) {
	spec, err := ParseCron("30 9 * * 1-5")
	if err != nil {
		t.Fatalf("parse cron failed: %v", err)
	}

	after := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	next := spec.Next(after)
	expected := time.Date(2026, 5, 11, 9, 30, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, next)
	}
}

func TestParseCronSupportsDayOrWeekRule(t *testing.T) {
	spec, err := ParseCron("0 9 15 * 1")
	if err != nil {
		t.Fatalf("parse cron failed: %v", err)
	}

	after := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	next := spec.Next(after)
	expected := time.Date(2026, 5, 15, 9, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, next)
	}
}

func TestParseCronRejectsInvalidSpec(t *testing.T) {
	if _, err := ParseCron("not-a-cron"); err == nil {
		t.Fatal("expected invalid cron spec to fail")
	}
}
