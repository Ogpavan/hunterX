# hunterX

`hunterX` is a Go-based recruiter intelligence scraper.
It currently scrapes job/company data from Naukri and enriches results with contact signals (phone/website) using Google search flows.

## Current Status

- Implemented: `Naukri` scraper
- Listed but not active in CLI flow: `Indeed`, `Internshala`
- Output formats: `CSV` and `JSON` in `output/`

## Features

- Interactive CLI workflow
- Platform selection (Naukri active)
- City + max page input
- Enrichment mode selection:
- `Basic Google Maps` (HTTP-based, faster)
- `Advanced ChromeDP` (browser-based, more comprehensive)
- Timestamped output files

## Prerequisites

- Go `1.25.5` (as defined in `go.mod`)
- Google Chrome installed (required by ChromeDP/browser automation paths)
- Internet access to target websites

## Run Locally

```bash
go mod tidy
go run ./cmd/hunterX
```

The CLI will ask for:

1. Platform choice
2. City
3. Max pages
4. Enrichment method
5. Final confirmation

## Build

```bash
go build -o hunterX.exe ./cmd/hunterX
```

## Output

Files are generated in `output/`:

- `recruiters_<timestamp>.csv`
- `recruiters_<timestamp>.json`

CSV columns:

- `Source`
- `CompanyName`
- `RecruiterName`
- `JobTitle`
- `Location`
- `Phone`
- `Email`
- `Website`

JSON includes the full `Recruiter` model fields such as:
- source identifiers
- company/job/location
- enrichment data
- collection timestamps

## Project Structure

```text
cmd/hunterX/            CLI entry point
internal/scrapers/      Platform scrapers
internal/enrich/        Contact enrichment modules
internal/orchestrator/  Scrape + enrich pipeline
internal/output/        CSV/JSON writers
assets/                 Static assets (sample config placeholder)
build/                  Build scripts (placeholder)
```

## Notes and Limitations

- Only Naukri is currently functional in the interactive flow.
- `assets/sample-config.json` and `build/build.ps1` are currently placeholders.
- Web scraping reliability can vary with site layout/network/anti-bot changes.
