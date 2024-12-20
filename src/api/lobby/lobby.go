package apiLobby

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/websocket"
)

func GetGameInterface(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func Search(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var search string
	for key, val := range r.Form {
		if key == "search" {
			search = val[0]
		}
	}

	search = "%" + search + "%"

	lobbies, err := database.SearchLobbies(search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/table-rows/lobby-table-rows.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML."))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "lobby-table-rows", lobbies)
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var message string
	var password string
	var passwordConfirm string
	var handSize int
	var creditLimit int
	var winStreakThreshold int
	var loseStreakThreshold int
	var deckIds = make([]uuid.UUID, 0)
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		} else if key == "message" {
			message = val[0]
		} else if key == "password" {
			password = val[0]
		} else if key == "passwordConfirm" {
			passwordConfirm = val[0]
		} else if key == "handSize" {
			handSize, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse hand size."))
				return
			}
		} else if key == "creditLimit" {
			creditLimit, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse credit limit."))
				return
			}
		} else if key == "winStreakThreshold" {
			winStreakThreshold, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse win streak threshold."))
				return
			}
		} else if key == "loseStreakThreshold" {
			loseStreakThreshold, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse lose streak threshold."))
				return
			}
		} else if strings.HasPrefix(key, "deckId") {
			deckId, err := uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse deck id."))
				return
			}
			deckIds = append(deckIds, deckId)
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if password != "" {
		if password != passwordConfirm {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Passwords do not match."))
			return
		}
	}

	if handSize < 6 {
		handSize = 6
	}

	if handSize > 16 {
		handSize = 16
	}

	if creditLimit < 0 {
		creditLimit = 0
	}

	if creditLimit > 10 {
		creditLimit = 10
	}

	if winStreakThreshold < 1 {
		winStreakThreshold = 1
	}

	if winStreakThreshold > 5 {
		winStreakThreshold = 5
	}

	if loseStreakThreshold < 1 {
		loseStreakThreshold = 1
	}

	if loseStreakThreshold > 5 {
		loseStreakThreshold = 5
	}

	if len(deckIds) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("At least one deck is required."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	existingLobbyId, err := database.GetLobbyId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if existingLobbyId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Lobby name already exists."))
		return
	}

	lobbyId, err := database.CreateLobby(name, message, password, handSize, creditLimit, winStreakThreshold, loseStreakThreshold)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.AddCardsToLobby(lobbyId, deckIds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.AddUserLobbyAccess(userId, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Redirect", "/lobby/"+lobbyId.String())
	w.WriteHeader(http.StatusCreated)
}

func DrawHand(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.DrawHand(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func PlayCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.PlayCard(player.Id, cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PurchaseCredits(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.PurchaseCredits(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Attempted to purchase credits for an unfair advantage... Everyone else receives a credit as a result.")
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<b>Shame on you.</b><br/><br/>This action has been reported in the lobby chat and everyone else has received a credit."))
}

func GambleCredits(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var credits int
	for key, val := range r.Form {
		if key == "credits" {
			credits, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse credits."))
				return
			}
		}
	}

	if credits < 1 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No credits provided."))
		return
	}

	if lobby.CreditLimit-player.CreditsSpent < credits {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("You do not have that many credits to gamble."))
		return
	}

	err = database.GambleCredits(player.Id, credits)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.PlayerBroadcast(player.Id, "refresh")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func BetOnWin(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if player.BetOnWin > 0 {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte(fmt.Sprintf("A bet of %d has already been placed.", player.BetOnWin)))
		return
	}

	lobby, err := database.GetLobby(lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var credits int
	for key, val := range r.Form {
		if key == "credits" {
			credits, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse credits."))
				return
			}
		}
	}

	if credits < 1 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No credits provided."))
		return
	}

	if lobby.CreditLimit-player.CreditsSpent < credits {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("You do not have that many credits to bet."))
		return
	}

	err = database.BetOnWin(player.Id, credits)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.PlayerBroadcast(player.Id, "refresh")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func AddExtraResponse(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.AddExtraResponse(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PlayStealCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.PlayStealCard(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PlaySurpriseCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.PlaySurpriseCard(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PlayFindCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var cardId uuid.UUID
	for key, val := range r.Form {
		if key == "cardId" {
			cardId, err = uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse card id."))
				return
			}
		}
	}

	err = database.PlayFindCard(player.Id, cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PlayWildCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var text string
	for key, val := range r.Form {
		if key == "text" {
			text = val[0]
		}
	}

	existingCardId, err := database.GetCardId(lobbyId, text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if existingCardId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Wild card text has already been played."))
		return
	}

	err = database.PlayWildCard(player.Id, text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func WithdrawCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	responseCardIdString := r.PathValue("responseCardId")
	responseCardId, err := uuid.Parse(responseCardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get response card id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.WithdrawCard(responseCardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func DiscardCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.DiscardCard(player.Id, cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func LockCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.LockCard(player.Id, cardId, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func UnlockCard(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.LockCard(player.Id, cardId, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func VoteToKick(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	subjectPlayerIdString := r.PathValue("playerId")
	subjectPlayerId, err := uuid.Parse(subjectPlayerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get player id from path."))
		return
	}

	subjectPlayer, err := database.GetPlayer(subjectPlayerId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	isKicked, err := database.VoteToKick(player.Id, subjectPlayer.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Voted to kick <green>"+subjectPlayer.Name+"</> out of the lobby")

	if isKicked {
		websocket.LobbyBroadcast(lobbyId, "<red>Player Kicked</>: <green>"+subjectPlayer.Name+"</>")
		websocket.PlayerBroadcast(subjectPlayerId, "kick")
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func VoteToKickUndo(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	subjectPlayerIdString := r.PathValue("playerId")
	subjectPlayerId, err := uuid.Parse(subjectPlayerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get player id from path."))
		return
	}

	subjectPlayer, err := database.GetPlayer(subjectPlayerId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.VoteToKickUndo(player.Id, subjectPlayer.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Removed their vote to kick <green>"+subjectPlayer.Name+"</> out of the lobby")

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func RevealResponse(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	responseIdString := r.PathValue("responseId")
	responseId, err := uuid.Parse(responseIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get response id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.RevealResponse(responseId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func ToggleRuleOutResponse(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	responseIdString := r.PathValue("responseId")
	responseId, err := uuid.Parse(responseIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get response id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.ToggleRuleOutResponse(responseId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PickWinner(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	responseIdString := r.PathValue("responseId")
	responseId, err := uuid.Parse(responseIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get response id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	cardTextStart, err := database.GetResponseCardTextStart(responseId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<blue>Winning Card</>: "+cardTextStart)

	winnerName, err := database.PickWinner(responseId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<blue>Winner</>: <green>"+winnerName+"</>")

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func PickRandomWinner(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Random Winner!")

	winnerName, err := database.PickRandomWinner(lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if winnerName == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Could not find a random response winner that isn't ruled out."))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<blue>Winner</>: <green>"+winnerName+"</>")

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func DiscardHand(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.DiscardHand(player.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	writeGameInterfaceHtml(w, player.Id)
}

func FlipTable(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: FLIP THE TABLE!")

	w.Header().Add("HX-Redirect", "/lobbies")
	w.WriteHeader(http.StatusOK)
}

func SkipPrompt(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.SkipPrompt(lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")
	w.WriteHeader(http.StatusOK)
}

func SetName(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	existingLobbyId, err := database.GetLobbyId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if existingLobbyId != uuid.Nil && existingLobbyId != lobbyId {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Lobby name already exists."))
		return
	}

	err = database.SetLobbyName(lobbyId, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Lobby name set to "+name)
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetMessage(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var message string
	for key, val := range r.Form {
		if key == "message" {
			message = val[0]
		}
	}

	err = database.SetLobbyMessage(lobbyId, message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "<green>"+player.Name+"</>: Lobby message set to "+message)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetHandSize(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var handSize int
	for key, val := range r.Form {
		if key == "handSize" {
			handSize, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse hand size."))
				return
			}
		}
	}

	if handSize < 6 {
		handSize = 6
	}

	if handSize > 16 {
		handSize = 16
	}

	err = database.SetLobbyHandSize(lobbyId, handSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Lobby hand size set to %d", player.Name, handSize))
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetCreditLimit(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var creditLimit int
	for key, val := range r.Form {
		if key == "creditLimit" {
			creditLimit, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse credit limit."))
				return
			}
		}
	}

	if creditLimit < 0 {
		creditLimit = 0
	}

	if creditLimit > 10 {
		creditLimit = 10
	}

	err = database.SetLobbyCreditLimit(lobbyId, creditLimit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Lobby credit limit set to %d", player.Name, creditLimit))
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetWinStreakThreshold(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var winStreakThreshold int
	for key, val := range r.Form {
		if key == "winStreakThreshold" {
			winStreakThreshold, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse win streak threshold."))
				return
			}
		}
	}

	if winStreakThreshold < 1 {
		winStreakThreshold = 1
	}

	if winStreakThreshold > 5 {
		winStreakThreshold = 5
	}

	err = database.SetLobbyWinStreakThreshold(lobbyId, winStreakThreshold)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Lobby win streak threshold set to %d", player.Name, winStreakThreshold))
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetLoseStreakThreshold(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	player, err := getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var loseStreakThreshold int
	for key, val := range r.Form {
		if key == "loseStreakThreshold" {
			loseStreakThreshold, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse lose streak threshold."))
				return
			}
		}
	}

	if loseStreakThreshold < 1 {
		loseStreakThreshold = 1
	}

	if loseStreakThreshold > 5 {
		loseStreakThreshold = 5
	}

	err = database.SetLobbyLoseStreakThreshold(lobbyId, loseStreakThreshold)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Lobby lose streak threshold set to %d", player.Name, loseStreakThreshold))
	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func SetResponseCount(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	_, err = getLobbyRequestPlayer(r, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var responseCount int
	for key, val := range r.Form {
		if key == "responseCount" {
			responseCount, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse response count."))
				return
			}
		}
	}

	if responseCount < 1 {
		responseCount = 1
	}

	if responseCount > 3 {
		responseCount = 3
	}

	err = database.SetJudgeResponseCount(lobbyId, responseCount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	websocket.LobbyBroadcast(lobbyId, "refresh")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func getLobbyRequestPlayer(r *http.Request, lobbyId uuid.UUID) (database.Player, error) {
	var player database.Player

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		return player, errors.New("failed to get user id")
	}

	player, err := database.GetLobbyUserPlayer(lobbyId, userId)
	if err != nil {
		return player, err
	}

	if player.Id == uuid.Nil {
		return player, errors.New("user not found in lobby")
	}

	return player, nil
}

func writeGameInterfaceHtml(w http.ResponseWriter, playerId uuid.UUID) {
	gameData, err := database.GetPlayerGameData(playerId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/game/game-interface.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML."))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "game-interface", gameData)
}
