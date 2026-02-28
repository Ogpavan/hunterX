package naukri

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"hunterX/internal/browser"
	"hunterX/internal/models"
	"hunterX/internal/scrapers"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

/* =========================
   CONFIG
========================= */

// toggle this to true when debugging
const debugMode = true

func logInfo(msg string, args ...any) {
	fmt.Printf("[naukri] "+msg+"\n", args...)
}

func logWarn(msg string, args ...any) {
	fmt.Printf("[naukri][WARN] "+msg+"\n", args...)
}

func logDebug(msg string, args ...any) {
	if debugMode {
		fmt.Printf("[naukri][DEBUG] "+msg+"\n", args...)
	}
}

/* =========================
   SCRAPER
========================= */

type NaukriScraper struct {
	Browser *browser.Pool
}

func New(browser *browser.Pool) *NaukriScraper {
	return &NaukriScraper{Browser: browser}
}

func (n *NaukriScraper) Name() string {
	return "naukri"
}

/* =========================
   API MODELS
========================= */

type apiResponse struct {
	JobDetails []struct {
		CompanyId int    `json:"companyId"`
		Company   string `json:"companyName"`
		Title     string `json:"title"`
		Created   int64  `json:"createdDate"`
		StaticUrl string `json:"staticUrl"`
		Placeholders []struct {
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"placeholders"`
	} `json:"jobDetails"`
}

/* =========================
   FETCH
========================= */

func (n *NaukriScraper) FetchRecruiters(
	ctx context.Context,
	input scrapers.SearchInput,
) ([]models.Recruiter, error) {

	logInfo("Initializing")
	logInfo("City=%s | MaxPages=%d", input.City, input.MaxPages)

	var recruiters []models.Recruiter

	err := n.Browser.WithContext(func(c context.Context) error {

		searchURL := fmt.Sprintf(
			"https://www.naukri.com/jobs-in-%s",
			url.QueryEscape(input.City),
		)

		var captured []byte
		apiCaptured := false

		/* ---------- CAPTURE SEARCH API ---------- */

		chromedp.ListenTarget(c, func(ev interface{}) {
			if apiCaptured {
				return
			}
			if e, ok := ev.(*network.EventResponseReceived); ok {
				if strings.Contains(e.Response.URL, "/jobapi/v3/search") {
					apiCaptured = true
					go func(id network.RequestID) {
						_ = chromedp.Run(c, chromedp.ActionFunc(func(ctx context.Context) error {
							body, err := network.GetResponseBody(id).Do(ctx)
							if err == nil {
								captured = body
							}
							return nil
						}))
					}(e.RequestID)
				}
			}
		})

		logInfo("Opening search page")
		if err := chromedp.Run(
			c,
			network.Enable(),
			chromedp.Navigate(searchURL),
			chromedp.Sleep(7*time.Second),
		); err != nil {
			return err
		}

		if len(captured) == 0 {
			return fmt.Errorf("naukri api response not captured")
		}

		logInfo("Search API captured")

		var api apiResponse
		_ = json.Unmarshal(captured, &api)

		/* ---------- OPTIONAL FULL RESPONSE SAVE ---------- */

		if debugMode {
			saveJSON(
				"output/naukri_complete_search_"+safe(input.City)+".json",
				captured,
			)
			logDebug("Saved complete search response")
		}

		/* ---------- COMPANY MAP ---------- */

		type companyInfo struct {
			Name      string
			Id        int
			StaticUrl string
		}

		companyMap := make(map[string]companyInfo)

		for _, j := range api.JobDetails {
			if j.Company != "" && j.CompanyId != 0 && j.StaticUrl != "" {
				companyMap[j.Company] = companyInfo{
					Name:      j.Company,
					Id:        j.CompanyId,
					StaticUrl: j.StaticUrl,
				}
			}
		}

		logInfo("Found %d companies", len(companyMap))

		/* ---------- COMPANY JOBS ---------- */

		i := 0
		for _, comp := range companyMap {
			i++
			logInfo("Processing companies: %d/%d", i, len(companyMap))

			var companyAPI []byte
			capturedCompany := false

			chromedp.ListenTarget(c, func(ev interface{}) {
				if capturedCompany {
					return
				}
				if e, ok := ev.(*network.EventResponseReceived); ok {
					if strings.Contains(e.Response.URL, "/jobapi/v3/search") {
						capturedCompany = true
						go func(id network.RequestID) {
							_ = chromedp.Run(c, chromedp.ActionFunc(func(ctx context.Context) error {
								body, err := network.GetResponseBody(id).Do(ctx)
								if err == nil {
									companyAPI = body
								}
								return nil
							}))
						}(e.RequestID)
					}
				}
			})

			_ = chromedp.Run(
				c,
				chromedp.Navigate("https://www.naukri.com/"+comp.StaticUrl),
				chromedp.Sleep(6*time.Second),
			)

			if len(companyAPI) == 0 {
				logWarn("No jobs captured for %s", comp.Name)
				continue
			}

			if debugMode {
				saveJSON(
					"output/naukri_summary_"+safe(comp.Name)+".json",
					companyAPI,
				)
				logDebug("Saved company summary: %s", comp.Name)
			}
		}

		/* ---------- RECRUITERS ---------- */

		seen := make(map[string]bool)

		for _, j := range api.JobDetails {
			if seen[j.Company] {
				continue
			}
			seen[j.Company] = true

			location := ""
			for _, p := range j.Placeholders {
				if p.Type == "location" {
					location = p.Label
					break
				}
			}

			recruiters = append(recruiters, models.Recruiter{
				Source:      "naukri",
				SourceID:    fmt.Sprint(j.CompanyId),
				CompanyName: strings.TrimSpace(j.Company),
				JobTitle:    j.Title,
				Location:    location,
				PostedDate:  time.Unix(j.Created/1000, 0),
				CollectedAt: time.Now(),
			})
		}

		return nil
	})

	logInfo("Completed")
	return recruiters, err
}

/* =========================
   HELPERS
========================= */

func safe(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

func saveJSON(filename string, raw []byte) {
	_ = os.MkdirAll("output", 0755)

	var v any
	if json.Unmarshal(raw, &v) != nil {
		return
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	_ = os.WriteFile(filename, b, 0644)
}
