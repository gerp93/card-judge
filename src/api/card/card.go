package apiCard

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
	"github.com/grantfbarnes/card-judge/services"
	"github.com/grantfbarnes/card-judge/static"
)

func Find(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var lobbyId uuid.UUID
	var textSearch string
	for key, val := range r.Form {
		if key == "lobbyId" {
			lobbyId, err = uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse lobby id."))
				return
			}
		} else if key == "text" {
			textSearch = val[0]
		}
	}

	cards, err := database.FindDrawPileCard(lobbyId, textSearch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/components/table-rows/find-card-table-rows.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML."))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "find-card-table-rows", cards)
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var deckId uuid.UUID
	var category string
	var text string
	var youtube string
	for key, val := range r.Form {
		if key == "deckId" {
			deckId, err = uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse deck id."))
				return
			}
		} else if key == "category" {
			category = val[0]
		} else if key == "text" {
			text = val[0]
		} else if key == "youtube" {
			youtube = val[0]
		}
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

	text, err = processCardText(text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No text found."))
		return
	}

	existingCardId, err := database.GetCardId(deckId, text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if existingCardId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Card text already exists."))
		return
	}

	if len(youtube) != 0 && len(youtube) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid YouTube Video ID."))
		return
	}

	_, err = database.CreateCard(deckId, category, text, youtube)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusCreated)
}

func Update(w http.ResponseWriter, r *http.Request) {
	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var deckId uuid.UUID
	var category string
	var text string
	var youtube string
	for key, val := range r.Form {
		if key == "deckId" {
			deckId, err = uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse deck id."))
				return
			}
		} else if key == "category" {
			category = val[0]
		} else if key == "text" {
			text = val[0]
		} else if key == "youtube" {
			youtube = val[0]
		}
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

	text, err = processCardText(text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No text found."))
		return
	}

	existingCardId, err := database.GetCardId(deckId, text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if existingCardId != uuid.Nil && existingCardId != cardId {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Card text already exists."))
		return
	}

	if len(youtube) != 0 && len(youtube) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid YouTube Video ID."))
		return
	}

	err = database.UpdateCard(cardId, category, text, youtube)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetImage(w http.ResponseWriter, r *http.Request) {
	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	err = r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var deckId uuid.UUID
	for key, val := range r.Form {
		if key == "deckId" {
			deckId, err = uuid.Parse(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse deck id."))
				return
			}
		}
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

	var imageBytes []byte
	imageFile, _, err := r.FormFile("image")
	if err == nil {
		defer imageFile.Close()

		imageBytes, err = io.ReadAll(imageFile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Failed to get image bytes."))
			return
		}

		if len(imageBytes) > 65000 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Image cannot be over 65 KB in size"))
			return
		}
	}

	err = database.SetCardImage(cardId, imageBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	cardIdString := r.PathValue("cardId")
	cardId, err := uuid.Parse(cardIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id."))
		return
	}

	card, err := database.GetCard(cardId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get card."))
		return
	}

	hasDeckAccess, err := database.UserHasDeckAccess(userId, card.DeckId)
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

	err = database.DeleteCard(cardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Recover(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("Id")
	id, err := uuid.Parse(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get id from path."))
		return
	}

	err = database.RecoverCard(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func PermanentlyDelete(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("Id")
	id, err := uuid.Parse(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get id from path."))
		return
	}

	err = database.PermanentlyDeleteCard(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func processCardText(text string) (string, error) {
	normalizedText := text

	blankRegExp, err := regexp.Compile(`__+`)
	if err != nil {
		return normalizedText, err
	}

	normalizedText = blankRegExp.ReplaceAllString(text, "_____")
	normalizedText = strings.TrimSpace(normalizedText)

	return normalizedText, err
}

type ValidateCardRequest struct {
	LobbyID      string `json:"lobby_id"`
	JudgeCard    string `json:"judge_card"`
	ResponseCard string `json:"response_card"`
}

func ValidateCard(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("[ValidateCard] Failed to parse form:", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form"))
		return
	}

	lobbyID := r.FormValue("lobby_id")
	judgeCard := r.FormValue("judge_card")
	responseCard := r.FormValue("response_card")

	log.Println("[ValidateCard] Received request")
	log.Println("[ValidateCard] Lobby ID:", lobbyID)
	log.Println("[ValidateCard] Judge card:", judgeCard)
	log.Println("[ValidateCard] Response card:", responseCard)

	if lobbyID == "" || judgeCard == "" || responseCard == "" {
		log.Println("[ValidateCard] Missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing required fields"))
		return
	}

	lobbyIDParsed, err := uuid.Parse(lobbyID)
	if err != nil {
		log.Println("[ValidateCard] Failed to parse lobby ID:", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid lobby id"))
		return
	}

	log.Println("[ValidateCard] Calling CheckGrammarIfEnabled")
	result, _ := services.CheckGrammarIfEnabled(r.Context(), lobbyIDParsed, judgeCard, responseCard)
	log.Println("[ValidateCard] Grammar check result - IsValid:", result.IsValid, "CorrectedText:", result.CorrectedText)

	tmpl, err := template.ParseFS(
		static.StaticFiles,
		"html/components/game/response-card-grammar-view.html",
	)
	if err != nil {
		log.Println("[ValidateCard] Failed to parse HTML:", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to parse HTML"))
		return
	}

	type data struct {
		JudgeCard     string
		ResponseCard  string
		CorrectedText string
		IsValid       bool
	}

	_ = tmpl.ExecuteTemplate(w, "response-card-grammar-view", data{
		JudgeCard:     judgeCard,
		ResponseCard:  responseCard,
		CorrectedText: result.CorrectedText,
		IsValid:       result.IsValid,
	})
}
