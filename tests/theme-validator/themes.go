package main

import (
	"log"
	"os"
	"regexp"
	"sort"

	"github.com/grantfbarnes/card-judge/tests/util"
)

// GetThemes returns all available theme class names by parsing colors.css
func GetThemes() []string {
	colorsCSS := getColorsCSSPath()
	content, err := os.ReadFile(colorsCSS)
	if err != nil {
		log.Fatalf("Failed to read colors.css: %v\n", err)
	}

	// Parse theme names from "body.{theme-name} {" selectors
	// Regex matches: body.{captured-theme-name} {
	re := regexp.MustCompile(`body\.([a-z0-9\-]+)\s*{`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	var themes []string
	seen := make(map[string]bool)

	// Extract unique theme names in order they appear
	for _, match := range matches {
		if len(match) > 1 {
			theme := match[1]
			if !seen[theme] {
				themes = append(themes, theme)
				seen[theme] = true
			}
		}
	}

	// Sort for consistency
	sort.Strings(themes)

	if len(themes) == 0 {
		log.Fatalf("No themes found in %s\n", colorsCSS)
	}

	return themes
}

// getColorsCSSPath finds the colors.css file
func getColorsCSSPath() string {
	return util.FindPath(
		util.ColorsCSSPath,
		"./"+util.ColorsCSSPath,
		"../"+util.ColorsCSSPath,
		"../../"+util.ColorsCSSPath,
		"../../../"+util.ColorsCSSPath,
	)
}
