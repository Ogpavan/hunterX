package enrich

import (
	"context"

	"hunterX/internal/models"
)

// Enricher defines lifecycle-based enrichment
type Enricher interface {
	Name() string

	// Init is called once before enrichment starts
	Init(ctx context.Context) error

	// Enrich enriches a single recruiter (reused context)
	Enrich(r *models.Recruiter) error

	// Close is called once after enrichment finishes
	Close() error
}
