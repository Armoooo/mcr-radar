package catalog

import (
	"testing"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

func TestDeduplicateKeepsMoreCompleteEvent(t *testing.T) {
	startsAt := time.Date(2026, 11, 14, 19, 0, 0, 0, time.UTC)
	events := []domain.Event{
		{ID: "one", Title: "A Great Gig", Venue: "The Hall", StartsAt: startsAt, Source: "one"},
		{ID: "two", Title: "A great gig!", Venue: "The Hall", StartsAt: startsAt, Source: "two", Summary: "Full details", Price: "£10"},
	}

	result := deduplicate(events)
	if len(result) != 1 {
		t.Fatalf("expected one event, got %d", len(result))
	}
	if result[0].ID != "two" {
		t.Fatalf("expected more complete event, got %q", result[0].ID)
	}
}

func TestDeduplicateDoesNotMergeDifferentDays(t *testing.T) {
	startsAt := time.Date(2026, 11, 14, 19, 0, 0, 0, time.UTC)
	events := []domain.Event{
		{ID: "one", Title: "A Great Gig", Venue: "The Hall", StartsAt: startsAt},
		{ID: "two", Title: "A Great Gig", Venue: "The Hall", StartsAt: startsAt.AddDate(0, 0, 1)},
	}
	if got := len(deduplicate(events)); got != 2 {
		t.Fatalf("expected two events, got %d", got)
	}
}
