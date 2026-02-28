// package enrich

// import (
// 	"context"
// 	"fmt"
// 	"regexp"
// 	"strings"
// 	"time"

// 	"hunterX/internal/models"

// 	"github.com/chromedp/chromedp"
// )

// /* =========================
//    CONFIG
// ========================= */

// const (
// 	chromedpDebugMode = false
// 	maxWebsitesToVisit = 3
// 	searchTimeout = 30 * time.Second
// 	websiteTimeout = 15 * time.Second
// )

// func logChromedpInfo(msg string, args ...any) {
// 	println("[chromedp_enricher] " + fmt.Sprintf(msg, args...))
// }

// func logChromedpWarn(msg string, args ...any) {
// 	println("[chromedp_enricher][WARN] " + fmt.Sprintf(msg, args...))
// }

// func logChromedpDebug(msg string, args ...any) {
// 	if chromedpDebugMode {
// 		println("[chromedp_enricher][DEBUG] " + fmt.Sprintf(msg, args...))
// 	}
// }

// /* =========================
//    CHROMEDP ENRICHER
// ========================= */

// type ChromeDPEnricher struct {
// 	ctx    context.Context
// 	cancel context.CancelFunc
// }

// func (c *ChromeDPEnricher) Name() string {
// 	return "chromedp_googlemaps"
// }

// func (c *ChromeDPEnricher) Init(parent context.Context) error {
// 	logChromedpInfo("Initializing ChromeDP enricher...")

// 	// Setup Chrome options for stealth mode
// 	opts := append(
// 		chromedp.DefaultExecAllocatorOptions[:],
// 		chromedp.Flag("headless", false), // Run in headless mode for production
// 		chromedp.Flag("disable-gpu", true),
// 		chromedp.Flag("no-sandbox", true),
// 		chromedp.Flag("disable-dev-shm-usage", true),
		
// 		// Stealth mode flags
// 		chromedp.Flag("disable-blink-features", "AutomationControlled"),
// 		chromedp.Flag("disable-web-security", true),
// 		chromedp.Flag("disable-features", "VizDisplayCompositor"),
// 		chromedp.Flag("disable-ipc-flooding-protection", true),
// 		chromedp.Flag("disable-renderer-backgrounding", true),
// 		chromedp.Flag("disable-backgrounding-occluded-windows", true),
// 		chromedp.Flag("disable-client-side-phishing-detection", true),
// 		chromedp.Flag("disable-sync", true),
// 		chromedp.Flag("disable-default-apps", true),
// 		chromedp.Flag("disable-extensions", true),
// 		chromedp.Flag("disable-component-extensions-with-background-pages", true),
// 		chromedp.Flag("no-default-browser-check", true),
// 		chromedp.Flag("no-first-run", true),
// 		chromedp.Flag("no-pings", true),
// 		chromedp.Flag("no-zygote", true),
// 		chromedp.Flag("disable-infobars", true),
// 		chromedp.Flag("disable-notifications", true),
// 		chromedp.Flag("disable-popup-blocking", true),
		
// 		// Additional anti-detection flags
// 		chromedp.Flag("exclude-switches", "enable-automation"),
// 		chromedp.Flag("disable-automation", true),
// 		chromedp.Flag("disable-extensions-file-access-check", true),
// 		chromedp.Flag("disable-extensions-http-throttling", true),
// 		chromedp.Flag("disable-background-timer-throttling", true),
// 		chromedp.Flag("disable-background-networking", true),
// 		chromedp.Flag("disable-add-to-shelf", true),
// 		chromedp.Flag("disable-breakpad", true),
// 		chromedp.Flag("disable-logging", true),
// 		chromedp.Flag("disable-plugins", true),
// 		chromedp.Flag("disable-plugins-discovery", true),
// 		chromedp.Flag("disable-preconnect", true),
		
// 		// Set a realistic user agent
// 		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
// 	)

// 	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
// 	c.ctx, c.cancel = chromedp.NewContext(allocCtx)

// 	// Store the allocator cancel function for cleanup
// 	go func() {
// 		<-c.ctx.Done()
// 		allocCancel()
// 	}()

