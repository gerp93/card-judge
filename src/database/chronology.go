package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

// ChronologyGame represents a Chronology game instance
type ChronologyGame struct {
	Id              uuid.UUID
	LobbyId         uuid.UUID
	CreatedOnDate   time.Time
	CurrentPlayerId uuid.NullUUID
	GameStatus      string
	CardsToWin      int
	WinnerId        uuid.NullUUID
}

// ChronologyTimelineCard represents a card in a player's timeline
type ChronologyTimelineCard struct {
	Id          uuid.UUID
	CardId      uuid.UUID
	CardText    string
	CardYear    int
	Position    int
	PlacedOnDate time.Time
}

// ChronologyCurrentCard represents the current card being played
type ChronologyCurrentCard struct {
	CardId   uuid.UUID
	CardText string
	CardYear int
}

// ChronologyPlayer represents a player in a Chronology game with their timeline
type ChronologyPlayer struct {
	PlayerId     uuid.UUID
	UserId       uuid.UUID
	UserName     string
	IsActive     bool
	TimelineSize int
	IsCurrent    bool
}

// ChronologyPlayerTimeline represents a player with their full timeline for display
type ChronologyPlayerTimeline struct {
	PlayerId   uuid.UUID
	PlayerName string
	IsCurrent  bool
	IsMe       bool
	Timeline   []ChronologyTimelineCard
}

// GetChronologyGame retrieves the Chronology game for a lobby
func GetChronologyGame(lobbyId uuid.UUID) (ChronologyGame, error) {
	return getChronologyGameByColumn("LOBBY_ID", lobbyId)
}

// GetChronologyGameById retrieves the Chronology game by its ID
func GetChronologyGameById(gameId uuid.UUID) (ChronologyGame, error) {
	return getChronologyGameByColumn("ID", gameId)
}

// getChronologyGameByColumn is a helper to retrieve a game by a specific column
func getChronologyGameByColumn(column string, value uuid.UUID) (ChronologyGame, error) {
	var game ChronologyGame

	sqlString := fmt.Sprintf(`
		SELECT
			ID,
			LOBBY_ID,
			CREATED_ON_DATE,
			CURRENT_PLAYER_ID,
			GAME_STATUS,
			CARDS_TO_WIN,
			WINNER_ID
		FROM CHRONOLOGY_GAME
		WHERE %s = ?
	`, column)
	rows, err := query(sqlString, value)
	if err != nil {
		return game, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&game.Id,
			&game.LobbyId,
			&game.CreatedOnDate,
			&game.CurrentPlayerId,
			&game.GameStatus,
			&game.CardsToWin,
			&game.WinnerId,
		); err != nil {
			log.Println(err)
			return game, errors.New("failed to scan row in query results")
		}
	}

	return game, nil
}

// CreateChronologyGame creates a new Chronology game for a lobby
func CreateChronologyGame(lobbyId uuid.UUID, cardsToWin int) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	sqlString := `
		INSERT INTO CHRONOLOGY_GAME(
			ID,
			LOBBY_ID,
			CARDS_TO_WIN
		)
		VALUES (?, ?, ?)
	`
	return id, execute(sqlString, id, lobbyId, cardsToWin)
}

// CreateChronologyLobby creates a new lobby specifically for Chronology
func CreateChronologyLobby(name string, password string) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	var passwordHash sql.NullString
	if password != "" {
		hash, err := getPasswordHash(password)
		if err != nil {
			log.Println(err)
			return id, errors.New("failed to hash password")
		}
		passwordHash = sql.NullString{String: hash, Valid: true}
	}

	sqlString := `
		INSERT INTO LOBBY(
			ID,
			NAME,
			GAME_TYPE,
			PASSWORD_HASH,
			DRAW_PRIORITY,
			HAND_SIZE,
			FREE_CREDITS,
			WIN_STREAK_THRESHOLD,
			LOSE_STREAK_THRESHOLD
		)
		VALUES (?, ?, 'chronology', ?, 'RANDOM', 0, 0, 0, 0)
	`
	if passwordHash.Valid {
		return id, execute(sqlString, id, name, passwordHash.String)
	}
	return id, execute(sqlString, id, name, nil)
}

