package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

const ticketmasterEndpoint = "https://app.ticketmaster.com/discovery/v2/events.json"

type Ticketmaster struct {
	APIKey string
	Client *http.Client
}

func (Ticketmaster) Name() string { return "Ticketmaster" }

func (s Ticketmaster) Fetch(ctx context.Context, from, to time.Time) ([]domain.Event, error) {
	query := url.Values{
		"apikey":        {s.APIKey},
		"city":          {"Manchester"},
		"countryCode":   {"GB"},
		"startDateTime": {from.UTC().Format("2006-01-02T15:04:05Z")},
		"endDateTime":   {to.UTC().Format("2006-01-02T15:04:05Z")},
		"size":          {"100"},
		"sort":          {"date,asc"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ticketmasterEndpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request events: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var payload ticketmasterResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}

	events := make([]domain.Event, 0, len(payload.Embedded.Events))
	for _, item := range payload.Embedded.Events {
		startsAt, err := parseTicketmasterTime(item.Dates.Start)
		if err != nil {
			continue
		}
		event := domain.Event{
			ID:        "ticketmaster-" + item.ID,
			Title:     item.Name,
			StartsAt:  startsAt,
			Source:    "Ticketmaster",
			SourceURL: item.URL,
			Category:  "Other",
			Price:     "Check tickets",
		}
		if len(item.Classifications) > 0 && item.Classifications[0].Segment.Name != "" {
			event.Category = normaliseCategory(item.Classifications[0].Segment.Name)
		}
		if len(item.PriceRanges) > 0 {
			event.Price = formatPrice(item.PriceRanges[0].Min, item.PriceRanges[0].Max, item.PriceRanges[0].Currency)
		}
		if len(item.Images) > 0 {
			event.ImageURL = item.Images[0].URL
		}
		if len(item.Embedded.Venues) > 0 {
			venue := item.Embedded.Venues[0]
			event.Venue = venue.Name
			event.Area = venue.City.Name
			event.Latitude = parseFloat(venue.Location.Latitude)
			event.Longitude = parseFloat(venue.Location.Longitude)
		}
		events = append(events, event)
	}
	return events, nil
}

type ticketmasterResponse struct {
	Embedded struct {
		Events []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			URL   string `json:"url"`
			Dates struct {
				Start struct {
					DateTime  string `json:"dateTime"`
					LocalDate string `json:"localDate"`
					LocalTime string `json:"localTime"`
				} `json:"start"`
			} `json:"dates"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			Classifications []struct {
				Segment struct {
					Name string `json:"name"`
				} `json:"segment"`
			} `json:"classifications"`
			PriceRanges []struct {
				Min      float64 `json:"min"`
				Max      float64 `json:"max"`
				Currency string  `json:"currency"`
			} `json:"priceRanges"`
			Embedded struct {
				Venues []struct {
					Name string `json:"name"`
					City struct {
						Name string `json:"name"`
					} `json:"city"`
					Location struct {
						Latitude  string `json:"latitude"`
						Longitude string `json:"longitude"`
					} `json:"location"`
				} `json:"venues"`
			} `json:"_embedded"`
		} `json:"events"`
	} `json:"_embedded"`
}

func parseTicketmasterTime(start struct {
	DateTime  string `json:"dateTime"`
	LocalDate string `json:"localDate"`
	LocalTime string `json:"localTime"`
}) (time.Time, error) {
	if start.DateTime != "" {
		return time.Parse(time.RFC3339, start.DateTime)
	}
	return time.ParseInLocation("2006-01-02 15:04:05", start.LocalDate+" "+start.LocalTime, time.Local)
}

func normaliseCategory(value string) string {
	switch strings.ToLower(value) {
	case "music":
		return "Music"
	case "sports":
		return "Sport"
	case "arts & theatre":
		return "Arts"
	default:
		return value
	}
}

func formatPrice(minimum, maximum float64, currency string) string {
	symbol := currency + " "
	if currency == "GBP" {
		symbol = "£"
	}
	if minimum == maximum {
		return fmt.Sprintf("%s%.0f", symbol, minimum)
	}
	return fmt.Sprintf("%s%.0f–%.0f", symbol, minimum, maximum)
}
