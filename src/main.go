package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grantfbarnes/card-judge/api"
	apiAccess "github.com/grantfbarnes/card-judge/api/access"
	apiCard "github.com/grantfbarnes/card-judge/api/card"
	apiDeck "github.com/grantfbarnes/card-judge/api/deck"
	apiLobby "github.com/grantfbarnes/card-judge/api/lobby"
	apiPages "github.com/grantfbarnes/card-judge/api/pages"
	apiStats "github.com/grantfbarnes/card-judge/api/stats"
	apiUser "github.com/grantfbarnes/card-judge/api/user"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/static"
	"github.com/grantfbarnes/card-judge/websocket"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	db, err := database.CreateDatabaseConnection()
	dbConnectAttemptCount := 0
	for err != nil && dbConnectAttemptCount < 6 {
		time.Sleep(10 * time.Second)
		dbConnectAttemptCount += 1
		db, err = database.CreateDatabaseConnection()
	}
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()

	for _, sqlFile := range static.SQLFiles {
		err = database.RunFile(sqlFile)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	// static files
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.StaticFiles))))

	// pages
	http.Handle("GET /", api.MiddlewareForPages(http.HandlerFunc(apiPages.Home)))
	http.Handle("GET /about", api.MiddlewareForPages(http.HandlerFunc(apiPages.About)))
	http.Handle("GET /login", api.MiddlewareForPages(http.HandlerFunc(apiPages.Login)))
	http.Handle("GET /account", api.MiddlewareForPages(http.HandlerFunc(apiPages.Account)))
	http.Handle("GET /stats", api.MiddlewareForPages(http.HandlerFunc(apiPages.Stats)))
	http.Handle("GET /stats/leaderboard", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsLeaderboard)))
	http.Handle("GET /stats/users", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsUsers)))
	http.Handle("GET /stats/user/{userId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsUser)))
	http.Handle("GET /stats/cards", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsCards)))
	http.Handle("GET /stats/card/{cardId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.StatsCard)))
	http.Handle("GET /users", api.MiddlewareForPages(http.HandlerFunc(apiPages.Users)))
	http.Handle("GET /review", api.MiddlewareForPages(http.HandlerFunc(apiPages.Review)))
	http.Handle("GET /lobbies", api.MiddlewareForPages(http.HandlerFunc(apiPages.Lobbies)))
	http.Handle("GET /lobby/{lobbyId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.Lobby)))
	http.Handle("GET /lobby/{lobbyId}/access", api.MiddlewareForPages(http.HandlerFunc(apiPages.LobbyAccess)))
	http.Handle("GET /decks", api.MiddlewareForPages(http.HandlerFunc(apiPages.Decks)))
	http.Handle("GET /deck/{deckId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.Deck)))
	http.Handle("GET /deck/{deckId}/access", api.MiddlewareForPages(http.HandlerFunc(apiPages.DeckAccess)))

	// user
	http.Handle("POST /api/user/create", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Create)))
	http.Handle("POST /api/user/create/admin", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.CreateAdmin)))
	http.Handle("POST /api/user/login", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Login)))
	http.Handle("POST /api/user/logout", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Logout)))
	http.Handle("PUT /api/user/{userId}/name", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetName)))
	http.Handle("PUT /api/user/{userId}/password", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetPassword)))
	http.Handle("PUT /api/user/{userId}/password/reset", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.ResetPassword)))
	http.Handle("PUT /api/user/{userId}/color-theme", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetColorTheme)))
	http.Handle("PUT /api/user/{userId}/approve", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Approve)))
	http.Handle("PUT /api/user/{userId}/is-admin", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetIsAdmin)))
	http.Handle("DELETE /api/user/{userId}", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Delete)))

	// deck
	http.Handle("GET /api/deck/{deckId}/card-export", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.GetCardExport)))
	http.Handle("POST /api/deck/create", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.Create)))
	http.Handle("PUT /api/deck/{deckId}/name", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetName)))
	http.Handle("PUT /api/deck/{deckId}/password", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetPassword)))
	http.Handle("PUT /api/deck/{deckId}/is-public-read-only", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetIsPublicReadOnly)))
	http.Handle("DELETE /api/deck/{deckId}", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.Delete)))

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
	http.Handle("POST /api/lobby/{lobbyId}/card/steal/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayStealCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/surprise/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlaySurpriseCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/find/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayFindCard)))
	http.Handle("POST /api/lobby/{lobbyId}/card/wild/play", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.PlayWildCard)))
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
	http.Handle("PUT /api/lobby/{lobbyId}/free-credits", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetFreeCredits)))
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

	if os.Getenv("CARD_JUDGE_LOG_FILE") != "" {
		logFile, err := os.OpenFile(os.Getenv("CARD_JUDGE_LOG_FILE"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	port := ":2016"
	if os.Getenv("CARD_JUDGE_PORT") != "" {
		port = ":" + os.Getenv("CARD_JUDGE_PORT")
	}

	log.Println("server is running...")
	if os.Getenv("CARD_JUDGE_CERT_FILE") != "" && os.Getenv("CARD_JUDGE_KEY_FILE") != "" {
		err = http.ListenAndServeTLS(port, os.Getenv("CARD_JUDGE_CERT_FILE"), os.Getenv("CARD_JUDGE_KEY_FILE"), nil)
	} else {
		err = http.ListenAndServe(port, nil)
	}
	if err != nil {
		log.Fatalln(err)
	}
}
