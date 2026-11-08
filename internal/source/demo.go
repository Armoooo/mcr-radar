package source

import (
	"context"
	"fmt"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

// Demo provides representative Manchester events when API keys are absent.
type Demo struct{}

func (Demo) Name() string { return "Demo" }

func (Demo) Fetch(_ context.Context, from, _ time.Time) ([]domain.Event, error) {
	friday := nextWeekday(from, time.Friday)
	saturday := nextWeekday(from, time.Saturday)
	sunday := nextWeekday(from, time.Sunday)

	return []domain.Event{
		demoEvent("record-fair", "Northern Quarter Record Fair", "The Makers Hall", "Northern Quarter", "Music", "Independent sellers, rare finds and a relaxed afternoon soundtrack.", friday, 18, 30, "Free"),
		demoEvent("canal-comedy", "Canal Street Comedy Club", "The Lantern Room", "Canal Street", "Comedy", "A mixed bill of northern comics in an intimate city-centre room.", friday, 20, 0, "From £12"),
		demoEvent("ancoats-supper", "Ancoats After Dark", "Cotton Yard", "Ancoats", "Food & drink", "Street food, small plates and local DJs under the courtyard lights.", saturday, 17, 0, "Free entry"),
		demoEvent("castlefield-cinema", "Canal-side Film Night", "Castlefield Bowl", "Castlefield", "Film", "An open-air screening with independent food traders nearby.", saturday, 20, 30, "From £8"),
		demoEvent("design-market", "Manchester Design Market", "New Century", "NOMA", "Arts", "Print, ceramics, textiles and homeware from independent makers.", sunday, 11, 0, "£3"),
		demoEvent("city-run", "Sunday Social 5K", "Coffee Cranks", "MediaCity", "Sport", "A friendly waterside run finishing with coffee and pastries.", sunday, 9, 30, "Free"),
	}, nil
}

func demoEvent(id, title, venue, area, category, summary string, day time.Time, hour, minute int, price string) domain.Event {
	return domain.Event{
		ID:        "demo-" + id,
		Title:     title,
		Venue:     venue,
		Area:      area,
		Category:  category,
		Summary:   summary,
		StartsAt:  time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, day.Location()),
		Price:     price,
		Source:    "Demo",
		Latitude:  53.4808,
		Longitude: -2.2426,
	}
}

func nextWeekday(from time.Time, weekday time.Weekday) time.Time {
	days := (int(weekday) - int(from.Weekday()) + 7) % 7
	if days == 0 && from.Hour() >= 12 {
		days = 7
	}
	return from.AddDate(0, 0, days)
}

func demoID(index int) string { return fmt.Sprintf("demo-%d", index) }
