package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/database"
)

type GrammarCheckResult struct {
	IsValid      bool   `json:"is_valid"`
	CorrectedText string `json:"corrected_text"`
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// CheckGrammarIfEnabled validates grammar only if enabled for the game
func CheckGrammarIfEnabled(ctx context.Context, lobbyID uuid.UUID, judgeCard string, responseCard string) (*GrammarCheckResult, error) {
	log.Println("[Grammar Check] Starting grammar check for lobby:", lobbyID)
	
	// Check if grammar checking is enabled for this lobby
	enabled, err := isGrammarCheckEnabled(ctx, lobbyID)
	if err != nil {
		log.Println("[Grammar Check] Error checking if enabled:", err)
		return &GrammarCheckResult{IsValid: true, CorrectedText: responseCard}, nil
	}
	
	if !enabled {
		log.Println("[Grammar Check] Grammar checking disabled for lobby:", lobbyID)
		return &GrammarCheckResult{IsValid: true, CorrectedText: responseCard}, nil
	}
	
	log.Println("[Grammar Check] Grammar checking enabled, calling checkGrammar")
	log.Println("[Grammar Check] Judge card:", judgeCard)
	log.Println("[Grammar Check] Response card:", responseCard)

	proposed := judgeCard + " " + responseCard
	return checkGrammar(judgeCard, responseCard, proposed)
}

func checkGrammar(judgeCard string, responseCard string, proposed string) (*GrammarCheckResult, error) {
	// Strict grammar correction: only fix verb tense and subject-verb agreement
	prompt := fmt.Sprintf(`Fix ONLY verb tense and subject-verb agreement errors. Do NOT rewrite or rephrase. Do NOT add any explanation or notes. If there are no errors, output the sentence unchanged.

Sentence: %s %s

Corrected:`, judgeCard, responseCard)

	req := ollamaRequest{
		Model:  "mistral:7b-instruct-v0.2-q4_0",
		Prompt: prompt,
		Stream: false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Println("[Grammar Check] Error marshaling request:", err)
		return nil, err
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	
	log.Println("[Grammar Check] Using Ollama host:", ollamaHost)

	// Use a longer timeout since models may need time to load
	client := &http.Client{Timeout: 30 * time.Second}
	fullURL := ollamaHost + "/api/generate"
	log.Println("[Grammar Check] Posting to:", fullURL)
	
	resp, err := client.Post(fullURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		// Ollama unavailable or timeout - graceful degradation
		log.Println("[Grammar Check] Ollama unavailable or timeout:", err)
		return &GrammarCheckResult{IsValid: true, CorrectedText: responseCard}, nil
	}
	defer resp.Body.Close()

	log.Println("[Grammar Check] Response status:", resp.Status)

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Println("[Grammar Check] Error decoding response:", err)
		return &GrammarCheckResult{IsValid: true, CorrectedText: responseCard}, nil
	}

	log.Println("[Grammar Check] Raw Ollama response:", ollamaResp.Response)
	
	correctedText := strings.TrimSpace(ollamaResp.Response)
	
	// If response is empty, treat as unavailable and degrade gracefully
	if correctedText == "" {
		log.Println("[Grammar Check] Empty response from Ollama - model may not be loaded")
		return &GrammarCheckResult{IsValid: true, CorrectedText: responseCard}, nil
	}
	
	// Strip off any explanatory notes (e.g., "(Note: ...)" or other parenthetical content)
	if idx := strings.Index(correctedText, " ("); idx != -1 {
		correctedText = strings.TrimSpace(correctedText[:idx])
	}
	
	isValid := correctedText == responseCard
	
	log.Println("[Grammar Check] Corrected text:", correctedText)
	log.Println("[Grammar Check] Is valid:", isValid)

	return &GrammarCheckResult{
		IsValid:       isValid,
		CorrectedText: correctedText,
	}, nil
}

func isGrammarCheckEnabled(ctx context.Context, lobbyID uuid.UUID) (bool, error) {
	return database.GetLobbyEnableLLMGrammarCheck(lobbyID)
}
