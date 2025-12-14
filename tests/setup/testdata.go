package setup

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/grantfbarnes/card-judge/tests/util"
)

// SeedTestData inserts minimal test data for screenshot testing
func SeedTestData(db *sql.DB) error {
	log.Println("Seeding test data...")

	// Create test users
	if err := createTestUsers(db); err != nil {
		return err
	}
	log.Println("✓ Created 4 test users (Test1-Test4)")

	// Get first user ID for deck/lobby ownership
	test1UserID := util.TestUser1ID

	// Create test decks (11+ triggers pagination)
	deckIDs, err := createTestDecks(db, test1UserID)
	if err != nil {
		return err
	}
	log.Println("✓ Created 11 test decks")

	// Create test cards in first deck
	if err := createTestCards(db, deckIDs[0]); err != nil {
		return err
	}
	log.Println("✓ Created test cards (1 prompt + 10 responses)")

	// Create test lobbies with players
	if err := createTestLobbies(db, deckIDs[0]); err != nil {
		return err
	}
	log.Println("✓ Created 2 test lobbies (with and without messages)")

	// Create stat data
	if err := createStatData(db, deckIDs[0]); err != nil {
		return err
	}
	log.Println("✓ Created stat data")

	log.Println("✓ Test data seeded successfully")
	return nil
}

func createTestUsers(db *sql.DB) error {
	// Password is "password" hashed with bcrypt
	passwordHash := "$2a$10$A45yhF.Cw5Kt0rEb9EUsheCAqIyv3P9vZxhhUUrRPV58okoGz5/eG"

	testUsers := []struct {
		id      string
		name    string
		isAdmin bool
	}{
		{util.TestUser1ID, util.TestUser1Name, true},
		{util.TestUser2ID, util.TestUser2Name, false},
		{util.TestUser3ID, util.TestUser3Name, false},
		{util.TestUser4ID, util.TestUser4Name, false},
	}

	for _, user := range testUsers {
		_, err := db.Exec(`
			INSERT INTO USER (ID, NAME, PASSWORD_HASH, IS_APPROVED, IS_ADMIN)
			VALUES (?, ?, ?, TRUE, ?)
			ON DUPLICATE KEY UPDATE NAME = ?
		`, user.id, user.name, passwordHash, user.isAdmin, user.name)
		if err != nil {
			return fmt.Errorf("failed to create test user %s: %w", user.name, err)
		}
	}
	return nil
}

func createTestDecks(db *sql.DB, userID string) ([]string, error) {
	var deckIDs []string
	for d := 0; d < util.TestDecksCount; d++ {
		deckID := fmt.Sprintf("%s%02d", util.TestDeckIDBase, d)
		deckName := fmt.Sprintf("Test Deck %d", d+1)

		_, err := db.Exec(`
			INSERT INTO DECK (ID, NAME, PASSWORD_HASH, IS_PUBLIC_READONLY)
			VALUES (?, ?, '', TRUE)
		`, deckID, deckName)
		if err != nil {
			return nil, fmt.Errorf("failed to create test deck %d: %w", d, err)
		}

		// Grant test user access to deck
		_, err = db.Exec(`
			INSERT INTO USER_ACCESS_DECK (USER_ID, DECK_ID)
			VALUES (?, ?)
		`, userID, deckID)
		if err != nil {
			return nil, fmt.Errorf("failed to grant deck access for deck %d: %w", d, err)
		}

		deckIDs = append(deckIDs, deckID)
	}
	return deckIDs, nil
}

func createTestCards(db *sql.DB, deckID string) error {
	// Create test prompt card
	promptCardID := util.TestCardID
	_, err := db.Exec(`
		INSERT INTO CARD (ID, DECK_ID, CATEGORY, TEXT)
		VALUES (?, ?, 'PROMPT', ?)
	`, promptCardID, deckID, util.TestPromptCardText)
	if err != nil {
		return fmt.Errorf("failed to create prompt card: %w", err)
	}

	// Create test response cards
	for i := 0; i < util.TestResponseCardsCount; i++ {
		responseCardID := fmt.Sprintf("%s%02d", util.TestResponseCardIDBase, i)
		_, err = db.Exec(`
			INSERT INTO CARD (ID, DECK_ID, CATEGORY, TEXT)
			VALUES (?, ?, 'RESPONSE', ?)
		`, responseCardID, deckID, fmt.Sprintf("Test response %d", i+1))
		if err != nil {
			return fmt.Errorf("failed to create response card %d: %w", i, err)
		}
	}
	return nil
}

func createTestLobbies(db *sql.DB, deckID string) error {
	// Test user IDs
	testUserIDs := []string{
		util.TestUser1ID,
		util.TestUser2ID,
		util.TestUser3ID,
		util.TestUser4ID,
	}

	// Create test players (base IDs for each user)
	testPlayers := []struct {
		playerIDBase string
		userID       string
	}{
		{util.TestPlayer1IDBase, testUserIDs[0]},
		{util.TestPlayer2IDBase, testUserIDs[1]},
		{util.TestPlayer3IDBase, testUserIDs[2]},
		{util.TestPlayer4IDBase, testUserIDs[3]},
	}

	// Create test lobbies - one with message, one without
	lobbies := []struct {
		id      string
		name    string
		message string
	}{
		{util.TestLobby1ID, "Test Lobby", util.TestLobby1Message},
		{util.TestLobby2ID, "Test Lobby (No Message)", ""},
	}

	for lobbyIdx, lobby := range lobbies {
		var messageValue interface{}
		if lobby.message == "" {
			messageValue = nil
		} else {
			messageValue = lobby.message
		}
		_, err := db.Exec(`
			INSERT INTO LOBBY (ID, NAME, MESSAGE)
			VALUES (?, ?, ?)
		`, lobby.id, lobby.name, messageValue)
		if err != nil {
			return fmt.Errorf("failed to create test lobby %s: %w", lobby.name, err)
		}

		// Add test deck cards to this lobby's draw pile
		_, err = db.Exec(`
			INSERT INTO DRAW_PILE (LOBBY_ID, CARD_ID)
			SELECT ?, ID FROM CARD WHERE DECK_ID = ?
		`, lobby.id, deckID)
		if err != nil {
			return fmt.Errorf("failed to add cards to draw pile for %s: %w", lobby.name, err)
		}

		// Add all 4 test players to this lobby
		for playerIdx, player := range testPlayers {
			// Create unique player ID by appending lobby and player indices
			uniquePlayerID := fmt.Sprintf("%s%d%d", player.playerIDBase, lobbyIdx, playerIdx)
			_, err = db.Exec(`
				INSERT INTO PLAYER (ID, LOBBY_ID, USER_ID)
				VALUES (?, ?, ?)
			`, uniquePlayerID, lobby.id, player.userID)
			if err != nil {
				return fmt.Errorf("failed to create player %d in lobby %s: %w", playerIdx, lobby.name, err)
			}
		}
	}
	return nil
}

