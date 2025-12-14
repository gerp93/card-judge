package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/grantfbarnes/card-judge/tests/util"
	"github.com/jung-kurt/gofpdf"
)

// GenerateThemePDFs creates a PDF for each theme with all its screenshots
func GenerateThemePDFs(screenshotBaseDir string, timestamp string) error {
	screenshotDir := filepath.Join(screenshotBaseDir, timestamp)

	// Map of theme -> list of (page name, file path)
	themes := make(map[string][]struct {
		page string
		path string
	})

	// Walk through the pages directory
	pagesDir := screenshotDir
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		return fmt.Errorf("failed to read screenshot directory: %w", err)
	}

	// Process each page directory
	for _, pageEntry := range entries {
		if !pageEntry.IsDir() {
			continue
		}

		pageName := pageEntry.Name()
		pageDir := filepath.Join(pagesDir, pageName)

		// Read PNG files from page directory
		pageFiles, err := os.ReadDir(pageDir)
		if err != nil {
			fmt.Printf("Warning: Failed to read page directory %s: %v\n", pageDir, err)
			continue
		}

		for _, fileEntry := range pageFiles {
			if !fileEntry.IsDir() && strings.HasSuffix(fileEntry.Name(), ".png") {
				// Extract theme from filename
				// Format: theme_timestamp.png
				filename := fileEntry.Name()
				parts := strings.Split(strings.TrimSuffix(filename, ".png"), "_")

				if len(parts) >= 2 {
					// Theme is everything except the last part (timestamp)
					theme := strings.Join(parts[:len(parts)-1], "_")
					filePath := filepath.Join(pageDir, filename)

					themes[theme] = append(themes[theme], struct {
						page string
						path string
					}{page: pageName, path: filePath})
				}
			}
		}
	}

	// Create PDF for each theme
	pdfDir := filepath.Join(filepath.Dir(screenshotBaseDir), util.ThemeReportDir, timestamp)
	if err := os.MkdirAll(pdfDir, 0755); err != nil {
		return fmt.Errorf("failed to create PDF directory: %w", err)
	}

	// Sort themes for consistent output
	themeNames := make([]string, 0, len(themes))
	for theme := range themes {
		themeNames = append(themeNames, theme)
	}
	sort.Strings(themeNames)

	successCount := 0
	failureCount := 0

	for _, theme := range themeNames {
		images := themes[theme]

		// Sort by page name
		sort.Slice(images, func(i, j int) bool {
			return images[i].page < images[j].page
		})

		if err := generateThemePDF(theme, images, pdfDir); err != nil {
			fmt.Printf("⚠️  Failed to generate PDF for theme %s: %v\n", theme, err)
			failureCount++
			continue
		}
		fmt.Printf("  ✓ %s\n", formatThemeName(theme))
		successCount++
	}

	fmt.Printf("\n📄 Theme PDFs saved to: %s\n", pdfDir)
	fmt.Printf("📊 Generated %d PDFs successfully", successCount)
	if failureCount > 0 {
		fmt.Printf(" (%d failed)", failureCount)
		return fmt.Errorf("%d PDF(s) failed to generate", failureCount)
	}
	fmt.Println()
	return nil
}

// generateThemePDF creates a single PDF for a theme
func generateThemePDF(theme string, images []struct {
	page string
	path string
}, outputDir string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(true)

	// Title page
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 28)
	pdf.Cell(0, 20, formatThemeName(theme))
	pdf.Ln(30)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Pages: %d", len(images)))
	pdf.Ln(5)
	pdf.Cell(0, 8, "All themes and pages for visual comparison")

	// Add pages with screenshots
	for _, img := range images {
		// Check if file exists before adding to PDF
		if _, err := os.Stat(img.path); err != nil {
			fmt.Printf("Warning: Screenshot file not found: %s\n", img.path)
			continue
		}

		pdf.AddPage()

		// Page header
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, formatPageName(img.page))
		pdf.Ln(15)

		// Add image
		imgWidth := 190.0
		imgHeight := 107.0 // A4 aspect ratio

		// Files are now actual PNG format
		pdf.Image(img.path, 10, pdf.GetY(), imgWidth, imgHeight, false, "PNG", 0, "")
		pdf.Ln(imgHeight + 10)
	}

	// Save PDF
	formattedTheme := formatThemeName(theme)
	// Replace spaces with hyphens for filename
	sanitizedTheme := strings.ReplaceAll(formattedTheme, " ", "-")
	filename := filepath.Join(outputDir, fmt.Sprintf("%s.pdf", sanitizedTheme))
	return pdf.OutputFileAndClose(filename)
}

func formatThemeName(theme string) string {
	// Convert kebab-case to Title Case
	parts := strings.Split(theme, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func formatPageName(page string) string {
	// Convert kebab-case to Title Case
	parts := strings.Split(page, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