// InitializeChronologyDrawPile populates the draw pile with cards from decks
// Cards must have a year in their text (extracted via regex)
func InitializeChronologyDrawPile(gameId uuid.UUID, deckIds []uuid.UUID) error {
	if len(deckIds) == 0 {
		return errors.New("no decks provided")
	}

	// Build placeholders for deck IDs
	placeholders := ""
	args := make([]interface{}, 0, len(deckIds)+1)
	args = append(args, gameId)
	for i, deckId := range deckIds {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args = append(args, deckId)
	}

	// Get all cards from the decks (using PROMPT cards as Chronology events)
	sqlString := `
		INSERT INTO CHRONOLOGY_DRAW_PILE (ID, CHRONOLOGY_GAME_ID, CARD_ID, CARD_YEAR)
		SELECT UUID(), ?, C.ID, 0
		FROM CARD C
		WHERE C.DECK_ID IN (` + placeholders + `)
		AND C.CATEGORY = 'PROMPT'
	`
	return execute(sqlString, args...)
}

// ParseYearFromText extracts a 4-digit year from card text
// Returns 0 if no year found
func ParseYearFromText(text string) int {
	// Match 4-digit years (1000-2999)
	re := regexp.MustCompile(`\b([12]\d{3})\b`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		year, err := strconv.Atoi(matches[1])
		if err == nil {
			return year
		}
	}
	return 0
}

// UpdateDrawPileYears updates the years for all cards in the draw pile
// by parsing the year from each card's text
func UpdateDrawPileYears(gameId uuid.UUID) error {
	// Get all cards in draw pile
	sqlString := `
		SELECT DP.ID, C.TEXT
		FROM CHRONOLOGY_DRAW_PILE DP
		INNER JOIN CARD C ON C.ID = DP.CARD_ID
		WHERE DP.CHRONOLOGY_GAME_ID = ?
	`
	rows, err := query(sqlString, gameId)
	if err != nil {
		return err
	}
	defer rows.Close()

	type cardYear struct {
		id   uuid.UUID
		year int
	}
	updates := make([]cardYear, 0)

	for rows.Next() {
		var id uuid.UUID
		var text string
		if err := rows.Scan(&id, &text); err != nil {
			log.Println(err)
			continue
		}
		year := ParseYearFromText(text)
		if year > 0 {
			updates = append(updates, cardYear{id: id, year: year})
		}
	}

	// Update each card's year
	for _, u := range updates {
		sqlUpdate := `UPDATE CHRONOLOGY_DRAW_PILE SET CARD_YEAR = ? WHERE ID = ?`
		if err := execute(sqlUpdate, u.year, u.id); err != nil {
			log.Println(err)
		}
	}

	// Remove cards without valid years
	sqlDelete := `DELETE FROM CHRONOLOGY_DRAW_PILE WHERE CHRONOLOGY_GAME_ID = ? AND CARD_YEAR = 0`
	return execute(sqlDelete, gameId)
}

// DrawChronologyCard draws a random card from the draw pile and sets it as current
func DrawChronologyCard(gameId uuid.UUID) error {
	// Clear any existing current card
	sqlClear := `DELETE FROM CHRONOLOGY_CURRENT_CARD WHERE CHRONOLOGY_GAME_ID = ?`
	if err := execute(sqlClear, gameId); err != nil {
		return err
	}

	// Get a random undrawn card
	sqlDraw := `
		INSERT INTO CHRONOLOGY_CURRENT_CARD (ID, CHRONOLOGY_GAME_ID, CARD_ID, CARD_YEAR)
		SELECT UUID(), ?, CARD_ID, CARD_YEAR
		FROM CHRONOLOGY_DRAW_PILE
		WHERE CHRONOLOGY_GAME_ID = ? AND DRAWN = 0
		ORDER BY RAND()
		LIMIT 1
	`
	if err := execute(sqlDraw, gameId, gameId); err != nil {
		return err
	}

	// Mark the card as drawn
	sqlMark := `
		UPDATE CHRONOLOGY_DRAW_PILE
		SET DRAWN = 1
		WHERE CHRONOLOGY_GAME_ID = ?
		AND CARD_ID = (SELECT CARD_ID FROM CHRONOLOGY_CURRENT_CARD WHERE CHRONOLOGY_GAME_ID = ?)
	`
	return execute(sqlMark, gameId, gameId)
}

