package scrapers

import (
	"context"

	"hunterX/internal/models"
)

// SearchInput defines recruiter search parameters
type SearchInput struct {
	Keyword  string
	City     string
	MaxPages int
}

// Scraper is implemented by each platform (Naukri, Indeed, etc.)
type Scraper interface {
	// Name returns unique scraper name (naukri, indeed, internshala)
	Name() string

	// FetchRecruiters fetches recruiter/company leads
	FetchRecruiters(
		ctx context.Context,
		input SearchInput,
	) ([]models.Recruiter, error)
}