func createStatData(db *sql.DB, deckID string) error {
	// Test user IDs
	testUserIDs := []string{
		util.TestUser1ID,
		util.TestUser2ID,
		util.TestUser3ID,
		util.TestUser4ID,
	}

	lobby1ID := util.TestLobby1ID
	playerID := util.TestPlayer1Lobby1ID

	// Create responses for stat tracking
	for i := 0; i < util.TestResponsesCount; i++ {
		responseID := fmt.Sprintf("%s%02d", util.TestResponseIDBase, i)
		_, err := db.Exec(`
			INSERT INTO RESPONSE (ID, PLAYER_ID, IS_REVEALED, IS_RULEDOUT)
			VALUES (?, ?, 1, 0)
		`, responseID, playerID)
		if err != nil {
			return fmt.Errorf("failed to create response: %w", err)
		}

		// Log a win for this response
		_, err = db.Exec(`
			INSERT INTO LOG_WIN (RESPONSE_ID)
			VALUES (?)
		`, responseID)
		if err != nil {
			return fmt.Errorf("failed to log win: %w", err)
		}
	}

	// Create card play logs
	responseCardID := fmt.Sprintf("%s00", util.TestResponseCardIDBase)
	judgeCardID := util.TestCardID
	playerCardID := fmt.Sprintf("%s01", util.TestResponseCardIDBase)

	for i := 0; i < util.TestResponseLogsCount; i++ {
		_, err := db.Exec(`
			INSERT INTO LOG_RESPONSE_CARD (LOBBY_ID, ROUND_ID, RESPONSE_ID, RESPONSE_CARD_ID, JUDGE_USER_ID, JUDGE_CARD_ID, PLAYER_USER_ID, PLAYER_CARD_ID)
			VALUES (?, UUID(), ?, ?, ?, ?, ?, ?)
		`, lobby1ID, fmt.Sprintf("%s0%d", util.TestResponseIDBase, i), responseCardID, judgeCardID, testUserIDs[1], testUserIDs[0], playerCardID)
		if err != nil {
			return fmt.Errorf("failed to log response card: %w", err)
		}
	}

	// Create review cards (11+ consecutive discards)
	reviewCards := []string{
		fmt.Sprintf("%s01", util.TestResponseCardIDBase),
		fmt.Sprintf("%s02", util.TestResponseCardIDBase),
		fmt.Sprintf("%s03", util.TestResponseCardIDBase),
	}

	for _, cardID := range reviewCards {
		// Create 12 discard logs for each review card
		for j := 0; j < util.TestDiscardsPerCardCount; j++ {
			_, err := db.Exec(`
				INSERT INTO LOG_DISCARD (LOBBY_ID, USER_ID, CARD_ID)
				VALUES (?, ?, ?)
			`, lobby1ID, testUserIDs[0], cardID)
			if err != nil {
				return fmt.Errorf("failed to log discard for review: %w", err)
			}
		}
	}

	// Create regular discards
	for i := 4; i < 6; i++ {
		discardCardID := fmt.Sprintf("%s%02d", util.TestResponseCardIDBase, i)
		_, err := db.Exec(`
			INSERT INTO LOG_DISCARD (LOBBY_ID, USER_ID, CARD_ID)
			VALUES (?, ?, ?)
		`, lobby1ID, testUserIDs[0], discardCardID)
		if err != nil {
			return fmt.Errorf("failed to log discard: %w", err)
		}
	}

	// Create skip logs
	for i := 0; i < util.TestSkipCardsCount; i++ {
		skipCardID := fmt.Sprintf("%s0%d", util.TestResponseCardIDBase, i)
		_, err := db.Exec(`
			INSERT INTO LOG_SKIP (LOBBY_ID, USER_ID, CARD_ID)
			VALUES (?, ?, ?)
		`, lobby1ID, testUserIDs[1], skipCardID)
		if err != nil {
			return fmt.Errorf("failed to log skip: %w", err)
		}
	}

	// Create credits spent logs
	_, err := db.Exec(`
		INSERT INTO LOG_CREDITS_SPENT (LOBBY_ID, USER_ID, AMOUNT, CATEGORY)
		VALUES (?, ?, 100, 'PURCHASE'), (?, ?, 50, 'GAMBLE'), (?, ?, 25, 'EXTRA-RESPONSE')
	`, lobby1ID, testUserIDs[0], lobby1ID, testUserIDs[1], lobby1ID, testUserIDs[2])
	if err != nil {
		return fmt.Errorf("failed to log credits spent: %w", err)
	}

	return nil
}
