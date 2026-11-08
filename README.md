# MCR Radar

MCR Radar is a personal Manchester event discovery service built in Go. It collects listings from multiple providers, converts them into one consistent format, removes duplicate events and presents the useful results through a responsive frontend.

The project grew from a familiar problem: deciding what to do at the weekend usually means searching several ticket sites, venue calendars and social feeds. MCR Radar aims to replace that with one small, useful shortlist.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Dependencies](https://img.shields.io/badge/runtime_dependencies-0-171716)
![License](https://img.shields.io/badge/license-MIT-e84d36)

## Current status

The first working version includes:

- A responsive event discovery frontend served from the Go binary
- Concurrent collectors for Ticketmaster and Skiddle
- Provider-independent event normalisation
- Deterministic duplicate detection across sources
- A free Open-Meteo weekend forecast
- Search, date and category filters
- A random “pick one for me” action
- A saved-event shortlist stored locally in the browser
- A demo mode that works without credentials
- Graceful degradation when an event or weather provider is unavailable
- Reduced-motion support, keyboard navigation and accessible dialog behaviour
- Continuous integration and a small multi-stage container image

The application currently keeps its refreshed event catalogue in memory. Persistent preferences and scheduled refreshes are planned for a later version.

## What it looks like

The interface is intentionally closer to a local culture guide than an administration dashboard. It uses event cards, quick filters and a small weather outlook to answer three questions quickly:

1. What is happening?
2. Is it relevant to me?
3. Is the weather likely to change the plan?

When API keys are not configured, the application labels itself as a preview and uses clearly marked demonstration events. This makes frontend work and automated testing possible without committing secrets or depending on third-party availability.

## How it works

```text
Ticketmaster ─┐
              ├── concurrent collectors ── normalise ── deduplicate ── catalogue
Skiddle ──────┘                                                       │
                                                                      ├── /api/events
Open-Meteo ─────────────────────── cached forecast ───────────────────┤
                                                                      └── Go-served frontend
```

At startup, every configured event source is called concurrently. Each provider adapter maps its response into the same `domain.Event` structure. The catalogue then generates a stable comparison key from the normalised event title, venue and calendar date.

If two providers return the same event, MCR Radar retains the more complete record. A provider failure becomes a warning rather than a full application failure. If no live provider is configured or every provider fails, the catalogue uses demonstration data.

The frontend assets are embedded with Go's `embed` package, so deployment produces one binary with no Node.js runtime and no frontend build step.

## Free data sources

### Ticketmaster Discovery API

Ticketmaster supplies larger concerts, theatre, comedy and sporting events. A free developer key is required.

- API: `https://app.ticketmaster.com/discovery/v2/events.json`
- Environment variable: `TICKETMASTER_API_KEY`
- Documentation: <https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/>

### Skiddle Events API

Skiddle adds local music, club, comedy, arts and nightlife listings. A free API key is available for approved uses.

- API: `https://www.skiddle.com/api/v1/events/search/`
- Environment variable: `SKIDDLE_API_KEY`
- Documentation: <https://www.skiddle.com/api/>

Skiddle's terms require visible attribution, its supplied event links and official branding. The interface already retains the source link and attribution. Before publishing live Skiddle data, add the current official logo asset and verify the latest provider terms.

### Open-Meteo

Open-Meteo provides Manchester's daily forecast without an API key for non-commercial use. Results are cached in memory for 30 minutes.

- API: `https://api.open-meteo.com/v1/forecast`
- Documentation: <https://open-meteo.com/en/docs>

Open-Meteo requires attribution under its data licence; this is shown beside the forecast after a successful request.

## Run locally

Requirements:

- Go 1.23 or newer
- Optional free Ticketmaster and Skiddle API keys

Clone and run:

```bash
git clone https://github.com/Armoooo/mcr-radar.git
cd mcr-radar
go run ./cmd/mcr-radar
```

Open <http://localhost:8080>. Without API keys, the application starts in demo mode.

To use live sources, copy `.env.example` to `.env` and add the keys:

```bash
TICKETMASTER_API_KEY=your-key
SKIDDLE_API_KEY=your-key
```

Then start the application from the repository root:

```bash
go run ./cmd/mcr-radar
```

The local `.env` file is loaded automatically, is ignored by Git and must never be committed. Real process environment variables take precedence when the application is deployed.

## Configuration

API keys are intentionally not included in this repository. `.env.example` contains blank placeholders so each developer can supply their own free provider credentials. The application remains fully usable in labelled demo mode when no keys are configured.

For a local live-data setup:

```bash
cp .env.example .env
```

Add your own values to `.env`. That file is ignored by Git and should remain local.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `PORT` | No | `8080` | HTTP listening port |
| `TICKETMASTER_API_KEY` | No | — | Enables Ticketmaster collection |
| `SKIDDLE_API_KEY` | No | — | Enables Skiddle collection |

## HTTP endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Embedded responsive frontend |
| `GET /api/events` | Current normalised event snapshot |
| `GET /api/weather` | Cached Manchester forecast |
| `GET /healthz` | Lightweight health check |

Example event response:

```json
{
  "events": [
    {
      "id": "ticketmaster-example",
      "title": "Example event",
      "venue": "Example venue",
      "area": "Manchester",
      "category": "Music",
      "startsAt": "2026-11-14T19:30:00Z",
      "price": "£15–25",
      "source": "Ticketmaster",
      "sourceUrl": "https://example.com/event"
    }
  ],
  "updatedAt": "2026-11-08T10:00:00Z",
  "mode": "live"
}
```

## Project structure

```text
cmd/mcr-radar/       application entry point and graceful shutdown
internal/catalog/    concurrent collection, snapshots and deduplication
internal/domain/     provider-independent event and weather types
internal/httpserver/ HTTP routes, middleware and security headers
internal/source/     Ticketmaster, Skiddle and demo adapters
internal/weather/    cached Open-Meteo client
web/                 embedded HTML, CSS and browser JavaScript
```

## Engineering decisions

### Standard library first

The service currently has no third-party Go dependencies. `net/http`, `encoding/json`, `embed` and the standard concurrency primitives cover the first version well. This keeps the binary small and makes the important behaviour easy to inspect.

### Concurrent, isolated providers

Providers are independent and execute concurrently. Results are merged only after every source has returned or the refresh context has expired. Adding another source means implementing one small interface rather than changing the catalogue or frontend.

```go
type Source interface {
    Name() string
    Fetch(context.Context, time.Time, time.Time) ([]domain.Event, error)
}
```

### Predictable deduplication

Deduplication intentionally starts simple. Punctuation and letter case are removed from the title and venue, then the calendar date is appended. This avoids hiding separate performances at the same venue while catching the common case where two ticket providers list the same event with slightly different punctuation.

A future version can add fuzzy title matching and shared artist identifiers while retaining deterministic database uniqueness rules.

### Failure is partial

A temporary provider failure should reduce coverage, not produce a blank website. The catalogue records provider warnings, keeps successful results and falls back to preview data only when no live events are available.

## Tests and checks

Run the complete local check:

```bash
make check
```

This formats Go source, runs `go vet` and executes the test suite. Current catalogue tests cover duplicate selection and the important boundary where identical events on different dates must remain separate.

## Roadmap

- Store events, preferences and refresh history in PostgreSQL
- Add scheduled background refreshes instead of startup-only collection
- Add idempotent database upserts using provider IDs and event fingerprints
- Rank events using explicit interests and saved/dismissed history
- Add venue-level feeds for smaller independent Manchester events
- Add a weekly personal digest
- Add integration tests using recorded provider responses
- Add metrics for provider latency, failures and duplicate rates
- Deploy the container on a free or low-cost service once live data is enabled

## Responsible use

MCR Radar links visitors back to the original provider for full details and ticket purchases. Every data source has its own licence, attribution and branding rules. Check those terms before deploying a public or commercial version, and do not scrape websites that have not granted permission.

## License

The application source is available under the [MIT License](LICENSE). Event listings, provider artwork and weather data remain subject to their respective source licences.
