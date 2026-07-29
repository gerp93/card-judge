package main

import (
	"log"
	"net/http"
	"time"

	gameshell "github.com/gerp93/gameshell-framework"
	"github.com/gerp93/gameshell-framework/api"
	"github.com/gerp93/gameshell-framework/auth"
	gsBootstrap "github.com/gerp93/gameshell-framework/bootstrap"
	gsDatabase "github.com/gerp93/gameshell-framework/database"
	gsStatic "github.com/gerp93/gameshell-framework/static"
	"github.com/gerp93/gameshell-framework/websocket"
	apiAccess "github.com/grantfbarnes/card-judge/api/access"
	apiCard "github.com/grantfbarnes/card-judge/api/card"
	apiDeck "github.com/grantfbarnes/card-judge/api/deck"
	apiLobby "github.com/grantfbarnes/card-judge/api/lobby"
	apiPages "github.com/grantfbarnes/card-judge/api/pages"
	apiStats "github.com/grantfbarnes/card-judge/api/stats"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/game"
	"github.com/grantfbarnes/card-judge/static"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	gameshell.Register(game.CardJudge{})

	api.SetBrandName("Card Judge")
	auth.SetCookiePrefix("CARD-JUDGE")
	api.SetPagePolicy(api.PagePolicy{
		LoginPaths:        []string{"/account", "/users", "/review", "/lobbies", "/decks"},
		LoginPathPrefixes: []string{"/stats", "/lobby", "/deck"},
		AdminPaths:        []string{"/users", "/review"},
	})

	gsDatabase.SetEnvVarPrefix("CARD_JUDGE")
	// WinCelebration and LobbyTurnTimer stay off deliberately: this game has
	// no win-image/message UI, and already has its own round-timer concept
	// (CJ_LOBBY_SETTINGS) rather than the framework's.
	features := gsBootstrap.Features{
		Decks:          true,
		WinCelebration: false,
		LobbyTurnTimer: false,
	}
	gsBootstrap.MountFeatures(features)

	db := gsBootstrap.ConnectWithRetry(6, 10*time.Second)
	defer db.Close()

	// framework schema must load before game schema
	gsBootstrap.ApplySchema(gsStatic.StaticFiles, gsStatic.SQLFiles)
	gsBootstrap.ApplyFeatureSchema(features)
	gsBootstrap.ApplySchema(static.StaticFiles, static.SQLFiles)

	// TODO(remove-me): dev-convenience seed (default/password admin + a few
	// test players) so a fresh local DB has an immediate login. Flagged for
	// likely removal — see database/seed_dev_users.go.
	if err := database.SeedDevUsersIfEmpty(); err != nil {
		log.Fatalln(err)
		return
	}

	// static files (game's own at /static/, shared framework assets at /gs/)
	gsBootstrap.MountStaticAssets(static.StaticFiles)

	// pages (game-owned; framework's core + Features-gated pages are wired by MountFeatures)
	// "/{$}" (not "/"): a bare "/" is a Go 1.22+ subtree wildcard matching
	// every unmatched path, silently serving Home for any bad URL instead of
	// a real 404. "{$}" restricts the match to the literal root only.
	http.Handle("GET /{$}", api.MiddlewareForPages(http.HandlerFunc(apiPages.Home)))
	http.Handle("GET /about", api.MiddlewareForPages(http.HandlerFunc(apiPages.About)))
	http.Handle("GET /stats", api.MiddlewareForPages(http.HandlerFunc(apiPages.Stats)))
	http.Handle("GET /stats/leaderboard", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsLeaderboard)))
	http.Handle("GET /stats/users", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsUsers)))
	http.Handle("GET /stats/user/{userId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsUser)))
	http.Handle("GET /stats/cards", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsCards)))
	http.Handle("GET /stats/card/{cardId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsCard)))
	http.Handle("GET /review", api.MiddlewareForPages(http.HandlerFunc(apiPages.Review)))
	http.Handle("GET /lobbies", api.MiddlewareForPages(http.HandlerFunc(apiPages.Lobbies)))
	http.Handle("GET /lobby/{lobbyId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.Lobby)))
	http.Handle("GET /lobby/{lobbyId}/access", api.MiddlewareForPages(http.HandlerFunc(apiPages.LobbyAccess)))
	http.Handle("GET /deck/{deckId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.Deck)))

	// deck (game-owned; card export/CSV is card-judge's own)
	http.Handle("GET /api/deck/{deckId}/card-export", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.GetCardExport)))

	// card
	http.Handle("POST /api/card/find", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.Find)))
	http.Handle("POST /api/card/create", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.Create)))
	http.Handle("PUT /api/card/{cardId}", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.Update)))
	http.Handle("PUT /api/card/{cardId}/image", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.SetImage)))
	http.Handle("DELETE /api/card/{cardId}", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.Delete)))

	// review card
	http.Handle("PUT /api/card/review/{Id}/recover", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.Recover)))
	http.Handle("DELETE /api/card/review/{Id}", api.MiddlewareForAPIs(http.HandlerFunc(apiCard.PermanentlyDelete)))

	// lobby
	http.Handle("GET /api/lobby/{lobbyId}/html/game-interface", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetGameInterfaceHTML)))
	http.Handle("GET /api/lobby/{lobbyId}/html/lobby-game-info", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetLobbyGameInfoHTML)))
	http.Handle("GET /api/lobby/{lobbyId}/html/player-hand", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetPlayerHandHTML)))
	http.Handle("GET /api/lobby/{lobbyId}/html/player-specials", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetPlayerSpecialsHTML)))
	http.Handle("GET /api/lobby/{lobbyId}/html/lobby-game-board", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetLobbyGameBoardHTML)))
	http.Handle("GET /api/lobby/{lobbyId}/html/lobby-game-stats", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GetLobbyGameStatsHTML)))
	http.Handle("POST /api/lobby/create", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.Create)))
	http.Handle("POST /api/lobby/{lobbyId}/card/{cardId}/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/force/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayForceCard)))
	http.Handle("POST /api/lobby/{lobbyId}/purchase-credits", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PurchaseCredits)))
	http.Handle("POST /api/lobby/{lobbyId}/skip-judge", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SkipJudge)))
	http.Handle("POST /api/lobby/{lobbyId}/reset-responses", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.ResetResponses)))
	http.Handle("POST /api/lobby/{lobbyId}/alert", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.AlertLobby)))
	http.Handle("POST /api/lobby/{lobbyId}/gamble-credits", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.GambleCredits)))
	http.Handle("POST /api/lobby/{lobbyId}/bet-on-win", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.BetOnWin)))
	http.Handle("POST /api/lobby/{lobbyId}/bet-on-win/undo", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.BetOnWinUndo)))
	http.Handle("POST /api/lobby/{lobbyId}/add-extra-response", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.AddExtraResponse)))
	http.Handle("POST /api/lobby/{lobbyId}/add-extra-response/undo", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.AddExtraResponseUndo)))
	http.Handle("POST /api/lobby/{lobbyId}/block-response", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.BlockResponse)))
	http.Handle("POST /api/lobby/{lobbyId}/card/surprise/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlaySurpriseCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/steal/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayStealCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/find/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayFindCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/wild/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayWildCard)))
	http.Handle("POST /api/lobby/{lobbyId}/perk/hand-size-advantage", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PerkHandSizeAdvantage)))
	http.Handle("POST /api/lobby/{lobbyId}/perk/discard-advantage", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PerkDiscardAdvantage)))
	http.Handle("POST /api/lobby/{lobbyId}/perk/handicap-advantage", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PerkHandicapAdvantage)))
	http.Handle("POST /api/lobby/{lobbyId}/perk/spy-advantage", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PerkSpyAdvantage)))
	http.Handle("POST /api/lobby/{lobbyId}/response-card/{responseCardId}/withdraw", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.WithdrawCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/{cardId}/discard", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.DiscardCard)))
	http.Handle("POST /api/lobby/{lobbyId}/player/{playerId}/kick", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.VoteToKick)))
	http.Handle("POST /api/lobby/{lobbyId}/player/{playerId}/kick/undo", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.VoteToKickUndo)))
	http.Handle("POST /api/lobby/{lobbyId}/response/{responseId}/reveal", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.RevealResponse)))
	http.Handle("POST /api/lobby/{lobbyId}/response/{responseId}/toggle-rule-out", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.ToggleRuleOutResponse)))
	http.Handle("POST /api/lobby/{lobbyId}/response/{responseId}/pick-winner", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PickWinner)))
	http.Handle("POST /api/lobby/{lobbyId}/pick-random-winner", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PickRandomWinner)))
	http.Handle("POST /api/lobby/{lobbyId}/flip", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.FlipTable)))
	http.Handle("POST /api/lobby/{lobbyId}/skip-prompt", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SkipPrompt)))
	http.Handle("PUT /api/lobby/{lobbyId}/name", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetName)))
	http.Handle("PUT /api/lobby/{lobbyId}/message", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetMessage)))
	http.Handle("PUT /api/lobby/{lobbyId}/draw-priority", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetDrawPriority)))
	http.Handle("PUT /api/lobby/{lobbyId}/hand-size", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetHandSize)))
	http.Handle("PUT /api/lobby/{lobbyId}/round-timer/set", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetRoundTimer)))
	http.Handle("PUT /api/lobby/{lobbyId}/round-timer/start", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.StartRoundTimer)))
	http.Handle("PUT /api/lobby/{lobbyId}/free-credits", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetFreeCredits)))
	http.Handle("PUT /api/lobby/{lobbyId}/free-special-cards", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetFreeSpecialCards)))
	http.Handle("PUT /api/lobby/{lobbyId}/win-streak-threshold", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetWinStreakThreshold)))
	http.Handle("PUT /api/lobby/{lobbyId}/lose-streak-threshold", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetLoseStreakThreshold)))
	http.Handle("PUT /api/lobby/{lobbyId}/response-count", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetResponseCount)))
	http.Handle("PUT /api/lobby/{lobbyId}/set-decks", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetDecks)))

	// access
	http.Handle("POST /api/access/lobby/{lobbyId}", api.MiddlewareForAPIs(http.HandlerFunc(apiAccess.Lobby)))
	http.Handle("POST /api/access/deck/{deckId}", api.MiddlewareForAPIs(http.HandlerFunc(apiAccess.Deck)))

	// stats
	http.Handle("POST /api/stats/leaderboard", api.MiddlewareForAPIs(http.HandlerFunc(apiStats.GetLeaderboard)))

	// websocket
	http.HandleFunc("GET /ws/lobby/{lobbyId}", websocket.ServeWs)

	gsBootstrap.Serve("CARD_JUDGE")
}