// 	// Initialize browser with anti-detection scripts
// 	err := chromedp.Run(c.ctx,
// 		chromedp.Navigate("https://google.com"),
// 		chromedp.Sleep(2*time.Second),
// 		chromedp.Evaluate(`
// 			// Override webdriver detection
// 			Object.defineProperty(navigator, 'webdriver', {
// 				get: () => undefined,
// 			});
			
// 			// Remove automation indicators safely
// 			try {
// 				if (window.chrome && window.chrome.runtime) {
// 					delete window.chrome.runtime.onConnect;
// 					delete window.chrome.runtime.onMessage;
// 				}
// 			} catch(e) {}
			
// 			// Override plugins
// 			Object.defineProperty(navigator, 'plugins', {
// 				get: () => [1, 2, 3, 4, 5],
// 			});
			
// 			// Override languages
// 			Object.defineProperty(navigator, 'languages', {
// 				get: () => ['en-US', 'en'],
// 			});
			
// 			// Set chrome object safely
// 			if (!window.chrome) {
// 				window.chrome = {};
// 			}
// 			if (!window.chrome.runtime) {
// 				window.chrome.runtime = {};
// 			}
// 		`, nil),
// 	)

// 	if err != nil {
// 		c.Close()
// 		return fmt.Errorf("failed to initialize ChromeDP: %v", err)
// 	}

// 	logChromedpInfo("ChromeDP enricher initialized successfully")
// 	return nil
// }

// func (c *ChromeDPEnricher) Enrich(r *models.Recruiter) error {
// 	if r.CompanyName == "" {
// 		return nil
// 	}

// 	query := strings.TrimSpace(r.CompanyName + " " + r.Location)
// 	logChromedpDebug("Enriching: %s with query: %s", r.CompanyName, query)

// 	// Create context with timeout for this operation
// 	ctx, cancel := context.WithTimeout(c.ctx, searchTimeout)
// 	defer cancel()

// 	// Perform Google search
// 	err := c.performSearch(ctx, query)
// 	if err != nil {
// 		logChromedpWarn("Search failed for %s: %v", r.CompanyName, err)
// 		return err
// 	}

// 	// Extract contact info from search results
// 	phones, websites, err := c.extractContactInfoFromSearch(ctx)
// 	if err != nil {
// 		logChromedpWarn("Contact extraction failed for %s: %v", r.CompanyName, err)
// 		return err
// 	}

// 	// Update recruiter with basic info
// 	if len(phones) > 0 && r.Phone == "" {
// 		r.Phone = phones[0]
// 		logChromedpDebug("Found phone for %s: %s", r.CompanyName, r.Phone)
// 	}
// 	if len(websites) > 0 && r.Website == "" {
// 		r.Website = websites[0]
// 		logChromedpDebug("Found website for %s: %s", r.CompanyName, r.Website)
// 	}

// 	// Visit websites for additional contact info (limited to avoid timeouts)
// 	if len(websites) > 0 {
// 		additionalEmails, additionalPhones := c.visitWebsites(ctx, websites[:min(len(websites), maxWebsitesToVisit)])
		
// 		// Update phone if we found a better one from website
// 		if len(additionalPhones) > 0 && r.Phone == "" {
// 			r.Phone = additionalPhones[0]
// 			logChromedpDebug("Found additional phone for %s: %s", r.CompanyName, r.Phone)
// 		}

// 		// Log additional contact info found
// 		if len(additionalEmails) > 0 {
// 			logChromedpDebug("Found %d additional emails for %s", len(additionalEmails), r.CompanyName)
// 		}
// 	}

// 	if r.Phone != "" || r.Website != "" {
// 		logChromedpInfo("Enriched %s: Phone=%v Website=%v", r.CompanyName, r.Phone != "", r.Website != "")
// 	}

// 	return nil
// }

// func (c *ChromeDPEnricher) Close() error {
// 	if c.cancel != nil {
// 		c.cancel()
// 	}
// 	logChromedpInfo("ChromeDP enricher closed")
// 	return nil
// }

// /* =========================
//    SEARCH OPERATIONS
// ========================= */

// func (c *ChromeDPEnricher) performSearch(ctx context.Context, query string) error {
// 	return chromedp.Run(ctx,
// 		// Wait for search box to be visible
// 		chromedp.WaitVisible(`textarea[name="q"]`, chromedp.ByQuery),
// 		chromedp.Click(`textarea[name="q"]`, chromedp.ByQuery),
		