// GetChronologyCurrentCard gets the current card being played
func GetChronologyCurrentCard(gameId uuid.UUID) (ChronologyCurrentCard, error) {
	var card ChronologyCurrentCard

	sqlString := `
		SELECT CC.CARD_ID, C.TEXT, CC.CARD_YEAR
		FROM CHRONOLOGY_CURRENT_CARD CC
		INNER JOIN CARD C ON C.ID = CC.CARD_ID
		WHERE CC.CHRONOLOGY_GAME_ID = ?
	`
	rows, err := query(sqlString, gameId)
	if err != nil {
		return card, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&card.CardId, &card.CardText, &card.CardYear); err != nil {
			log.Println(err)
			return card, errors.New("failed to scan row in query results")
		}
	}

	return card, nil
}

// GetPlayerTimeline gets all cards in a player's timeline for a game
func GetPlayerTimeline(gameId uuid.UUID, playerId uuid.UUID) ([]ChronologyTimelineCard, error) {
	sqlString := `
		SELECT PT.ID, PT.CARD_ID, C.TEXT, PT.CARD_YEAR, PT.POSITION, PT.PLACED_ON_DATE
		FROM CHRONOLOGY_PLAYER_TIMELINE PT
		INNER JOIN CARD C ON C.ID = PT.CARD_ID
		WHERE PT.CHRONOLOGY_GAME_ID = ? AND PT.PLAYER_ID = ?
		ORDER BY PT.POSITION ASC
	`
	rows, err := query(sqlString, gameId, playerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]ChronologyTimelineCard, 0)
	for rows.Next() {
		var card ChronologyTimelineCard
		if err := rows.Scan(
			&card.Id,
			&card.CardId,
			&card.CardText,
			&card.CardYear,
			&card.Position,
			&card.PlacedOnDate,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, card)
	}

	return result, nil
}

// GetAllPlayerTimelines gets all players' timelines for a game, ordered with current player first
func GetAllPlayerTimelines(gameId uuid.UUID, currentPlayerId uuid.UUID, viewingPlayerId uuid.UUID) ([]ChronologyPlayerTimeline, error) {
	// Get all active players
	players, err := GetChronologyPlayers(gameId)
	if err != nil {
		return nil, err
	}

	result := make([]ChronologyPlayerTimeline, 0, len(players))
	
	// First add current player
	for _, p := range players {
		if p.IsActive && p.PlayerId == currentPlayerId {
			timeline, err := GetPlayerTimeline(gameId, p.PlayerId)
			if err != nil {
				timeline = []ChronologyTimelineCard{}
			}
			result = append(result, ChronologyPlayerTimeline{
				PlayerId:   p.PlayerId,
				PlayerName: p.UserName,
				IsCurrent:  true,
				IsMe:       p.PlayerId == viewingPlayerId,
				Timeline:   timeline,
			})
			break
		}
	}

	// Then add other players in order
	for _, p := range players {
		if p.IsActive && p.PlayerId != currentPlayerId {
			timeline, err := GetPlayerTimeline(gameId, p.PlayerId)
			if err != nil {
				timeline = []ChronologyTimelineCard{}
			}
			result = append(result, ChronologyPlayerTimeline{
				PlayerId:   p.PlayerId,
				PlayerName: p.UserName,
				IsCurrent:  false,
				IsMe:       p.PlayerId == viewingPlayerId,
				Timeline:   timeline,
			})
		}
	}

	return result, nil
}

