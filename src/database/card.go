package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type Card struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	DeckId   uuid.UUID
	Category string
	Text     string
	YouTube  sql.NullString
	Image    sql.NullString
}

type DisplayCard struct {
	Id            uuid.UUID
	CreatedOnDate time.Time

	DeckName string
	Category string
	Text     string
	YouTube  sql.NullString
	Image    sql.NullString
}

type LobbyCard struct {
	LobbyId uuid.UUID
	Card
}

func SearchCardsInDeck(deckId uuid.UUID, category string, text string, page int) ([]Card, error) {
	if category == "" {
		category = "%"
	}

	text = "%" + text + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			C.ID,
			C.CREATED_ON_DATE,
			C.CHANGED_ON_DATE,
			C.DECK_ID,
			C.CATEGORY,
			C.TEXT,
			C.YOUTUBE,
			C.IMAGE
		FROM CARD AS C
		WHERE C.DECK_ID = ?
			AND C.CATEGORY LIKE ?
			AND C.TEXT LIKE ?
		ORDER BY C.CHANGED_ON_DATE DESC,
			C.TEXT ASC
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, deckId, category, text, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Card, 0)
	for rows.Next() {
		var card Card
		var imageBytes []byte
		if err := rows.Scan(
			&card.Id,
			&card.CreatedOnDate,
			&card.ChangedOnDate,
			&card.DeckId,
			&card.Category,
			&card.Text,
			&card.YouTube,
			&imageBytes,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}

		result = append(result, card)
	}
	return result, nil
}

func CountCardsInDeck(deckId uuid.UUID, category string, text string) (int, error) {
	if category == "" {
		category = "%"
	}

	text = "%" + text + "%"

	sqlString := `
		SELECT
			COUNT(*)
		FROM CARD AS C
		WHERE C.DECK_ID = ?
			AND C.CATEGORY LIKE ?
			AND C.TEXT LIKE ?
	`
	rows, err := query(sqlString, deckId, category, text)
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

func SearchCardsInReview(page int) ([]DisplayCard, error) {
	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			RC.ID,
			RC.CREATED_ON_DATE,
			D.NAME AS DECK_NAME,
			RC.CATEGORY,
			RC.TEXT,
			RC.YOUTUBE,
			RC.IMAGE
		FROM REVIEW_CARD AS RC
			INNER JOIN DECK AS D ON D.ID = RC.DECK_ID
		ORDER BY RC.CREATED_ON_DATE
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]DisplayCard, 0)
	for rows.Next() {
		var card DisplayCard
		var imageBytes []byte
		if err := rows.Scan(
			&card.Id,
			&card.CreatedOnDate,
			&card.DeckName,
			&card.Category,
			&card.Text,
			&card.YouTube,
			&imageBytes,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}

		result = append(result, card)
	}
	return result, nil
}

func CountCardsInReview() (int, error) {
	sqlString := `
		SELECT
			COUNT(*)
		FROM REVIEW_CARD AS RC
	`
	rows, err := query(sqlString)
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

func SearchCardsWithAccess(userId uuid.UUID, deckName string, category string, text string, page int) ([]DisplayCard, error) {
	deckName = "%" + deckName + "%"
	if category == "" {
		category = "%"
	}
	text = "%" + text + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			C.ID,
			C.CREATED_ON_DATE,
			D.NAME AS DECK_NAME,
			C.CATEGORY,
			C.TEXT,
			C.YOUTUBE,
			C.IMAGE
		FROM CARD AS C
			INNER JOIN DECK AS D ON D.ID = C.DECK_ID
		WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
			AND D.NAME LIKE ?
			AND C.CATEGORY LIKE ?
			AND C.TEXT LIKE ?
		ORDER BY C.CHANGED_ON_DATE DESC,
			C.TEXT ASC
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, userId, deckName, category, text, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]DisplayCard, 0)
	for rows.Next() {
		var card DisplayCard
		var imageBytes []byte
		if err := rows.Scan(
			&card.Id,
			&card.CreatedOnDate,
			&card.DeckName,
			&card.Category,
			&card.Text,
			&card.YouTube,
			&imageBytes,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}

		result = append(result, card)
	}
	return result, nil
}

func CountCardsWithAccess(userId uuid.UUID, deckName string, category string, text string) (int, error) {
	deckName = "%" + deckName + "%"
	if category == "" {
		category = "%"
	}
	text = "%" + text + "%"

	sqlString := `
		SELECT
			COUNT(*)
		FROM CARD AS C
			INNER JOIN DECK AS D ON D.ID = C.DECK_ID
		WHERE FN_USER_HAS_DECK_ACCESS(?, C.DECK_ID)
			AND D.NAME LIKE ?
			AND C.CATEGORY LIKE ?
			AND C.TEXT LIKE ?
	`
	rows, err := query(sqlString, userId, deckName, category, text)
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

func FindDrawPileCard(lobbyId uuid.UUID, text string) ([]LobbyCard, error) {
	sqlString := `
		SELECT
			LOBBY_ID,
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			DECK_ID,
			CATEGORY,
			TEXT,
			YOUTUBE,
			IMAGE
		FROM (
				SELECT
					MATCH (TEXT) AGAINST(? IN NATURAL LANGUAGE MODE) AS SCORE,
					DP.LOBBY_ID,
					C.ID,
					C.CREATED_ON_DATE,
					C.CHANGED_ON_DATE,
					C.DECK_ID,
					C.CATEGORY,
					C.TEXT,
					C.YOUTUBE,
					C.IMAGE
				FROM CARD AS C
					INNER JOIN DRAW_PILE AS DP ON DP.CARD_ID = C.ID
				WHERE DP.LOBBY_ID = ?
					AND C.CATEGORY = 'RESPONSE'
					AND MATCH (TEXT) AGAINST(? IN NATURAL LANGUAGE MODE)
			) AS T
		ORDER BY SCORE DESC
		LIMIT 10
	`
	rows, err := query(sqlString, text, lobbyId, text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]LobbyCard, 0)
	for rows.Next() {
		var card LobbyCard
		var imageBytes []byte
		if err := rows.Scan(
			&card.LobbyId,
			&card.Id,
			&card.CreatedOnDate,
			&card.ChangedOnDate,
			&card.DeckId,
			&card.Category,
			&card.Text,
			&card.YouTube,
			&imageBytes); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}

		result = append(result, card)
	}
	return result, nil
}

