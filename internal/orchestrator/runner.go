package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hunterX/internal/enrich"
	"hunterX/internal/models"
	"hunterX/internal/scrapers"
)

func RunScrapers(
	ctx context.Context,
	input scrapers.SearchInput,
	scraperList []scrapers.Scraper,
) []models.Recruiter {

	var all []models.Recruiter
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, s := range scraperList {
		wg.Add(1)

		go func(sc scrapers.Scraper) {
			defer wg.Done()

			data, err := sc.FetchRecruiters(ctx, input)
			if err != nil {
				fmt.Println("[orchestrator] scraper error:", err)
				return
			}

			mu.Lock()
			all = append(all, data...)
			mu.Unlock()
		}(s)
	}

	wg.Wait()

	for i := range all {
		all[i].CollectedAt = time.Now()
	}

	return all
}

// ✅ FIXED: lifecycle-aware enrichment
func RunEnrichment(
	ctx context.Context,
	recruiters []models.Recruiter,
	enrichers []enrich.Enricher,
) {
	for _, e := range enrichers {
		if err := e.Init(ctx); err != nil {
			fmt.Println("[enrich] init failed:", e.Name(), err)
			return
		}
		defer e.Close()
	}

	for i := range recruiters {
		for _, e := range enrichers {
			_ = e.Enrich(&recruiters[i])
		}
	}
}
