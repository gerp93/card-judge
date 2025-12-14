package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// PerformLogin authenticates with the application and returns true if successful
func PerformLogin(ctx context.Context, baseURL, username, password string) bool {
	log.Println("Logging in...")

	// Navigate to login page and submit credentials
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/login"),
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="name"]`, username, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="password"]`, password, chromedp.ByQuery),
	)

	if err != nil {
		log.Printf("ERROR: Failed to fill login form - %v\n", err)
		return false
	}

	log.Println("Submitting login form...")

	// Submit and wait
	err = chromedp.Run(ctx,
		chromedp.Click(`input[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for HTMX to process
	)

	if err != nil {
		log.Printf("ERROR: Failed to submit login form - %v\n", err)
		return false
	}

	// Check current URL - if we're not on /login, login succeeded
	var currentURL string
	chromedp.Run(ctx, chromedp.Location(&currentURL))
	log.Printf("Current URL: %s\n", currentURL)

	if currentURL == baseURL+"/login" {
		// Still on login page - check for error message
		var errorMsg string
		chromedp.Run(ctx, chromedp.Text(`.htmx-result`, &errorMsg, chromedp.ByQuery))
		log.Printf("ERROR: Login failed - still on login page\n")
		if errorMsg != "" {
			log.Printf("Error message from server: %s\n", errorMsg)
		}
		log.Printf("Check credentials: username=%s, password=%s\n", username, password)
		return false
	}

	log.Println("✅ Login successful! Session is active.")
	return true
}