func GetCardsInDeckExport(deckId uuid.UUID) ([]Card, error) {
	sqlString := `
		SELECT
			C.CATEGORY,
			C.TEXT
		FROM CARD AS C
		WHERE C.DECK_ID = ?
		ORDER BY C.CATEGORY ASC,
			C.TEXT ASC
	`
	rows, err := query(sqlString, deckId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Card, 0)
	for rows.Next() {
		var card Card
		if err := rows.Scan(&card.Category, &card.Text); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, card)
	}
	return result, nil
}

func GetCard(id uuid.UUID) (Card, error) {
	var card Card

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			DECK_ID,
			CATEGORY,
			TEXT,
			YOUTUBE,
			IMAGE
		FROM CARD
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return card, err
	}
	defer rows.Close()

	for rows.Next() {
		var imageBytes []byte
		if err := rows.Scan(
			&card.Id,
			&card.CreatedOnDate,
			&card.ChangedOnDate,
			&card.DeckId,
			&card.Category,
			&card.Text,
			&card.YouTube,
			&imageBytes); err != nil {
			log.Println(err)
			return card, errors.New("failed to scan row in query results")
		}

		card.Image.Valid = imageBytes != nil
		if card.Image.Valid {
			card.Image.String = base64.StdEncoding.EncodeToString(imageBytes)
		}
	}

	return card, nil
}

func CreateCard(deckId uuid.UUID, category string, text string, youtube string) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	sqlString := `
		INSERT INTO CARD(ID, DECK_ID, CATEGORY, TEXT, YOUTUBE)
		VALUES (?, ?, ?, ?, ?)
	`
	if len(youtube) == 0 {
		return id, execute(sqlString, id, deckId, category, text, nil)
	}
	return id, execute(sqlString, id, deckId, category, text, youtube)
}

func GetCardId(deckId uuid.UUID, text string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM CARD
		WHERE DECK_ID = ?
			AND TEXT = ?
	`
	rows, err := query(sqlString, deckId, text)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Println(err)
			return id, errors.New("failed to scan row in query results")
		}
	}

	return id, nil
}

func GetResponseCardTextStart(responseId uuid.UUID) (string, error) {
	var text string

	sqlString := `
		SELECT
			C.TEXT
		FROM RESPONSE AS R
			INNER JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
			INNER JOIN CARD AS C ON C.ID = RC.CARD_ID
		WHERE R.ID = ?
		ORDER BY RC.CREATED_ON_DATE
		LIMIT 1
	`
	rows, err := query(sqlString, responseId)
	if err != nil {
		return text, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&text); err != nil {
			log.Println(err)
			return text, errors.New("failed to scan row in query results")
		}
	}

	if len(text) > 100 {
		text = text[:100] + "..."
	}

	return text, nil
}

func UpdateCard(id uuid.UUID, category string, text string, youtube string) error {
	sqlString := `
		UPDATE CARD
		SET CATEGORY = ?,
			TEXT = ?,
			YOUTUBE = ?
		WHERE ID = ?
	`
	if len(youtube) == 0 {
		return execute(sqlString, category, text, nil, id)
	}
	return execute(sqlString, category, text, youtube, id)
}

func SetCardImage(id uuid.UUID, imageBytes []byte) error {
	sqlString := `
		UPDATE CARD
		SET IMAGE = ?
		WHERE ID = ?
	`
	return execute(sqlString, imageBytes, id)
}

func DeleteCard(id uuid.UUID) error {
	sqlString := `
		DELETE
		FROM CARD
		WHERE ID = ?
	`
	return execute(sqlString, id)
}

// AuditDeckCardsAsDeleted snapshots all of a deck's cards into AUDIT_CARD as
// 'DELETE'. Called from the OnDeckDeleting hook because MariaDB FK cascade does
// not fire the CARD delete trigger when the framework deletes the deck.
func AuditDeckCardsAsDeleted(deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO AUDIT_CARD(AUDIT_TYPE, CARD_ID, DECK_ID, CATEGORY, TEXT, YOUTUBE, IMAGE)
		SELECT 'DELETE', ID, DECK_ID, CATEGORY, TEXT, YOUTUBE, IMAGE
		FROM CARD
		WHERE DECK_ID = ?
	`
	return execute(sqlString, deckId)
}

func RecoverCard(id uuid.UUID) error {
	sqlString := `
		INSERT INTO CARD(DECK_ID, CATEGORY, TEXT, YOUTUBE, IMAGE)
		SELECT
			DECK_ID,
			CATEGORY,
			TEXT,
			YOUTUBE,
			IMAGE
		FROM REVIEW_CARD
		WHERE ID = ?
	`
	err := execute(sqlString, id)
	if err != nil {
		return err
	}

	return PermanentlyDeleteCard(id)
}

func PermanentlyDeleteCard(id uuid.UUID) error {
	sqlString := `
		DELETE
		FROM REVIEW_CARD
		WHERE ID = ?
	`
	return execute(sqlString, id)
}
