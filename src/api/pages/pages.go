package apiPages

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/static"
)

func Home(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Home"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/home.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func About(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - About"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/about.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func Login(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Login"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/login.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func Account(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Account"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/account.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func Users(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Users"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountUsers(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	users, err := database.SearchUsers(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/users.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Users    []database.User
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Users:        users,
	})
}

func Review(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Review"

	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountCardsInReview()
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	cards, err := database.SearchCardsInReview(page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/review.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Page     int
		LastPage int
		RowCount int
		Cards    []database.DisplayCard
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Cards:        cards,
	})
}

func Stats(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func StatsLeaderboard(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats - Leaderboard"

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats-leaderboard.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func StatsUsers(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats - Users"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountUsers(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	users, err := database.SearchUsers(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats-users.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Users    []database.User
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Users:        users,
	})
}

func StatsUser(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats - User"

	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse id."))
		return
	}

	userStats, err := database.GetStatsUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to get user stats."))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats-user.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		database.StatUser
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		StatUser:     userStats,
	})
}

func StatsCards(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats - Cards"

	var deckName string
	var category string
	var text string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "deckName":
			deckName = val[0]
		case "category":
			category = val[0]
		case "text":
			text = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountCardsWithAccess(basePageData.User.Id, deckName, category, text)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	cards, err := database.SearchCardsWithAccess(basePageData.User.Id, deckName, category, text, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats-cards.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		DeckName string
		Category string
		Text     string
		Page     int
		LastPage int
		RowCount int
		Cards    []database.DisplayCard
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		DeckName:     deckName,
		Category:     category,
		Text:         text,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Cards:        cards,
	})
}

func StatsCard(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Stats - Card"

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse id."))
		return
	}

	cardStats, err := database.GetStatsCard(cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to get card stats."))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/stats-card.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		database.StatCard
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		StatCard:     cardStats,
	})
}

func Lobbies(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Lobbies"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountLobbies(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	lobbies, err := database.SearchLobbies(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	decks, err := database.GetReadableDecks(basePageData.User.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get user decks"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/lobbies.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Lobbies  []database.LobbyDetails
		Decks    []database.Deck
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Lobbies:      lobbies,
		Decks:        decks,
	})
}

func Lobby(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	if lobby.Id == uuid.Nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Lobby"

	hasLobbyAccess, err := database.UserHasLobbyAccess(basePageData.User.Id, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check lobby access"))
		return
	}

	if !hasLobbyAccess {
		http.Redirect(w, r, fmt.Sprintf("/lobby/%s/access", lobbyId), http.StatusSeeOther)
		return
	}

	decks, err := database.GetReadableDecks(basePageData.User.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get user decks"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/lobby.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	playerId, err := database.AddUserToLobby(lobbyId, basePageData.User.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to join lobby"))
		return
	}

	type data struct {
		api.BasePageData
		Lobby    database.Lobby
		PlayerId uuid.UUID
		Decks    []database.Deck
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Lobby:        lobby,
		PlayerId:     playerId,
		Decks:        decks,
	})
}

func LobbyAccess(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	if lobby.Id == uuid.Nil {
		http.Redirect(w, r, "/lobbies", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Lobby Access"

	hasLobbyAccess, err := database.UserHasLobbyAccess(basePageData.User.Id, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check lobby access"))
		return
	}

	if hasLobbyAccess {
		http.Redirect(w, r, fmt.Sprintf("/lobby/%s", lobbyId), http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/lobby-access.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Lobby database.Lobby
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Lobby:        lobby,
	})
}

func Decks(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Decks"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountDecks(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	decks, err := database.SearchDecks(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/decks.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Decks    []database.DeckDetails
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Decks:        decks,
	})
}

func Deck(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	deck, err := database.GetDeck(deckId)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	if deck.Id == uuid.Nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Deck"

	hasDeckAccess, err := database.UserHasDeckAccess(basePageData.User.Id, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check deck access"))
		return
	}

	if !hasDeckAccess {
		http.Redirect(w, r, fmt.Sprintf("/deck/%s/access", deckId), http.StatusSeeOther)
		return
	}

	var category string
	var text string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "category":
			category = val[0]
		case "text":
			text = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountCardsInDeck(deckId, category, text)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}

	if page > totalPageCount {
		page = totalPageCount
	}

	cards, err := database.SearchCardsInDeck(deckId, category, text, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/deck.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Deck     database.Deck
		Category string
		Text     string
		Page     int
		LastPage int
		RowCount int
		Cards    []database.Card
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Deck:         deck,
		Category:     category,
		Text:         text,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Cards:        cards,
	})
}

