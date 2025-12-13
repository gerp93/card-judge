package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/grantfbarnes/card-judge/tests/util"
)

// AccessibilityResult holds the axe results for a page/theme combo
type AccessibilityResult struct {
	Theme       string
	Page        string
	Violations  int
	Passes      int
	Incomplete  int
	InapplicableRules int
	ContrastIssues int
	WCAGLevel   string
	Details     axeResults
}

// axeResults mirrors the axe-core output structure
type axeResults struct {
	Violations   []map[string]interface{} `json:"violations"`
	Passes       []map[string]interface{} `json:"passes"`
	Incomplete   []map[string]interface{} `json:"incomplete"`
	Inapplicable []map[string]interface{} `json:"inapplicable"`
}

// RunAccessibilityCheck injects axe-core and runs accessibility tests
func RunAccessibilityCheck(ctx context.Context, theme string) (AccessibilityResult, error) {
	result := AccessibilityResult{
		Theme: theme,
	}

	// Inject axe-core library from CDN
	err := chromedp.Run(ctx,
		chromedp.Evaluate(getAxeInjection(), nil),
	)
	if err != nil {
		return result, fmt.Errorf("failed to inject axe-core: %w", err)
	}

	// Give axe-core additional time to load if needed
	time.Sleep(500 * time.Millisecond)

	// Check if axe loaded
	var axeLoaded bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`typeof axe !== 'undefined'`, &axeLoaded),
	)
	if err != nil || !axeLoaded {
		return result, fmt.Errorf("axe-core failed to load")
	}

	// Run axe scan - we need to use a synchronous approach since chromedp doesn't handle async well
	// First, start the axe scan and store results in a global variable
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			window.__axeResults = null;
			window.__axeDone = false;
			axe.run().then(function(results) {
				window.__axeResults = results;
				window.__axeDone = true;
			}).catch(function(e) {
				window.__axeResults = {error: e.message};
				window.__axeDone = true;
			});
		`, nil),
	)
	if err != nil {
		return result, fmt.Errorf("failed to start axe scan: %w", err)
	}

	// Poll for completion
	var done bool
	for i := 0; i < 100; i++ { // Max 10 seconds
		time.Sleep(100 * time.Millisecond)
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`window.__axeDone`, &done),
		)
		if err == nil && done {
			break
		}
	}

	if !done {
		return result, fmt.Errorf("axe scan timed out")
	}

	// Get the results
	var axeOutput interface{}
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`window.__axeResults`, &axeOutput),
	)
	if err != nil {
		return result, fmt.Errorf("failed to get axe results: %w", err)
	}

	// Convert to JSON then back to our struct
	jsonBytes, err := json.Marshal(axeOutput)
	if err != nil {
		return result, fmt.Errorf("failed to marshal axe output: %w", err)
	}

	// Parse results
	var axeRes axeResults
	if err := json.Unmarshal(jsonBytes, &axeRes); err != nil {
		log.Printf("DEBUG: JSON unmarshal error on output: %s", string(jsonBytes))
		return result, fmt.Errorf("failed to parse axe results: %w", err)
	}

	result.Details = axeRes
	result.Violations = len(axeRes.Violations)
	result.Passes = len(axeRes.Passes)
	result.Incomplete = len(axeRes.Incomplete)
	result.InapplicableRules = len(axeRes.Inapplicable)

	// Count contrast issues
	for _, violation := range axeRes.Violations {
		if id, ok := violation["id"].(string); ok && strings.Contains(id, "contrast") {
			if nodes, ok := violation["nodes"].([]interface{}); ok {
				result.ContrastIssues += len(nodes)
			}
		}
	}

	// Determine WCAG level
	result.WCAGLevel = determineWCAGLevel(result)

	return result, nil
}

// determineWCAGLevel determines WCAG compliance level based on violations
func determineWCAGLevel(result AccessibilityResult) string {
	if result.Violations == 0 {
		return util.WCAGLevelAAA
	}
	if result.ContrastIssues == 0 && result.Violations <= util.MaxViolationsForAA {
		return util.WCAGLevelAA
	}
	return util.WCAGLevelA
}

// GenerateAccessibilityReport creates a report from all accessibility results
func GenerateAccessibilityReport(results []AccessibilityResult, timestamp string) error {
	if len(results) == 0 {
		return nil
	}

	// Create theme-reports directory structure
	reportsDir := filepath.Join("theme-reports", timestamp)
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Create report file in the theme-reports dated folder
	reportPath := filepath.Join(reportsDir, "accessibility-report.txt")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "╔════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(file, "║                     THEME COLOR CONTRAST REPORT                              ║\n")
	fmt.Fprintf(file, "║                    WCAG Contrast Ratio Analysis by Theme                     ║\n")
	fmt.Fprintf(file, "╚════════════════════════════════════════════════════════════════════════════════╝\n\n")
	
	// Extract only color-contrast violations
	contrastIssuesByTheme := make(map[string][]map[string]interface{})
	totalContrastViolations := 0
	
	for _, result := range results {
		for _, violation := range result.Details.Violations {
			if id, ok := violation["id"].(string); ok && id == "color-contrast" {
				contrastIssuesByTheme[result.Theme] = append(contrastIssuesByTheme[result.Theme], violation)
				totalContrastViolations++
			}
		}
	}

	// Summary
	fmt.Fprintf(file, "📊 SUMMARY\n")
	fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(file, "Total Color Contrast Issues: %d\n", totalContrastViolations)
	fmt.Fprintf(file, "Themes Affected: %d\n", len(contrastIssuesByTheme))
	fmt.Fprintf(file, "\n")

	// Sort themes by contrast issue count
	type themeCount struct {
		name  string
		count int
	}
	themeCounts := make([]themeCount, 0, len(contrastIssuesByTheme))
	for theme, violations := range contrastIssuesByTheme {
		themeCounts = append(themeCounts, themeCount{theme, len(violations)})
	}
	sort.Slice(themeCounts, func(i, j int) bool {
		return themeCounts[i].count > themeCounts[j].count
	})

	// Severity ranking
	fmt.Fprintf(file, "🎨 THEMES BY CONTRAST SEVERITY\n")
	fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
	for i, tc := range themeCounts {
		var severity string
		if tc.count > 50 {
			severity = "🔴 CRITICAL"
		} else if tc.count > 20 {
			severity = "🟠 HIGH"
		} else if tc.count > 5 {
			severity = "🟡 MEDIUM"
		} else {
			severity = "🟢 LOW"
		}
		fmt.Fprintf(file, "%2d. %s - %d issues %s\n", i+1, tc.name, tc.count, severity)
	}
	fmt.Fprintf(file, "\n\n")

	// Detailed issues per theme
	fmt.Fprintf(file, "📋 DETAILED CONTRAST VIOLATIONS BY THEME\n")
	fmt.Fprintf(file, "═════════════════════════════════════════════════════════════════════════════════\n\n")

	for _, tc := range themeCounts {
		violations := contrastIssuesByTheme[tc.name]
		fmt.Fprintf(file, "🎨 THEME: %s\n", strings.ToUpper(tc.name))
		fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
		fmt.Fprintf(file, "Total Violations: %d\n\n", len(violations))

		// Group violations by element
		for idx, violation := range violations {
			fmt.Fprintf(file, "Issue #%d\n", idx+1)
			
			// Extract violation details
			if help, ok := violation["help"].(string); ok {
				fmt.Fprintf(file, "  Problem: %s\n", help)
			}
			
			if nodes, ok := violation["nodes"].([]interface{}); ok && len(nodes) > 0 {
				if node, ok := nodes[0].(map[string]interface{}); ok {
					if html, ok := node["html"].(string); ok {
						fmt.Fprintf(file, "  Element: %s\n", html)
					}
					
					// Try to extract detailed failure info
					if any, ok := node["any"].([]interface{}); ok && len(any) > 0 {
						if failureInfo, ok := any[0].(map[string]interface{}); ok {
							if message, ok := failureInfo["message"].(string); ok {
								fmt.Fprintf(file, "  Details: %s\n", message)
							}
						}
					}
				}
			}
			
			fmt.Fprintf(file, "\n")
		}
		fmt.Fprintf(file, "\n")
	}

	// Themes with no contrast issues
	fmt.Fprintf(file, "✅ THEMES WITH NO CONTRAST ISSUES\n")
	fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
	noContrastThemes := 0
	for _, tc := range themeCounts {
		if tc.count == 0 {
			fmt.Fprintf(file, "  ✓ %s\n", tc.name)
			noContrastThemes++
		}
	}
	if noContrastThemes == 0 {
		fmt.Fprintf(file, "  (None - all themes have at least one contrast issue)\n")
	}

	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "📚 WCAG CONTRAST REQUIREMENTS\n")
	fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(file, "  Level A:   4.5:1 for normal text, 3:1 for large text\n")
	fmt.Fprintf(file, "  Level AA:  4.5:1 for normal text, 3:1 for large text (WCAG 2.0 standard)\n")
	fmt.Fprintf(file, "  Level AAA: 7:1 for normal text, 4.5:1 for large text\n")
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "💡 RECOMMENDATIONS\n")
	fmt.Fprintf(file, "─────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(file, "High priority themes needing color adjustments:\n")
	for i, tc := range themeCounts {
		if i >= 5 {
			break
		}
		if tc.count > 0 {
			fmt.Fprintf(file, "  • %s (%d issues)\n", tc.name, tc.count)
		}
	}
	fmt.Fprintf(file, "\nTo fix: Increase text brightness/darkness or adjust background colors to meet\n")
	fmt.Fprintf(file, "the minimum contrast ratios listed above.\n\n")
	fmt.Fprintf(file, "For visual inspection with color overlays, use axe DevTools browser extension:\n")
	fmt.Fprintf(file, "https://www.deque.com/axe/devtools/\n")

	log.Printf("📋 Accessibility report saved to: %s\n", reportPath)
	return nil
}

// getAxeInjection loads axe-core from CDN
func getAxeInjection() string {
	return `
		new Promise((resolve) => {
			if (typeof axe !== 'undefined') {
				resolve();
				return;
			}
			// Load axe-core from CDN
			var script = document.createElement('script');
			script.src = 'https://cdnjs.cloudflare.com/ajax/libs/axe-core/4.7.2/axe.min.js';
			script.onload = function() {
				resolve();
			};
			script.onerror = function() {
				console.error('Failed to load axe-core');
				resolve();
			};
			document.head.appendChild(script);
		});
	`
}
