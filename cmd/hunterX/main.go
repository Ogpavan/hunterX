package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"hunterX/internal/browser"
	"hunterX/internal/enrich"
	"hunterX/internal/orchestrator"
	"hunterX/internal/output"
	"hunterX/internal/scrapers"
	"hunterX/internal/scrapers/naukri"
)

/* =========================
   ANSI COLORS
========================= */

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

var useColor = true

/* =========================
   ANSI SUPPORT (WINDOWS)
========================= */

func initAnsiSupport() {
	// --no-color flag
	for _, a := range os.Args {
		if a == "--no-color" {
			useColor = false
			return
		}
	}

	// Non-Windows always OK
	if runtime.GOOS != "windows" {
		return
	}

	// Enable ANSI in Windows via registry
	cmd := exec.Command(
		"reg",
		"add",
		"HKCU\\Console",
		"/v",
		"VirtualTerminalLevel",
		"/t",
		"REG_DWORD",
		"/d",
		"1",
		"/f",
	)

	if err := cmd.Run(); err != nil {
		useColor = false
	}
}

func c(s string) string {
	if !useColor {
		return ""
	}
	return s
}

/* =========================
   LOGO
========================= */

func printLogo() {
	fmt.Println(c(ColorCyan + ColorBold))
	fmt.Println("██╗  ██╗██╗   ██╗███╗   ██╗████████╗███████╗██████╗ ██╗  ██╗")
	fmt.Println("██║  ██║██║   ██║████╗  ██║╚══██╔══╝██╔════╝██╔══██╗╚██╗██╔╝")
	fmt.Println("███████║██║   ██║██╔██╗ ██║   ██║   █████╗  ██████╔╝ ╚███╔╝ ")
	fmt.Println("██╔══██║██║   ██║██║╚██╗██║   ██║   ██╔══╝  ██╔══██╗ ██╔██╗ ")
	fmt.Println("██║  ██║╚██████╔╝██║ ╚████║   ██║   ███████╗██║  ██║██╔╝ ██╗")
	fmt.Println("╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝")
	fmt.Println(c(ColorReset))
	fmt.Println(c(ColorYellow) + "Recruiter Intelligence Engine" + c(ColorReset))
	fmt.Println("------------------------------------------------------------")
}

/* =========================
   MAIN
========================= */