// 		// Clear the search box and type new query
// 		chromedp.Evaluate(`document.querySelector('textarea[name="q"]').value = ''`, nil),
// 		chromedp.SendKeys(`textarea[name="q"]`, query, chromedp.ByQuery),
// 		chromedp.Sleep(500*time.Millisecond),
		
// 		// Press Enter to search
// 		chromedp.KeyEvent("\r"),
// 		chromedp.Sleep(3*time.Second),
		
// 		// Wait for search results
// 		chromedp.WaitVisible(`#search`, chromedp.ByQuery),
// 	)
// }

// func (c *ChromeDPEnricher) extractContactInfoFromSearch(ctx context.Context) ([]string, []string, error) {
// 	var pageContent string
// 	err := chromedp.Run(ctx,
// 		chromedp.InnerHTML("body", &pageContent, chromedp.ByQuery),
// 	)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to get page content: %v", err)
// 	}

// 	phones, websites := c.extractContactInfo(pageContent)
// 	return phones, websites, nil
// }

// /* =========================
//    CONTENT EXTRACTION
// ========================= */

// func (c *ChromeDPEnricher) extractContactInfo(content string) ([]string, []string) {
// 	// Phone extraction with aria-label pattern
// 	ariaLabelPhoneRegex := regexp.MustCompile(`aria-label="(?:[Cc]all\s+)?(?:[Pp]hone\s+(?:number\s+)?)?([+\d\s\-().]+)"`)
	
// 	// Website extraction from Google search results
// 	websiteLinkRegex := regexp.MustCompile(`<a[^>]*class="[^"]*n1obkb[^"]*"[^>]*href="([^"]+)"`)

// 	// Extract phones
// 	var phones []string
// 	phoneMatches := ariaLabelPhoneRegex.FindAllStringSubmatch(content, -1)
// 	phoneMap := make(map[string]bool)
	
// 	for _, match := range phoneMatches {
// 		if len(match) >= 2 {
// 			phone := strings.TrimSpace(match[1])
// 			cleanPhone := c.cleanPhoneNumber(phone)
// 			if len(cleanPhone) >= 7 && !phoneMap[cleanPhone] {
// 				phones = append(phones, cleanPhone)
// 				phoneMap[cleanPhone] = true
// 			}
// 		}
// 	}

// 	// Extract websites
// 	var websites []string
// 	websiteMatches := websiteLinkRegex.FindAllStringSubmatch(content, -1)
// 	websiteMap := make(map[string]bool)
	
// 	for _, match := range websiteMatches {
// 		if len(match) >= 2 {
// 			fullURL := match[1]
// 			domainMatch := regexp.MustCompile(`https?://(?:www\.)?([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`).FindStringSubmatch(fullURL)
// 			if len(domainMatch) >= 2 {
// 				domain := domainMatch[1]
// 				if !c.isCommonDomain(domain) && !websiteMap[domain] {
// 					websites = append(websites, domain)
// 					websiteMap[domain] = true
// 				}
// 			}
// 		}
// 	}

// 	return phones, websites
// }

// /* =========================
//    WEBSITE VISITING
// ========================= */

// func (c *ChromeDPEnricher) visitWebsites(ctx context.Context, websites []string) ([]string, []string) {
// 	var allEmails []string
// 	var allPhones []string
	
// 	emailMap := make(map[string]bool)
// 	phoneMap := make(map[string]bool)
	
// 	for _, website := range websites {
// 		if website == "" {
// 			continue
// 		}
		
// 		// Create timeout context for this website
// 		websiteCtx, websiteCancel := context.WithTimeout(ctx, websiteTimeout)
		
// 		emails, phones := c.visitSingleWebsite(websiteCtx, website)
		
// 		// Add unique emails
// 		for _, email := range emails {
// 			if !emailMap[email] {
// 				allEmails = append(allEmails, email)
// 				emailMap[email] = true
// 			}
// 		}
		
// 		// Add unique phones
// 		for _, phone := range phones {
// 			if !phoneMap[phone] {
// 				allPhones = append(allPhones, phone)
// 				phoneMap[phone] = true
// 			}
// 		}
		