func DeckAccess(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	deck, err := database.GetDeck(deckId)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	if deck.Id == uuid.Nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Deck"

	hasDeckAccess, err := database.UserHasDeckAccess(basePageData.User.Id, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check deck access"))
		return
	}

	if hasDeckAccess {
		http.Redirect(w, r, fmt.Sprintf("/deck/%s", deckId), http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/deck-access.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Deck database.Deck
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Deck:         deck,
	})
}

// ChronologyLobbies displays the list of Chronology games
func ChronologyLobbies(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Chronology"

	decks, err := database.GetReadableDecks(basePageData.User.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get user decks"))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/chronology-lobbies.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Decks []database.Deck
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Decks:        decks,
	})
}

// ChronologyLobby displays a specific Chronology game
func ChronologyLobby(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	if lobby.Id == uuid.Nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Chronology"

	hasLobbyAccess, err := database.UserHasLobbyAccess(basePageData.User.Id, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check lobby access"))
		return
	}

	if !hasLobbyAccess {
		http.Redirect(w, r, fmt.Sprintf("/chronology/%s/access", lobbyId), http.StatusSeeOther)
		return
	}

	// Get or create player for this user
	playerId, err := database.AddUserToLobby(lobbyId, basePageData.User.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to join lobby"))
		return
	}

	// Get the Chronology game - auto-create if it doesn't exist
	game, err := database.GetChronologyGame(lobbyId)
	if err != nil || game.Id == uuid.Nil {
		log.Printf("[INFO ChronologyLobby] Game not found for lobby %s, auto-creating...", lobbyId)
		// Auto-create the game with default settings
		gameId, createErr := database.CreateChronologyGame(lobbyId, 5) // 5 cards to win default
		if createErr != nil {
			log.Printf("[ERROR ChronologyLobby] Failed to auto-create game for lobby %s: %v", lobbyId, createErr)
			http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
			return
		}
		// Initialize draw pile with the Chronology deck
		chronologyDeckId, _ := uuid.Parse("88026803-d22a-11f0-b4d2-60cf84649547")
		if initErr := database.InitializeChronologyDrawPile(gameId, []uuid.UUID{chronologyDeckId}); initErr != nil {
			log.Printf("[ERROR ChronologyLobby] Failed to initialize draw pile for lobby %s: %v", lobbyId, initErr)
		}
		// Parse years from card text
		if yearErr := database.UpdateDrawPileYears(gameId); yearErr != nil {
			log.Printf("[ERROR ChronologyLobby] Failed to update draw pile years for lobby %s: %v", lobbyId, yearErr)
		}
		// Re-fetch the game
		game, err = database.GetChronologyGame(lobbyId)
		if err != nil || game.Id == uuid.Nil {
			log.Printf("[ERROR ChronologyLobby] Still no game after auto-create for lobby %s", lobbyId)
			http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
			return
		}
		log.Printf("[INFO ChronologyLobby] Auto-created game %s for lobby %s", game.Id, lobbyId)
	}

	// Get current player name if game is active
	var currentPlayerName string
	var isMyTurn bool
	if game.CurrentPlayerId.Valid {
		player, _ := database.GetPlayer(game.CurrentPlayerId.UUID)
		currentPlayerName = player.Name
		isMyTurn = game.CurrentPlayerId.UUID == playerId
	}

	// Get winner name if game is finished
	var winnerName string
	if game.WinnerId.Valid {
		user, _ := database.GetUser(game.WinnerId.UUID)
		winnerName = user.Name
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/chronology.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Lobby             database.Lobby
		Game              database.ChronologyGame
		PlayerId          uuid.UUID
		CurrentPlayerName string
		IsMyTurn          bool
		WinnerName        string
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData:      basePageData,
		Lobby:             lobby,
		Game:              game,
		PlayerId:          playerId,
		CurrentPlayerName: currentPlayerName,
		IsMyTurn:          isMyTurn,
		WinnerName:        winnerName,
	})
}

// ChronologyLobbyAccess displays the access page for a Chronology game
func ChronologyLobbyAccess(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	if lobby.Id == uuid.Nil {
		http.Redirect(w, r, "/chronology/lobbies", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = "Card Judge - Chronology Access"

	hasLobbyAccess, err := database.UserHasLobbyAccess(basePageData.User.Id, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check lobby access"))
		return
	}

	if hasLobbyAccess {
		http.Redirect(w, r, fmt.Sprintf("/chronology/%s", lobbyId), http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/pages/base.html",
		"html/pages/body/lobby-access.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Lobby database.Lobby
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Lobby:        lobby,
	})
}
