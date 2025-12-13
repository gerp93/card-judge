package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/grantfbarnes/card-judge/tests/util"
)

// PageConfig defines a page to capture with optional setup
type PageConfig struct {
	Name string
	Path string
	// Setup is an optional function to run before taking screenshot (e.g., clicking buttons, waiting for dynamic content)
	Setup func(ctx context.Context) error
}

// GetPageConfigurations returns all page configurations for screenshot capture
func GetPageConfigurations() []PageConfig {
	return []PageConfig{
		{Name: "home", Path: "/"},
		{
			Name: "about",
			Path: "/about",
			Setup: expandDetailsSetup,
		},
		{
			Name: "account",
			Path: "/account",
			Setup: expandDetailsSetup,
		},
		{Name: "review", Path: "/review"},
		{Name: "stats", Path: "/stats"},
		{Name: "stats-leaderboard", Path: "/stats/leaderboard"},
		{Name: "stats-users", Path: "/stats/users"},
		{Name: "stats-user", Path: "/stats/user/" + util.TestUser1ID},
		{Name: "stats-cards", Path: "/stats/cards"},
		{Name: "stats-card", Path: "/stats/card/" + util.TestCardID},
		{Name: "users", Path: "/users"},
		{
			Name: "users-create-modal",
			Path: "/users",
			Setup: openModalSetup("user-create-dialog"),
		},
		{Name: "lobbies", Path: "/lobbies"},
		{
			Name: "lobbies-create-modal",
			Path: "/lobbies",
			Setup: openModalSetup("lobby-create-dialog"),
		},
		{Name: "decks", Path: "/decks"},
		{
			Name: "decks-create-modal",
			Path: "/decks",
			Setup: openModalSetup("deck-create-dialog"),
		},
		{
			Name: "deck-view",
			Path: "/decks",
			Setup: openFirstDeckSetup,
		},
		{
			Name: "lobby-game",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: nil,
		},
		{
			Name: "lobby-game-message-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("lobby-message-dialog"),
		},
		{
			Name: "lobby-game-update-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("lobby-update-dialog"),
		},
		{
			Name: "lobby-game-draw-pile-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("lobby-draw-pile-dialog"),
		},
		{
			Name: "lobby-game-purchase-credits-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("purchase-credits-dialog"),
		},
		{
			Name: "lobby-game-find-card-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("find-card-dialog"),
		},
		{
			Name: "lobby-game-wild-card-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("wild-card-dialog"),
		},
		{
			Name: "lobby-game-alert-lobby-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("alert-lobby-dialog"),
		},
		{
			Name: "lobby-game-gamble-credits-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("gamble-credits-dialog"),
		},
		{
			Name: "lobby-game-bet-on-win-modal",
			Path: "/lobby/" + util.TestLobby1ID,
			Setup: openModalSetup("bet-on-win-dialog"),
		},
		{
			Name: "lobby-game-no-message",
			Path: "/lobby/" + util.TestLobby2ID,
			Setup: nil,
		},
	}
}

// expandDetailsSetup expands all details sections on the page
func expandDetailsSetup(ctx context.Context) error {
	return chromedp.Run(ctx,
		chromedp.Evaluate(`
			document.querySelectorAll('details').forEach(details => {
				details.open = true;
			});
		`, nil),
		chromedp.Sleep(500*time.Millisecond),
	)
}

// openModalSetup returns a setup function that opens a modal by ID and expands details
func openModalSetup(dialogID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.getElementById('%s').showModal()`, dialogID), nil),
			chromedp.Evaluate(fmt.Sprintf(`
				document.getElementById('%s').querySelectorAll('details').forEach(details => {
					details.open = true;
				});
			`, dialogID), nil),
			chromedp.Sleep(500*time.Millisecond),
		)
	}
}

// openFirstDeckSetup navigates to the first deck in the list
func openFirstDeckSetup(ctx context.Context) error {
	// Get the href of the first deck link
	var href string
	err := chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second), // Wait for page to render
		chromedp.AttributeValue(`a[href*="/deck/"]`, "href", &href, nil, chromedp.ByQuery),
	)
	if err != nil || href == "" {
		return fmt.Errorf("no decks found - please create a deck first")
	}

	// Navigate to that deck
	return chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("http://%s:%d", util.DefaultHost, util.DefaultPort)+href),
		chromedp.Sleep(1*time.Second), // Wait for deck to load
	)
}