// 		websiteCancel()
		
// 		// Small delay between websites
// 		time.Sleep(1 * time.Second)
// 	}
	
// 	return allEmails, allPhones
// }

// func (c *ChromeDPEnricher) visitSingleWebsite(ctx context.Context, website string) ([]string, []string) {
// 	fullURL := "https://" + website
// 	logChromedpDebug("Visiting: %s", fullURL)
	
// 	var pageContent string
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate(fullURL),
// 		chromedp.Sleep(3*time.Second),
// 		chromedp.InnerHTML("body", &pageContent, chromedp.ByQuery),
// 	)
	
// 	if err != nil {
// 		// Try with www prefix
// 		fullURL = "https://www." + website
// 		logChromedpDebug("Retry with www: %s", fullURL)
		
// 		err = chromedp.Run(ctx,
// 			chromedp.Navigate(fullURL),
// 			chromedp.Sleep(3*time.Second),
// 			chromedp.InnerHTML("body", &pageContent, chromedp.ByQuery),
// 		)
		
// 		if err != nil {
// 			logChromedpDebug("Failed to visit %s: %v", website, err)
// 			return nil, nil
// 		}
// 	}
	
// 	return c.extractWebsiteContactInfo(pageContent)
// }

// /* =========================
//    WEBSITE CONTENT EXTRACTION
// ========================= */

// func (c *ChromeDPEnricher) extractWebsiteContactInfo(content string) ([]string, []string) {
// 	var emails []string
// 	var phones []string
	
// 	// Email extraction
// 	emailRegex := regexp.MustCompile(`(?i)\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
// 	emailMatches := emailRegex.FindAllString(content, -1)
// 	emailMap := make(map[string]bool)
	
// 	for _, email := range emailMatches {
// 		if !emailMap[email] && c.isValidBusinessEmail(email) {
// 			emails = append(emails, email)
// 			emailMap[email] = true
// 		}
// 	}
	
// 	// Phone extraction with multiple patterns
// 	phonePatterns := []*regexp.Regexp{
// 		regexp.MustCompile(`\+91[\s-]?[6-9]\d{9}`),
// 		regexp.MustCompile(`\b0\d{2,4}[\s-]?\d{6,8}\b`),
// 		regexp.MustCompile(`\b[6-9]\d{9}\b`),
// 		regexp.MustCompile(`\b\(?\+?91\)?[\s-]?\(?\d{3,5}\)?[\s-]?\d{3,4}[\s-]?\d{3,4}\b`),
// 	}
	
// 	phoneMap := make(map[string]bool)
// 	for _, pattern := range phonePatterns {
// 		phoneMatches := pattern.FindAllString(content, -1)
// 		for _, phone := range phoneMatches {
// 			cleanPhone := c.cleanPhoneNumber(phone)
// 			if !phoneMap[cleanPhone] && c.isValidIndianPhone(cleanPhone) {
// 				phones = append(phones, cleanPhone)
// 				phoneMap[cleanPhone] = true
// 			}
// 		}
// 	}
	
// 	return emails, phones
// }

// /* =========================
//    UTILITY FUNCTIONS
// ========================= */

// func (c *ChromeDPEnricher) isCommonDomain(domain string) bool {
// 	commonDomains := []string{
// 		"google.com", "youtube.com", "facebook.com", "twitter.com", "instagram.com",
// 		"linkedin.com", "wikipedia.org", "amazon.com", "ebay.com", "reddit.com",
// 		"pinterest.com", "tumblr.com", "snapchat.com", "tiktok.com", "whatsapp.com",
// 		"telegram.org", "discord.com", "skype.com", "zoom.us", "microsoft.com",
// 		"apple.com", "adobe.com", "netflix.com", "spotify.com", "github.com",
// 		"stackoverflow.com", "w3schools.com", "mozilla.org", "chrome.com",
// 	}
	
// 	for _, common := range commonDomains {
// 		if domain == common {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (c *ChromeDPEnricher) cleanPhoneNumber(phone string) string {
// 	cleaned := regexp.MustCompile(`[\s\-\(\)]+`).ReplaceAllString(phone, "")
// 	if regexp.MustCompile(`^\+?91`).MatchString(cleaned) {
// 		cleaned = regexp.MustCompile(`^\+?91`).ReplaceAllString(cleaned, "+91")
// 	}
// 	return cleaned
// }

