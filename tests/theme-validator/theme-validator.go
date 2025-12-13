package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/grantfbarnes/card-judge/tests/setup"
	"github.com/grantfbarnes/card-judge/tests/util"
)

var (
	baseURL       = fmt.Sprintf("http://%s:%d", util.DefaultHost, util.DefaultPort)
	loginUsername = util.TestUsername
	loginPassword = util.TestPassword
)

func main() {
	startTime := time.Now()

	// Always run setup: kill dev server and recreate test database
	if err := runSetup(); err != nil {
		log.Fatalf("Setup failed: %v\n", err)
	}

	// Start server
	serverMgr := setup.NewServerManager(baseURL)
	if err := serverMgr.Start(); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
	defer serverMgr.Stop()

	// Initialize browser
	ctx, cancel := initBrowser()
	defer cancel()

	// Login
	if !PerformLogin(ctx, baseURL, loginUsername, loginPassword) {
		log.Fatal("Login failed")
	}

	// Get configurations
	themes := GetThemes()
	pages := GetPageConfigurations()

	// Let user select pages
	pagesToCapture := selectPages(pages)
	if len(pagesToCapture) == 0 {
		log.Fatal("No pages selected")
	}

	// Create output directory
	timestamp := time.Now().Format("20060102_150405")
	outputDir := filepath.Join("./"+util.ScreenshotDir, timestamp)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Capture screenshots
	total := len(themes) * len(pagesToCapture)
	log.Printf("\n📸 SCREENSHOT CAPTURE: %d themes × %d pages = %d total\n", len(themes), len(pagesToCapture), total)
	log.Println(strings.Repeat("─", 80))

	accessibilityResults := captureScreenshots(ctx, baseURL, outputDir, themes, pagesToCapture, total)

	log.Println(strings.Repeat("─", 80))

	// Generate PDFs
	elapsed := time.Since(startTime)
	log.Println("Screenshot capture complete!")
	log.Printf("Screenshots saved to: %s\n", outputDir)
	log.Printf("\n✅ Successfully captured %d screenshots across %d pages and %d themes\n", total, len(pagesToCapture), len(themes))
	log.Printf("⏱️  Completed in %v\n", elapsed)

	log.Println("\n📄 Generating theme PDFs...")
	if err := GenerateThemePDFs("./"+util.ScreenshotDir, timestamp); err != nil {
		log.Printf("⚠️  Warning: PDF generation partially failed: %v\n", err)
	} else {
		log.Println("✅ All PDFs generated successfully!")
	}

	// Generate accessibility report
	log.Println("\n📋 Generating accessibility report...")
	if err := GenerateAccessibilityReport(accessibilityResults, timestamp); err != nil {
		log.Printf("⚠️  Warning: Accessibility report generation failed: %v\n", err)
	} else {
		log.Printf("✅ Accessibility report saved to: theme-reports/%s/accessibility-report.txt\n", timestamp)
	}
}

func runSetup() error {
	log.Println("🔧 Running setup tasks...")

	if err := setup.KillServerOnPort(util.DefaultPort); err != nil {
		log.Printf("Warning: Failed to kill existing server: %v\n", err)
	}

	if err := setup.SetupTestDatabase(); err != nil {
		return fmt.Errorf("failed to set up test database: %w", err)
	}

	log.Println("")
	log.Println("✅ Setup complete! Ready to capture screenshots.")
	return nil
}

