package apiAccess

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/database"
)

func Lobby(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	lobbyPasswordHash, err := database.GetLobbyPasswordHash(lobbyId)
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

	var password string
	for key, val := range r.Form {
		if key != "password" {
			continue
		}
		password = val[0]
		break
	}

	if !auth.PasswordMatchesHash(password, lobbyPasswordHash.String) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Provided password is not valid."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	err = database.AddUserLobbyAccess(userId, lobbyId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to add access."))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Deck(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	deckPasswordHash, err := database.GetDeckPasswordHash(deckId)
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

	var password string
	for key, val := range r.Form {
		if key != "password" {
			continue
		}
		password = val[0]
		break
	}

	if !auth.PasswordMatchesHash(password, deckPasswordHash) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Provided password is not valid."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	err = database.AddUserDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to add access."))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