// func (c *ChromeDPEnricher) isValidIndianPhone(phone string) bool {
// 	clean := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
	
// 	patterns := []string{
// 		`^\+91[6-9]\d{9}$`,
// 		`^91[6-9]\d{9}$`,
// 		`^[6-9]\d{9}$`,
// 		`^0\d{2,4}\d{6,8}$`,
// 	}
	
// 	for _, pattern := range patterns {
// 		if matched, _ := regexp.MatchString(pattern, clean); matched {
// 			return len(clean) >= 10 && len(clean) <= 13
// 		}
// 	}
	
// 	return false
// }

// func (c *ChromeDPEnricher) isValidBusinessEmail(email string) bool {
// 	commonDomains := []string{
// 		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "live.com",
// 		"aol.com", "icloud.com", "protonmail.com", "tutanota.com",
// 	}
	
// 	for _, domain := range commonDomains {
// 		if regexp.MustCompile(`(?i)@`+regexp.QuoteMeta(domain)+`$`).MatchString(email) {
// 			return false
// 		}
// 	}
// 	return true
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }



package enrich

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"hunterX/internal/models"

	"github.com/chromedp/chromedp"
)

/* =========================
   CONFIG
========================= */

const (
	chromedpDebugMode = false

	maxWebsitesToVisit = 3

	searchTimeout  = 30 * time.Second
	websiteTimeout = 15 * time.Second
)

func logChromedpInfo(msg string, args ...any) {
	println("[chromedp_enricher] " + fmt.Sprintf(msg, args...))
}

func logChromedpWarn(msg string, args ...any) {
	println("[chromedp_enricher][WARN] " + fmt.Sprintf(msg, args...))
}

func logChromedpDebug(msg string, args ...any) {
	if chromedpDebugMode {
		println("[chromedp_enricher][DEBUG] " + fmt.Sprintf(msg, args...))
	}
}

/* =========================
   ENRICHER
========================= */

type ChromeDPEnricher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *ChromeDPEnricher) Name() string {
	return "chromedp_googlemaps"
}

func (c *ChromeDPEnricher) Init(parent context.Context) error {
	logChromedpInfo("Initializing ChromeDP enricher...")

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
	c.ctx, c.cancel = chromedp.NewContext(allocCtx)

	go func() {
		<-c.ctx.Done()
		allocCancel()
	}()

	err := chromedp.Run(
		c.ctx,
		chromedp.Navigate("https://google.com"),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
			Object.defineProperty(navigator, 'plugins', { get: () => [1,2,3] });
		`, nil),
	)

	if err != nil {
		c.Close()
		return err
	}

	logChromedpInfo("ChromeDP enricher initialized")
	return nil
}

func (c *ChromeDPEnricher) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

/* =========================
   ENRICH
========================= */

func (c *ChromeDPEnricher) Enrich(r *models.Recruiter) error {
	if r.CompanyName == "" {
		return nil
	}

	query := strings.TrimSpace(r.CompanyName + " " + r.Location)

	ctx, cancel := context.WithTimeout(c.ctx, searchTimeout)
	defer cancel()

	if err := c.performSearch(ctx, query); err != nil {
		logChromedpWarn("Search failed: %v", err)
		return nil
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.InnerHTML("body", &body)); err != nil {
		return nil
	}

	phones, websites := extractGoogleContactInfo(body)

	if r.Phone == "" && len(phones) > 0 {
		r.Phone = phones[0]
	}
	if r.Website == "" && len(websites) > 0 {
		r.Website = websites[0]
	}

	if len(websites) > 0 {
		emails, sitePhones := c.visitWebsites(ctx, websites[:min(len(websites), maxWebsitesToVisit)])
		if r.Phone == "" && len(sitePhones) > 0 {
			r.Phone = sitePhones[0]
		}
		if len(emails) > 0 {
			logChromedpDebug("Found %d emails", len(emails))
		}
	}

	return nil
}

/* =========================
   SEARCH
========================= */

func (c *ChromeDPEnricher) performSearch(ctx context.Context, query string) error {
	return chromedp.Run(ctx,
		chromedp.WaitVisible(`textarea[name="q"]`),
		chromedp.Evaluate(`document.querySelector('textarea[name="q"]').value = ''`, nil),
		chromedp.SendKeys(`textarea[name="q"]`, query),
		chromedp.KeyEvent("\r"),
		chromedp.WaitVisible(`#search`),
	)
}

