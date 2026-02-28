package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// Pool manages shared Chrome instances
type Pool struct {
	sem chan struct{}
}

// NewPool creates a browser pool
func NewPool(maxConcurrent int) *Pool {
	return &Pool{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// WithContext runs a task using a stealth Chrome instance
func (p *Pool) WithContext(fn func(ctx context.Context) error) error {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],

		// MUST be headful for Naukri
		chromedp.Flag("headless", false),

		// Stealth hardening
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),

		// Stability
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Let Chrome fully bootstrap
	if err := chromedp.Run(
		ctx,
		chromedp.Sleep(500*time.Millisecond),
		stealthJS(), // ✅ NOW DEFINED BELOW
	); err != nil {
		return err
	}

	return fn(ctx)
}

/* ---------------- STEALTH PATCH ---------------- */

func stealthJS() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		script := `
		// Hide webdriver flag
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// Fake chrome runtime
		window.chrome = { runtime: {} };

		// Permissions fix
		const originalQuery = navigator.permissions.query;
		navigator.permissions.query = (parameters) =>
			parameters.name === 'notifications'
				? Promise.resolve({ state: Notification.permission })
				: originalQuery(parameters);

		// Fake plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5]
		});

		// Fake languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});
		`
		return chromedp.Evaluate(script, nil).Do(ctx)
	})
}
