package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Armoooo/mcr-radar/internal/catalog"
	"github.com/Armoooo/mcr-radar/internal/weather"
	"github.com/Armoooo/mcr-radar/web"
)

type Server struct {
	catalog *catalog.Catalog
	weather *weather.Client
	logger  *slog.Logger
	static  http.Handler
}

func New(events *catalog.Catalog, forecasts *weather.Client, logger *slog.Logger) http.Handler {
	server := &Server{
		catalog: events,
		weather: forecasts,
		logger:  logger,
		static:  http.FileServer(http.FS(web.Assets)),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events", server.events)
	mux.HandleFunc("GET /api/weather", server.forecasts)
	mux.HandleFunc("GET /healthz", server.health)
	mux.Handle("GET /", server.static)
	return server.securityHeaders(server.accessLog(mux))
}

func (s *Server) events(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.catalog.Snapshot())
}

func (s *Server) forecasts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 9*time.Second)
	defer cancel()
	forecasts, err := s.weather.Forecast(ctx)
	if err != nil {
		s.logger.Warn("weather unavailable", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Weather is temporarily unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"forecasts": forecasts, "source": "Open-Meteo"})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; connect-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
