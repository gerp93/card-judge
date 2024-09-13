package apiPlayer

import (
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/database"
)

func GetGameInterfaceHtml(w http.ResponseWriter, r *http.Request) {
	playerIdString := r.PathValue("playerId")
	playerId, err := uuid.Parse(playerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get player id from path."))
		return
	}

	gameData, err := database.GetPlayerGameData(playerId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/game/game-interface.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse HTML."))
		return
	}

	tmpl.ExecuteTemplate(w, "game-interface", gameData)
}

func DrawPlayerHand(w http.ResponseWriter, r *http.Request) {
	playerIdString := r.PathValue("playerId")
	playerId, err := uuid.Parse(playerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get player id from path."))
		return
	}

	err = database.DrawPlayerHand(playerId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("&#9989;"))
}

func PlayPlayerCard(w http.ResponseWriter, r *http.Request) {
	playerIdString := r.PathValue("playerId")
	playerId, err := uuid.Parse(playerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get player id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get card id from path."))
		return
	}

	err = database.PlayPlayerCard(playerId, cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("&#9989;"))
}

func DiscardPlayerHand(w http.ResponseWriter, r *http.Request) {
	playerIdString := r.PathValue("playerId")
	playerId, err := uuid.Parse(playerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get player id from path."))
		return
	}

	err = database.DiscardPlayerHand(playerId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("&#9989;"))
}

func DiscardPlayerCard(w http.ResponseWriter, r *http.Request) {
	playerIdString := r.PathValue("playerId")
	playerId, err := uuid.Parse(playerIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get player id from path."))
		return
	}

	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get card id from path."))
		return
	}

	err = database.DiscardPlayerCard(playerId, cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("&#9989;"))
}
