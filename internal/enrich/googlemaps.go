package enrich

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"fmt"
	"hunterX/internal/models"
)

/* =========================
   CONFIG
========================= */

// turn ON only when debugging
const debugMode = false

func logInfo(msg string, args ...any) {
	println("[google_maps] " + sprintf(msg, args...))
}

func logWarn(msg string, args ...any) {
	println("[google_maps][WARN] " + sprintf(msg, args...))
}

func logDebug(msg string, args ...any) {
	if debugMode {
		println("[google_maps][DEBUG] " + sprintf(msg, args...))
	}
}

func sprintf(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

/* =========================
   ENRICHER
========================= */

type GoogleMapsEnricher struct {
	client *http.Client
}

func (g *GoogleMapsEnricher) Name() string {
	return "google_maps"
}

// Init keeps same interface
func (g *GoogleMapsEnricher) Init(parent context.Context) error {
	g.client = &http.Client{
		Timeout: 20 * time.Second,
	}
	logInfo("Initialized")
	return nil
}

// Enrich fills Phone + Website
func (g *GoogleMapsEnricher) Enrich(r *models.Recruiter) error {
	query := strings.TrimSpace(r.CompanyName + " " + r.Location)
	if query == "" {
		return nil
	}

	logDebug("Query=%q", query)

	content, err := g.fetchMapsHTML(query)
	if err != nil {
		logWarn("Request failed for %s", r.CompanyName)
		return err
	}

	phones := extractPhones(content)
	websites := extractWebsites(content)

	if len(phones) > 0 {
		r.Phone = phones[0]
	}
	if len(websites) > 0 {
		r.Website = websites[0]
	}

	if r.Phone != "" || r.Website != "" {
		logInfo(
			"Enriched: %s (Phone=%v Website=%v)",
			r.CompanyName,
			r.Phone != "",
			r.Website != "",
		)
	}

	logDebug("Phones=%d Websites=%d", len(phones), len(websites))
	return nil
}

func (g *GoogleMapsEnricher) Close() error {
	logInfo("Completed")
	return nil
}

/* =========================
   HTTP
========================= */

func (g *GoogleMapsEnricher) fetchMapsHTML(query string) (string, error) {
	baseURL := "https://www.google.com/search"
	params := url.Values{
		"tbm": {"map"},
		"hl":  {"en"},
		"gl":  {"in"},
		"q":   {query},
	}

	req, err := http.NewRequest(
		"GET",
		baseURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

/* =========================
   EXTRACTION
========================= */

func extractPhones(content string) []string {
	re := regexp.MustCompile(`tel:\+?\d{8,15}`)
	raw := re.FindAllString(content, -1)

	seen := make(map[string]struct{})
	var out []string
	for _, p := range raw {
		if _, ok := seen[p]; !ok {
			out = append(out, p)
			seen[p] = struct{}{}
		}
	}
	return out
}

func extractWebsites(content string) []string {
	re := regexp.MustCompile(`https?://[^\s"'<>]+`)
	all := re.FindAllString(content, -1)

	unwanted := []string{
		"google.com",
		"gstatic.com",
		"googleusercontent.com",
	}

	seen := make(map[string]struct{})
	var sites []string

	for _, u := range all {
		clean := cleanupURL(u)
		if clean == "" {
			continue
		}

		skip := false
		for _, bad := range unwanted {
			if strings.Contains(clean, bad) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if _, ok := seen[clean]; !ok {
			sites = append(sites, clean)
			seen[clean] = struct{}{}
		}
	}
	return sites
}

func cleanupURL(u string) string {
	u = strings.TrimSpace(u)

	if idx := strings.IndexAny(u, " \n\t"); idx != -1 {
		u = u[:idx]
	}

	u = regexp.MustCompile(`\\u[0-9a-fA-F]{4,}`).ReplaceAllString(u, "")
	u = strings.TrimRight(u, ".,;!?")

	return u
}