/* =========================
   GOOGLE EXTRACTION (NEW)
========================= */

func extractGoogleContactInfo(html string) ([]string, []string) {
	phones := make(map[string]bool)
	websites := make(map[string]bool)

	ariaPhone := regexp.MustCompile(`aria-label="(?:Call\s+)?(?:Phone\s+number\s+)?([+\d\s().-]+)"`)
	for _, m := range ariaPhone.FindAllStringSubmatch(html, -1) {
		p := cleanPhone(m[1])
		if isValidIndianPhone(p) {
			phones[p] = true
		}
	}

	siteRegex := regexp.MustCompile(`<a[^>]+href="(https?://[^"]+)"`)
	for _, m := range siteRegex.FindAllStringSubmatch(html, -1) {
		if domain := extractDomain(m[1]); domain != "" && !isCommonDomain(domain) {
			websites[domain] = true
		}
	}

	return mapKeys(phones), mapKeys(websites)
}

/* =========================
   WEBSITE VISIT (NEW)
========================= */

func (c *ChromeDPEnricher) visitWebsites(ctx context.Context, sites []string) ([]string, []string) {
	emails := map[string]bool{}
	phones := map[string]bool{}

	for _, s := range sites {
		for _, url := range []string{"https://" + s, "https://www." + s} {
			wctx, cancel := context.WithTimeout(ctx, websiteTimeout)
			var body string

			err := chromedp.Run(wctx,
				chromedp.Navigate(url),
				chromedp.Sleep(2*time.Second),
				chromedp.InnerHTML("body", &body),
			)
			cancel()

			if err == nil && body != "" {
				e, p := extractWebsiteContacts(body)
				for _, x := range e {
					emails[x] = true
				}
				for _, x := range p {
					phones[x] = true
				}
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	return mapKeys(emails), mapKeys(phones)
}

/* =========================
   WEBSITE EXTRACTION (NEW)
========================= */

func extractWebsiteContacts(html string) ([]string, []string) {
	emails := map[string]bool{}
	phones := map[string]bool{}

	emailRx := regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
	for _, e := range emailRx.FindAllString(html, -1) {
		if isValidBusinessEmail(e) {
			emails[e] = true
		}
	}

	phoneRx := regexp.MustCompile(`\+91[\s-]?[6-9]\d{9}|\b[6-9]\d{9}\b`)
	for _, p := range phoneRx.FindAllString(html, -1) {
		cp := cleanPhone(p)
		if isValidIndianPhone(cp) {
			phones[cp] = true
		}
	}

	return mapKeys(emails), mapKeys(phones)
}

/* =========================
   HELPERS
========================= */

func cleanPhone(p string) string {
	p = regexp.MustCompile(`[\s().-]+`).ReplaceAllString(p, "")
	if strings.HasPrefix(p, "91") && !strings.HasPrefix(p, "+91") {
		p = "+91" + p[2:]
	}
	if strings.HasPrefix(p, "+91") {
		return p
	}
	return p
}

func isValidIndianPhone(p string) bool {
	p = strings.TrimPrefix(p, "+91")
	return regexp.MustCompile(`^[6-9]\d{9}$`).MatchString(p)
}

func isValidBusinessEmail(e string) bool {
	return !regexp.MustCompile(`@(gmail|yahoo|hotmail|outlook|icloud)\.`).MatchString(strings.ToLower(e))
}

func extractDomain(u string) string {
	m := regexp.MustCompile(`https?://(?:www\.)?([^/]+)`).FindStringSubmatch(u)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func isCommonDomain(d string) bool {
	block := []string{"google.com", "facebook.com", "linkedin.com", "youtube.com"}
	for _, b := range block {
		if d == b {
			return true
		}
	}
	return false
}

func mapKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
