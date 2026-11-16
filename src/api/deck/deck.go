package apiDeck

import (
	"encoding/csv"
	"net/http"

	"github.com/gerp93/gameshell-framework/api"
	gsDatabase "github.com/gerp93/gameshell-framework/database"
	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/database"
)

// GetCardExport streams a deck's cards as CSV. Deck management itself lives in
// the framework's api/deck; this card-specific export stays with the game.
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

	hasDeckAccess, err := gsDatabase.UserHasDeckAccess(userId, deckId)
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