// PlaceCardInTimeline attempts to place the current card in a player's timeline
// Returns true if placement was correct, false otherwise
func PlaceCardInTimeline(gameId uuid.UUID, playerId uuid.UUID, position int) (bool, error) {
	// Get the current card
	currentCard, err := GetChronologyCurrentCard(gameId)
	if err != nil {
		return false, err
	}
	if currentCard.CardId == uuid.Nil {
		return false, errors.New("no current card to place")
	}

	// Get player's current timeline
	timeline, err := GetPlayerTimeline(gameId, playerId)
	if err != nil {
		return false, err
	}

	// Validate position (must be between 0 and len(timeline))
	if position < 0 || position > len(timeline) {
		return false, errors.New("invalid position")
	}

	// Check if placement is correct
	correct := true
	if position > 0 {
		// Card before this position must have year <= current card's year
		if timeline[position-1].CardYear > currentCard.CardYear {
			correct = false
		}
	}
	if position < len(timeline) {
		// Card after this position must have year >= current card's year
		if timeline[position].CardYear < currentCard.CardYear {
			correct = false
		}
	}

	if correct {
		// Shift existing cards to make room
		sqlShift := `
			UPDATE CHRONOLOGY_PLAYER_TIMELINE
			SET POSITION = POSITION + 1
			WHERE CHRONOLOGY_GAME_ID = ? AND PLAYER_ID = ? AND POSITION >= ?
		`
		if err := execute(sqlShift, gameId, playerId, position); err != nil {
			return false, err
		}

		// Insert the new card
		id, err := uuid.NewUUID()
		if err != nil {
			return false, err
		}
		sqlInsert := `
			INSERT INTO CHRONOLOGY_PLAYER_TIMELINE (ID, CHRONOLOGY_GAME_ID, PLAYER_ID, CARD_ID, CARD_YEAR, POSITION)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		if err := execute(sqlInsert, id, gameId, playerId, currentCard.CardId, currentCard.CardYear, position); err != nil {
			return false, err
		}
	}

	// Clear current card
	sqlClear := `DELETE FROM CHRONOLOGY_CURRENT_CARD WHERE CHRONOLOGY_GAME_ID = ?`
	if err := execute(sqlClear, gameId); err != nil {
		return correct, err
	}

	return correct, nil
}

// GetChronologyPlayers gets all players in a Chronology game with their timeline sizes
func GetChronologyPlayers(gameId uuid.UUID) ([]ChronologyPlayer, error) {
	sqlString := `
		SELECT 
			P.ID,
			P.USER_ID,
			U.NAME,
			P.IS_ACTIVE,
			COALESCE(
				(SELECT COUNT(*) FROM CHRONOLOGY_PLAYER_TIMELINE PT 
				 WHERE PT.CHRONOLOGY_GAME_ID = CG.ID AND PT.PLAYER_ID = P.ID),
				0
			) AS TIMELINE_SIZE,
			CASE WHEN CG.CURRENT_PLAYER_ID = P.ID THEN 1 ELSE 0 END AS IS_CURRENT
		FROM CHRONOLOGY_GAME CG
		INNER JOIN LOBBY L ON L.ID = CG.LOBBY_ID
		INNER JOIN PLAYER P ON P.LOBBY_ID = L.ID
		INNER JOIN USER U ON U.ID = P.USER_ID
		WHERE CG.ID = ?
		ORDER BY P.JOIN_ORDER ASC
	`
	rows, err := query(sqlString, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]ChronologyPlayer, 0)
	for rows.Next() {
		var player ChronologyPlayer
		if err := rows.Scan(
			&player.PlayerId,
			&player.UserId,
			&player.UserName,
			&player.IsActive,
			&player.TimelineSize,
			&player.IsCurrent,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, player)
	}

	return result, nil
}

// SetChronologyCurrentPlayer sets whose turn it is
func SetChronologyCurrentPlayer(gameId uuid.UUID, playerId uuid.UUID) error {
	sqlString := `UPDATE CHRONOLOGY_GAME SET CURRENT_PLAYER_ID = ? WHERE ID = ?`
	return execute(sqlString, playerId, gameId)
}

// AdvanceChronologyTurn moves to the next active player
func AdvanceChronologyTurn(gameId uuid.UUID) error {
	players, err := GetChronologyPlayers(gameId)
	if err != nil {
		return err
	}

	// Find current player index
	currentIdx := -1
	for i, p := range players {
		if p.IsCurrent {
			currentIdx = i
			break
		}
	}

	// Find next active player
	nextIdx := currentIdx
	for i := 0; i < len(players); i++ {
		nextIdx = (nextIdx + 1) % len(players)
		if players[nextIdx].IsActive {
			break
		}
	}

	if nextIdx < len(players) {
		return SetChronologyCurrentPlayer(gameId, players[nextIdx].PlayerId)
	}

	return errors.New("no active players found")
}

// StartChronologyGame starts the game by setting status and first player
func StartChronologyGame(gameId uuid.UUID) error {
	players, err := GetChronologyPlayers(gameId)
	if err != nil {
		return err
	}

	if len(players) == 0 {
		return errors.New("no players in game")
	}

	// Deal one card to each player to start their timeline
	for _, player := range players {
		if player.IsActive {
			// Draw a card from the pile
			var cardId uuid.UUID
			var cardYear int
			sqlGetCard := `
				SELECT CARD_ID, CARD_YEAR
				FROM CHRONOLOGY_DRAW_PILE
				WHERE CHRONOLOGY_GAME_ID = ? AND DRAWN = 0
				ORDER BY RAND()
				LIMIT 1
			`
			rows, err := query(sqlGetCard, gameId)
			if err != nil {
				return err
			}
			defer rows.Close()
			
			if rows.Next() {
				if err := rows.Scan(&cardId, &cardYear); err != nil {
					return err
				}
			} else {
				return errors.New("not enough cards to deal initial cards")
			}

			// Mark card as drawn
			sqlMarkDrawn := `UPDATE CHRONOLOGY_DRAW_PILE SET DRAWN = 1 WHERE CHRONOLOGY_GAME_ID = ? AND CARD_ID = ?`
			if err := execute(sqlMarkDrawn, gameId, cardId); err != nil {
				return err
			}

			// Add to player's timeline at position 0
			id, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			sqlAddToTimeline := `
				INSERT INTO CHRONOLOGY_PLAYER_TIMELINE (ID, CHRONOLOGY_GAME_ID, PLAYER_ID, CARD_ID, CARD_YEAR, POSITION)
				VALUES (?, ?, ?, ?, ?, 0)
			`
			if err := execute(sqlAddToTimeline, id, gameId, player.PlayerId, cardId, cardYear); err != nil {
				return err
			}
		}
	}

	// Find first active player
	var firstPlayer uuid.UUID
	for _, p := range players {
		if p.IsActive {
			firstPlayer = p.PlayerId
			break
		}
	}

	if firstPlayer == uuid.Nil {
		return errors.New("no active players")
	}

	// Set game as active and set first player
	sqlString := `UPDATE CHRONOLOGY_GAME SET GAME_STATUS = 'active', CURRENT_PLAYER_ID = ? WHERE ID = ?`
	if err := execute(sqlString, firstPlayer, gameId); err != nil {
		return err
	}

	// Draw first card for play
	return DrawChronologyCard(gameId)
}

// ResetChronologyGame resets a finished game to play again
func ResetChronologyGame(gameId uuid.UUID) error {
	// Clear all player timelines
	sqlClearTimelines := `DELETE FROM CHRONOLOGY_PLAYER_TIMELINE WHERE CHRONOLOGY_GAME_ID = ?`
	if err := execute(sqlClearTimelines, gameId); err != nil {
		return err
	}

	// Clear current card
	sqlClearCurrentCard := `DELETE FROM CHRONOLOGY_CURRENT_CARD WHERE CHRONOLOGY_GAME_ID = ?`
	if err := execute(sqlClearCurrentCard, gameId); err != nil {
		return err
	}

	// Reset draw pile - mark all cards as not drawn
	sqlResetDrawPile := `UPDATE CHRONOLOGY_DRAW_PILE SET DRAWN = 0 WHERE CHRONOLOGY_GAME_ID = ?`
	if err := execute(sqlResetDrawPile, gameId); err != nil {
		return err
	}

	// Reset game status to waiting
	sqlResetGame := `UPDATE CHRONOLOGY_GAME SET GAME_STATUS = 'waiting', CURRENT_PLAYER_ID = NULL, WINNER_ID = NULL WHERE ID = ?`
	if err := execute(sqlResetGame, gameId); err != nil {
		return err
	}

	return nil
}

// CheckChronologyWinner checks if any player has won
func CheckChronologyWinner(gameId uuid.UUID) (uuid.UUID, error) {
	game, err := GetChronologyGameById(gameId)
	if err != nil {
		return uuid.Nil, err
	}

	players, err := GetChronologyPlayers(gameId)
	if err != nil {
		return uuid.Nil, err
	}

	for _, p := range players {
		if p.TimelineSize >= game.CardsToWin {
			// Set winner
			sqlString := `UPDATE CHRONOLOGY_GAME SET GAME_STATUS = 'finished', WINNER_ID = ? WHERE ID = ?`
			if err := execute(sqlString, p.UserId, gameId); err != nil {
				return uuid.Nil, err
			}
			return p.UserId, nil
		}
	}

	return uuid.Nil, nil
}

// GetChronologyDrawPileCount returns the number of cards remaining in the draw pile
func GetChronologyDrawPileCount(gameId uuid.UUID) (int, error) {
	sqlString := `
		SELECT COUNT(*)
		FROM CHRONOLOGY_DRAW_PILE
		WHERE CHRONOLOGY_GAME_ID = ? AND DRAWN = 0
	`
	rows, err := query(sqlString, gameId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println(err)
			return 0, errors.New("failed to scan row in query results")
		}
	}

	return count, nil
}

// ChronologyLobbyDetails represents a Chronology lobby for listing
type ChronologyLobbyDetails struct {
	Id          uuid.UUID
	Name        string
	PlayerCount int
	GameStatus  string
	HasPassword bool
}

// SearchChronologyLobbies searches for Chronology-type lobbies
func SearchChronologyLobbies(name string, page int) ([]ChronologyLobbyDetails, error) {
	name = "%" + name + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			L.ID,
			L.NAME,
			L.PASSWORD_HASH IS NOT NULL AS HAS_PASSWORD,
			COALESCE(CG.GAME_STATUS, 'waiting') AS GAME_STATUS,
			COUNT(P.ID) AS PLAYER_COUNT
		FROM LOBBY AS L
			LEFT JOIN CHRONOLOGY_GAME AS CG ON CG.LOBBY_ID = L.ID
			LEFT JOIN PLAYER AS P ON P.LOBBY_ID = L.ID AND P.IS_ACTIVE = 1
		WHERE L.NAME LIKE ?
			AND L.GAME_TYPE = 'chronology'
		GROUP BY L.ID
		ORDER BY L.CREATED_ON_DATE DESC
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, name, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]ChronologyLobbyDetails, 0)
	for rows.Next() {
		var ld ChronologyLobbyDetails
		if err := rows.Scan(
			&ld.Id,
			&ld.Name,
			&ld.HasPassword,
			&ld.GameStatus,
			&ld.PlayerCount,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, ld)
	}
	return result, nil
}

// CountChronologyLobbies counts Chronology-type lobbies matching name
func CountChronologyLobbies(name string) (int, error) {
	name = "%" + name + "%"

	sqlString := `
		SELECT COUNT(*)
		FROM LOBBY
		WHERE NAME LIKE ? AND GAME_TYPE = 'chronology'
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println(err)
			return 0, errors.New("failed to scan row in query results")
		}
	}

	return count, nil
}

// Helper function to get password hash (reuse from auth package)
func getPasswordHash(password string) (string, error) {
	// Use the auth package's password hashing
	return auth.GetPasswordHash(password)
}