func findBrowserPath() string {
	// Windows paths
	if runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Opera\opera.exe`,
			`C:\Program Files\Opera\opera.exe`,
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// macOS paths
	if runtime.GOOS == "darwin" {
		paths := []string{
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Opera.app/Contents/MacOS/Opera",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// Linux paths
	if runtime.GOOS == "linux" {
		commands := []string{"microsoft-edge", "google-chrome", "chromium", "chromium-browser", "opera"}
		for _, cmd := range commands {
			if path, err := exec.LookPath(cmd); err == nil {
				return path
			}
		}
	}

	// Default to let chromedp find it
	return ""
}

func initBrowser() (context.Context, context.CancelFunc) {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	
	// Auto-detect browser path if available
	if browserPath := findBrowserPath(); browserPath != "" {
		log.Printf("Using browser: %s\n", browserPath)
		opts = append(opts, chromedp.ExecPath(browserPath))
	} else {
		log.Println("Using default browser detection")
	}
	
	opts = append(opts,
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-features", "PrivateNetworkAccessRespectPreflightResults"),
		chromedp.Flag("log-level", "3"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, ctxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	// Return a combined cancel function
	return ctx, func() {
		ctxCancel()
		cancel()
	}
}

func selectPages(pages []PageConfig) []PageConfig {
	fmt.Println("\nWhich pages would you like to capture?")
	fmt.Println("0. All pages")
	for i, page := range pages {
		fmt.Printf("%d. %s\n", i+1, page.Name)
	}
	fmt.Print("\nEnter your choice (0 for all): ")

	var choice int
	fmt.Scanln(&choice)

	if choice == 0 {
		return pages
	} else if choice > 0 && choice <= len(pages) {
		return []PageConfig{pages[choice-1]}
	}

	log.Println("Invalid choice.")
	return nil
}

func captureScreenshots(ctx context.Context, baseURL, outputDir string, themes []string, pages []PageConfig, total int) []AccessibilityResult {
	current := 0
	var accessibilityResults []AccessibilityResult

	for _, page := range pages {
		log.Printf("\n📄 Page: %s\n", page.Name)
		log.Println(strings.Repeat("  ", 40))

		pageDir := filepath.Join(outputDir, page.Name)
		if err := os.MkdirAll(pageDir, 0755); err != nil {
			log.Printf("Error creating directory for %s: %v\n", page.Name, err)
			continue
		}

		// Create persistent context for this page
		pageCtx, pageCancel := context.WithTimeout(ctx, 5*time.Minute)

		// Navigate to page
		url := baseURL + page.Path
		log.Printf("  → Navigating to %s...\n", url)
		err := chromedp.Run(pageCtx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Sleep(500*time.Millisecond),
		)

		if err != nil {
			log.Printf("Error navigating to %s: %v\n", page.Name, err)
			pageCancel()
			continue
		}

		// Run page-specific setup
		if page.Setup != nil {
			if err := page.Setup(pageCtx); err != nil {
				log.Printf("Warning: Setup failed for %s: %v\n", page.Name, err)
				pageCancel()
				continue
			}
		}

		// Capture all themes for this page
		for _, theme := range themes {
			current++
			result := captureTheme(pageCtx, pageDir, theme, page.Name, current, total)
			accessibilityResults = append(accessibilityResults, result)
		}

		pageCancel()
	}
	return accessibilityResults
}

func captureTheme(ctx context.Context, pageDir, theme, pageName string, current, total int) AccessibilityResult {
	screenshotTime := time.Now().Format("150405")
	filename := filepath.Join(pageDir, fmt.Sprintf("%s_%s.png", theme, screenshotTime))
	var buf []byte
	result := AccessibilityResult{
		Theme: theme,
		Page:  pageName,
	}

	log.Printf("  [%3d/%d] %5.1f%% %s\n", current, total, float64(current)/float64(total)*100, theme)

	// Apply theme and capture
	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.body.className = '%s'`, theme), nil),
		chromedp.Evaluate(`
			(function() {
				var existing = document.getElementById('debug-url-bar');
				if (existing) existing.remove();
				
				var urlBar = document.createElement('div');
				urlBar.id = 'debug-url-bar';
				urlBar.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#333;color:#0f0;font-family:monospace;font-size:14px;padding:5px 10px;z-index:999999;border-bottom:2px solid #0f0;';
				urlBar.textContent = '📍 ' + window.location.href;
				document.body.insertBefore(urlBar, document.body.firstChild);
			})();
		`, nil),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.CaptureScreenshot(&buf),
	)

	if err != nil {
		log.Printf("Error capturing theme %s: %v\n", theme, err)
		return result
	}

	// Run accessibility check
	accessResult, err := RunAccessibilityCheck(ctx, theme)
	if err != nil {
		log.Printf("    ⚠️  Accessibility check failed: %v\n", err)
	} else {
		accessResult.Page = pageName
		result = accessResult
	}

	if err := os.WriteFile(filename, buf, 0644); err != nil {
		log.Printf("Error saving screenshot %s: %v\n", filename, err)
		return result
	}

	log.Printf("  [%3d/%d] %5.1f%% %s ✓\n", current, total, float64(current)/float64(total)*100, theme)
	return result
}

