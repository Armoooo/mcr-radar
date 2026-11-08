package source

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

const skiddleEndpoint = "https://www.skiddle.com/api/v1/events/search/"

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

type Skiddle struct {
	APIKey string
	Client *http.Client
}

func (Skiddle) Name() string { return "Skiddle" }

func (s Skiddle) Fetch(ctx context.Context, from, to time.Time) ([]domain.Event, error) {
	query := url.Values{
		"api_key":     {s.APIKey},
		"latitude":    {"53.4808"},
		"longitude":   {"-2.2426"},
		"radius":      {"15"},
		"minDate":     {from.Format("2006-01-02")},
		"maxDate":     {to.Format("2006-01-02")},
		"description": {"1"},
		"order":       {"date"},
		"limit":       {"100"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, skiddleEndpoint+"?"+query.Encode(), nil)
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

	var payload skiddleResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}

	events := make([]domain.Event, 0, len(payload.Results))
	for _, item := range payload.Results {
		startsAt, err := time.ParseInLocation("2006-01-02 15:04:05", item.Date+" "+defaultTime(item.OpeningTimes.DoorsOpen), time.Local)
		if err != nil {
			continue
		}
		events = append(events, domain.Event{
			ID:        fmt.Sprintf("skiddle-%d", item.ID),
			Title:     item.EventName,
			Venue:     item.Venue.Name,
			Area:      firstNonEmpty(item.Venue.Town, item.Venue.City),
			Category:  skiddleCategory(item.EventCode),
			Summary:   cleanText(item.Description),
			StartsAt:  startsAt,
			Price:     firstNonEmpty(item.EntryPrice, "Check tickets"),
			Source:    "Skiddle",
			SourceURL: item.Link,
			ImageURL:  firstNonEmpty(item.LargeImageURL, item.ImageURL),
			Latitude:  item.Venue.Latitude,
			Longitude: item.Venue.Longitude,
		})
	}
	return events, nil
}

type skiddleResponse struct {
	Results []struct {
		ID            int    `json:"id"`
		EventName     string `json:"eventname"`
		Description   string `json:"description"`
		Date          string `json:"date"`
		EventCode     string `json:"EventCode"`
		EntryPrice    string `json:"entryprice"`
		Link          string `json:"link"`
		ImageURL      string `json:"imageurl"`
		LargeImageURL string `json:"largeimageurl"`
		OpeningTimes  struct {
			DoorsOpen string `json:"doorsopen"`
		} `json:"openingtimes"`
		Venue struct {
			Name      string  `json:"name"`
			Town      string  `json:"town"`
			City      string  `json:"city"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"venue"`
	} `json:"results"`
}

func defaultTime(value string) string {
	if len(value) >= 8 {
		return value[:8]
	}
	if len(value) == 5 {
		return value + ":00"
	}
	return "19:00:00"
}

func skiddleCategory(code string) string {
	switch strings.ToUpper(code) {
	case "LIVE", "CLUB", "FEST":
		return "Music"
	case "COMEDY":
		return "Comedy"
	case "THEATRE", "ARTS", "EXHIB":
		return "Arts"
	case "SPORT":
		return "Sport"
	case "BARPUB":
		return "Food & drink"
	case "KIDS":
		return "Family"
	default:
		return "Other"
	}
}

func cleanText(value string) string {
	value = html.UnescapeString(htmlTagPattern.ReplaceAllString(value, " "))
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 220 {
		return value[:217] + "…"
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
