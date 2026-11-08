package source

import (
	"context"
	"time"

	"github.com/Armoooo/mcr-radar/internal/domain"
)

// Source knows how to retrieve and normalise events from one provider.
type Source interface {
	Name() string
	Fetch(context.Context, time.Time, time.Time) ([]domain.Event, error)
}
