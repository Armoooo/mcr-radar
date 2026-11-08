package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

const openMeteoEndpoint = "https://api.open-meteo.com/v1/forecast"

type Client struct {
	HTTPClient *http.Client

	mu        sync.Mutex
	cachedAt  time.Time
	forecasts []domain.Forecast
}

func (c *Client) Forecast(ctx context.Context) ([]domain.Forecast, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.cachedAt) < 30*time.Minute && len(c.forecasts) > 0 {
		return append([]domain.Forecast(nil), c.forecasts...), nil
	}

	query := url.Values{
		"latitude":      {"53.4808"},
		"longitude":     {"-2.2426"},
		"daily":         {"weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max"},
		"timezone":      {"Europe/London"},
		"forecast_days": {"10"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openMeteoEndpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build weather request: %w", err)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request weather: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected weather status %s", resp.Status)
	}

	var payload struct {
		Daily struct {
			Time        []string  `json:"time"`
			Maximum     []float64 `json:"temperature_2m_max"`
			Minimum     []float64 `json:"temperature_2m_min"`
			RainChance  []int     `json:"precipitation_probability_max"`
			WeatherCode []int     `json:"weather_code"`
		} `json:"daily"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode weather: %w", err)
	}

	forecasts := make([]domain.Forecast, 0, len(payload.Daily.Time))
	for index, date := range payload.Daily.Time {
		if index >= len(payload.Daily.Maximum) || index >= len(payload.Daily.Minimum) || index >= len(payload.Daily.RainChance) || index >= len(payload.Daily.WeatherCode) {
			break
		}
		forecasts = append(forecasts, domain.Forecast{
			Date:           date,
			Minimum:        payload.Daily.Minimum[index],
			Maximum:        payload.Daily.Maximum[index],
			RainChance:     payload.Daily.RainChance[index],
			WeatherCode:    payload.Daily.WeatherCode[index],
			Recommendation: recommendation(payload.Daily.RainChance[index], payload.Daily.Maximum[index]),
		})
	}
	c.cachedAt = time.Now()
	c.forecasts = forecasts
	return append([]domain.Forecast(nil), forecasts...), nil
}

func recommendation(rainChance int, maximum float64) string {
	switch {
	case rainChance >= 65:
		return "Plan an indoor backup"
	case rainChance >= 35:
		return "Take a light jacket"
	case maximum >= 18:
		return "Good for outdoor plans"
	default:
		return "Dry, but bring a layer"
	}
}