func main() {
	initAnsiSupport()

	reader := bufio.NewReader(os.Stdin)

	printLogo()

	fmt.Println(c(ColorWhite) + "Interactive CLI Mode\n" + c(ColorReset))

	// -------- STEP 1 --------
	fmt.Println(c(ColorBlue) + "Step 1/3: Select Job Platform" + c(ColorReset))
	fmt.Println("--------------------------------")
	fmt.Println(c(ColorGreen) + "[1] Naukri" + c(ColorReset))
	fmt.Println(c(ColorYellow) + "[2] Indeed (coming soon)" + c(ColorReset))
	fmt.Println(c(ColorYellow) + "[3] Internshala (coming soon)" + c(ColorReset))
	fmt.Println(c(ColorGreen) + "[4] All (only Naukri active)" + c(ColorReset))
	fmt.Print(c(ColorCyan) + "Enter choice [1-4]: " + c(ColorReset))

	siteChoice := readInt(reader, 1)

	if siteChoice != 1 && siteChoice != 4 {
		fmt.Println(c(ColorRed) + "\nOnly Naukri is implemented currently." + c(ColorReset))
		return
	}

	// -------- STEP 2 --------
	fmt.Println(c(ColorBlue) + "\nStep 2/3: Search Filters" + c(ColorReset))
	fmt.Println("--------------------------------")
	fmt.Print(c(ColorCyan) + "City (e.g. Delhi, Bangalore): " + c(ColorReset))
	city := readString(reader)

	fmt.Print(c(ColorCyan) + "Max pages to scan [default: 2]: " + c(ColorReset))
	pages := readInt(reader, 2)

	// -------- ENRICHMENT CHOICE --------
	fmt.Println(c(ColorBlue) + "\nEnrichment Method" + c(ColorReset))
	fmt.Println("--------------------------------")
	fmt.Println(c(ColorGreen) + "[1] Basic Google Maps (HTTP-based, faster)" + c(ColorReset))
	fmt.Println(c(ColorYellow) + "[2] Advanced ChromeDP (Browser-based, more comprehensive)" + c(ColorReset))
	fmt.Print(c(ColorCyan) + "Choose enrichment method [default: 1]: " + c(ColorReset))
	enrichChoice := readInt(reader, 1)

	// -------- STEP 3 --------
	fmt.Println(c(ColorBlue) + "\nStep 3/3: Confirmation" + c(ColorReset))
	fmt.Println("--------------------------------")
	fmt.Println(c(ColorWhite)+"Platform     :"+c(ColorReset), c(ColorGreen)+"Naukri"+c(ColorReset))
	fmt.Println(c(ColorWhite)+"City         :"+c(ColorReset), city)
	fmt.Println(c(ColorWhite)+"Max Pages    :"+c(ColorReset), pages)
	enrichMethodText := "Basic Google Maps"
	if enrichChoice == 2 {
		enrichMethodText = "Advanced ChromeDP"
	}
	fmt.Println(c(ColorWhite)+"Enrichment   :"+c(ColorReset), enrichMethodText)
	fmt.Print(c(ColorYellow) + "\nProceed? (y/n): " + c(ColorReset))

	if !confirm(reader) {
		fmt.Println(c(ColorRed) + "Aborted." + c(ColorReset))
		return
	}

	// -------- INIT --------
	ctx := context.Background()
	browserPool := browser.NewPool(3)

	input := scrapers.SearchInput{
		City:     city,
		MaxPages: pages,
	}

	scraperList := []scrapers.Scraper{
		naukri.New(browserPool),
	}

	fmt.Println(c(ColorPurple) + "\nInitializing browser pool..." + c(ColorReset))
	fmt.Println(c(ColorPurple) + "Scraping started...\n" + c(ColorReset))

	// -------- SCRAPE --------
	recruiters := orchestrator.RunScrapers(ctx, input, scraperList)

	// -------- ENRICH --------
	var enrichers []enrich.Enricher
	
	if enrichChoice == 2 {
		enrichers = []enrich.Enricher{
			&enrich.ChromeDPEnricher{},
		}
		fmt.Println(c(ColorPurple) + "Using Advanced ChromeDP enrichment..." + c(ColorReset))
	} else {
		enrichers = []enrich.Enricher{
			&enrich.GoogleMapsEnricher{},
		}
		fmt.Println(c(ColorPurple) + "Using Basic Google Maps enrichment..." + c(ColorReset))
	}

	orchestrator.RunEnrichment(ctx, recruiters, enrichers)

	fmt.Println(c(ColorGreen) + "[✓] Scraping completed" + c(ColorReset))
	fmt.Println(c(ColorGreen) + "[✓] Enrichment completed" + c(ColorReset))

	fmt.Printf(
		c(ColorGreen)+"\nCollected %d recruiters\n"+c(ColorReset),
		len(recruiters),
	)

	if len(recruiters) == 0 {
		fmt.Println(c(ColorYellow) + "No data collected. Nothing to save." + c(ColorReset))
		return
	}

	// -------- OUTPUT --------
	_ = os.MkdirAll("output", 0755)

	ts := time.Now().Format("20060102_150405")
	csvFile := "output/recruiters_" + ts + ".csv"
	jsonFile := "output/recruiters_" + ts + ".json"

	fmt.Println(c(ColorBlue) + "\nOutput Files" + c(ColorReset))
	fmt.Println("--------------------------------")

	if err := output.SaveCSV(csvFile, recruiters); err != nil {
		fmt.Println(c(ColorRed)+"CSV save error:"+c(ColorReset), err)
	} else {
		fmt.Println(c(ColorGreen)+"CSV  :"+c(ColorReset), csvFile)
	}

	if err := output.SaveJSON(jsonFile, recruiters); err != nil {
		fmt.Println(c(ColorRed)+"JSON save error:"+c(ColorReset), err)
	} else {
		fmt.Println(c(ColorGreen)+"JSON :"+c(ColorReset), jsonFile)
	}

	fmt.Println(c(ColorGreen) + "\nDone." + c(ColorReset))
}

/* =========================
   HELPERS
========================= */

func readString(reader *bufio.Reader) string {
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Print(c(ColorRed) + "Value cannot be empty. Try again: " + c(ColorReset))
	}
}

func readInt(reader *bufio.Reader, defaultVal int) int {
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(input)
	if err != nil || val <= 0 {
		fmt.Print(c(ColorRed) + "Invalid number. Try again: " + c(ColorReset))
		return readInt(reader, defaultVal)
	}
	return val
}

func confirm(reader *bufio.Reader) bool {
	for {
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Print(c(ColorYellow) + "Please enter y or n: " + c(ColorReset))
	}
}
