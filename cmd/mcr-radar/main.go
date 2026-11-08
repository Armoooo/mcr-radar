package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Armoooo/mcr-radar/internal/catalog"
	"github.com/Armoooo/mcr-radar/internal/config"
	"github.com/Armoooo/mcr-radar/internal/httpserver"
	"github.com/Armoooo/mcr-radar/internal/source"
	"github.com/Armoooo/mcr-radar/internal/weather"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := config.LoadEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Warn("local environment file could not be loaded", "error", err)
	}
	providers := configuredSources()
	events := catalog.New(providers, source.Demo{})

	refreshCtx, cancelRefresh := context.WithTimeout(context.Background(), 18*time.Second)
	snapshot := events.Refresh(refreshCtx, time.Now(), time.Now().AddDate(0, 1, 0))
	cancelRefresh()
	logger.Info("event catalog ready", "mode", snapshot.Mode, "events", len(snapshot.Events), "providers", len(providers))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpserver.New(events, &weather.Client{}, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown", "error", err)
		}
	}()

	logger.Info("MCR Radar listening", "url", fmt.Sprintf("http://localhost:%s", port))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func configuredSources() []source.Source {
	providers := make([]source.Source, 0, 2)
	if key := os.Getenv("TICKETMASTER_API_KEY"); key != "" {
		providers = append(providers, source.Ticketmaster{APIKey: key})
	}
	if key := os.Getenv("SKIDDLE_API_KEY"); key != "" {
		providers = append(providers, source.Skiddle{APIKey: key})
	}
	return providers
}
