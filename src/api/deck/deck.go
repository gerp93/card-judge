package apiDeck

import (
	"encoding/csv"
	"net/http"
	"text/template"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
)

func GetCardExport(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to check deck access."))
		return
	}

	if !hasDeckAccess {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	cards, err := database.GetCardsInDeckExport(deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()
	for _, card := range cards {
		_ = writer.Write([]string{card.Category, card.Text})
	}
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

	decks, err := database.SearchDecks(search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/table-rows/deck-table-rows.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML."))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "deck-table-rows", decks)
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var password string
	var passwordConfirm string
	var isPublicReadOnly bool
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		} else if key == "password" {
			password = val[0]
		} else if key == "passwordConfirm" {
			passwordConfirm = val[0]
		} else if key == "isPublicReadOnly" {
			isPublicReadOnly = val[0] == "1"
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No password found."))
		return
	}

	if password != passwordConfirm {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Passwords do not match."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	existingDeckId, err := database.GetDeckId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if existingDeckId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Deck name already exists."))
		return
	}

	id, err := database.CreateDeck(name, password, isPublicReadOnly)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = database.AddUserDeckAccess(userId, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Redirect", "/deck/"+id.String())
	w.WriteHeader(http.StatusCreated)
}

func SetName(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to check deck access."))
		return
	}

	if !hasDeckAccess {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
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

	existingDeckId, err := database.GetDeckId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if existingDeckId != uuid.Nil && existingDeckId != deckId {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Deck name already exists."))
		return
	}

	err = database.SetDeckName(deckId, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetPassword(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to check deck access."))
		return
	}

	if !hasDeckAccess {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var password string
	var passwordConfirm string
	for key, val := range r.Form {
		if key == "password" {
			password = val[0]
		} else if key == "passwordConfirm" {
			passwordConfirm = val[0]
		}
	}

	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No password found."))
		return
	}

	if password != passwordConfirm {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Passwords do not match."))
		return
	}

	err = database.SetDeckPassword(deckId, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetIsPublicReadOnly(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to check deck access."))
		return
	}

	if !hasDeckAccess {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var isPublicReadOnly bool
	for key, val := range r.Form {
		if key == "isPublicReadOnly" {
			isPublicReadOnly = val[0] == "1"
		}
	}

	err = database.SetIsPublicReadOnly(deckId, isPublicReadOnly)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to check deck access."))
		return
	}

	if !hasDeckAccess {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = database.DeleteDeck(deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Redirect", "/decks")
	w.WriteHeader(http.StatusOK)
}
