package domain

import "time"

// Event is the provider-independent representation used throughout MCR Radar.
type Event struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Venue     string    `json:"venue"`
	Area      string    `json:"area"`
	Category  string    `json:"category"`
	Summary   string    `json:"summary"`
	StartsAt  time.Time `json:"startsAt"`
	Price     string    `json:"price"`
	Source    string    `json:"source"`
	SourceURL string    `json:"sourceUrl,omitempty"`
	ImageURL  string    `json:"imageUrl,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
}

// Forecast is a small, UI-focused view of a daily weather forecast.
type Forecast struct {
	Date           string  `json:"date"`
	Minimum        float64 `json:"minimum"`
	Maximum        float64 `json:"maximum"`
	RainChance     int     `json:"rainChance"`
	WeatherCode    int     `json:"weatherCode"`
	Recommendation string  `json:"recommendation"`
}
