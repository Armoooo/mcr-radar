package catalog

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Armoooo/mcr-radar/internal/domain"
	"github.com/Armoooo/mcr-radar/internal/source"
)

type Snapshot struct {
	Events    []domain.Event `json:"events"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Mode      string         `json:"mode"`
	Warnings  []string       `json:"warnings,omitempty"`
}

// Catalog owns the most recent successful event snapshot.
type Catalog struct {
	mu       sync.RWMutex
	sources  []source.Source
	fallback source.Source
	snapshot Snapshot
}

func New(sources []source.Source, fallback source.Source) *Catalog {
	return &Catalog{sources: sources, fallback: fallback}
}

// Refresh fetches all configured providers concurrently and atomically swaps
// the visible snapshot. One failed provider does not discard successful data.
func (c *Catalog) Refresh(ctx context.Context, from, to time.Time) Snapshot {
	type result struct {
		name   string
		events []domain.Event
		err    error
	}

	results := make(chan result, len(c.sources))
	var group sync.WaitGroup
	for _, provider := range c.sources {
		provider := provider
		group.Add(1)
		go func() {
			defer group.Done()
			events, err := provider.Fetch(ctx, from, to)
			results <- result{name: provider.Name(), events: events, err: err}
		}()
	}
	go func() {
		group.Wait()
		close(results)
	}()

	var allEvents []domain.Event
	var warnings []string
	for item := range results {
		if item.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s is temporarily unavailable", item.name))
			continue
		}
		allEvents = append(allEvents, item.events...)
	}

	mode := "live"
	if len(allEvents) == 0 && c.fallback != nil {
		mode = "demo"
		fallbackEvents, err := c.fallback.Fetch(ctx, from, to)
		if err == nil {
			allEvents = fallbackEvents
		} else {
			warnings = append(warnings, "Demo events are temporarily unavailable")
		}
	}

	allEvents = deduplicate(allEvents)
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].StartsAt.Before(allEvents[j].StartsAt)
	})
	sort.Strings(warnings)

	snapshot := Snapshot{
		Events:    allEvents,
		UpdatedAt: time.Now().UTC(),
		Mode:      mode,
		Warnings:  warnings,
	}
	c.mu.Lock()
	c.snapshot = snapshot
	c.mu.Unlock()
	return cloneSnapshot(snapshot)
}

func (c *Catalog) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return cloneSnapshot(c.snapshot)
}

func cloneSnapshot(snapshot Snapshot) Snapshot {
	snapshot.Events = append([]domain.Event(nil), snapshot.Events...)
	snapshot.Warnings = append([]string(nil), snapshot.Warnings...)
	return snapshot
}

func deduplicate(events []domain.Event) []domain.Event {
	seen := make(map[string]int, len(events))
	result := make([]domain.Event, 0, len(events))
	for _, event := range events {
		key := eventKey(event)
		if existingIndex, exists := seen[key]; exists {
			result[existingIndex] = preferComplete(result[existingIndex], event)
			continue
		}
		seen[key] = len(result)
		result = append(result, event)
	}
	return result
}

func eventKey(event domain.Event) string {
	return strings.Join([]string{
		normalise(event.Title),
		normalise(event.Venue),
		event.StartsAt.Format("2006-01-02"),
	}, "|")
}

func normalise(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, value)
}

func preferComplete(current, candidate domain.Event) domain.Event {
	currentScore := completeness(current)
	candidateScore := completeness(candidate)
	if candidateScore > currentScore {
		return candidate
	}
	return current
}

func completeness(event domain.Event) int {
	values := []string{event.Title, event.Venue, event.Area, event.Category, event.Summary, event.Price, event.SourceURL, event.ImageURL}
	score := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			score++
		}
	}
	return score
}
