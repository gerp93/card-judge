package database

import (
	"errors"
	"log"
	"time"

	"github.com/gerp93/gameshell-framework/auth"
	"github.com/google/uuid"
)

type Deck struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name             string
	PasswordHash     string
	IsPublicReadOnly bool
}

type DeckDetails struct {
	Deck
	CardCount int
}

func SearchDecks(name string, page int) ([]DeckDetails, error) {
	name = "%" + name + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			D.ID,
			D.NAME,
			COUNT(C.ID) AS CARD_COUNT,
			D.IS_PUBLIC_READONLY
		FROM DECK AS D
			LEFT JOIN CARD AS C ON C.DECK_ID = D.ID
		WHERE D.IS_LOBBY_WILD_DECK = FALSE
			AND D.NAME LIKE ?
		GROUP BY D.ID
		ORDER BY D.NAME
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, name, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]DeckDetails, 0)
	for rows.Next() {
		var deckDetails DeckDetails
		if err := rows.Scan(
			&deckDetails.Id,
			&deckDetails.Name,
			&deckDetails.CardCount,
			&deckDetails.IsPublicReadOnly,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, deckDetails)
	}
	return result, nil
}

func CountDecks(name string) (int, error) {
	name = "%" + name + "%"

	sqlString := `
		SELECT
			COUNT(*)
		FROM DECK AS D
		WHERE D.IS_LOBBY_WILD_DECK = FALSE
			AND D.NAME LIKE ?
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

func GetReadableDecks(userId uuid.UUID) ([]Deck, error) {
	sqlString := "CALL SP_GET_READABLE_DECKS (?)"
	rows, err := query(sqlString, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Deck, 0)
	for rows.Next() {
		var deck Deck
		if err := rows.Scan(
			&deck.Id,
			&deck.Name); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, deck)
	}
	return result, nil
}

func GetDeck(id uuid.UUID) (Deck, error) {
	var deck Deck

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			IS_PUBLIC_READONLY
		FROM DECK
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return deck, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&deck.Id,
			&deck.CreatedOnDate,
			&deck.ChangedOnDate,
			&deck.Name,
			&deck.PasswordHash,
			&deck.IsPublicReadOnly); err != nil {
			log.Println(err)
			return deck, errors.New("failed to scan row in query results")
		}
	}

	return deck, nil
}

func GetDeckPasswordHash(id uuid.UUID) (string, error) {
	var passwordHash string

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM DECK
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return passwordHash, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func CreateDeck(name string, password string, isPublicReadOnly bool) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to hash password")
	}

	sqlString := `
		INSERT INTO DECK(ID, NAME, PASSWORD_HASH, IS_PUBLIC_READONLY)
		VALUES (?, ?, ?, ?)
	`
	return id, execute(sqlString, id, name, passwordHash, isPublicReadOnly)
}

func GetDeckId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM DECK
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
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

func SetDeckName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE DECK
		SET NAME = ?
		WHERE ID = ?
	`
	return execute(sqlString, name, id)
}

func SetIsPublicReadOnly(id uuid.UUID, isPublicReadOnly bool) error {
	sqlString := `
		UPDATE DECK
		SET IS_PUBLIC_READONLY = ?
		WHERE ID = ?
	`
	return execute(sqlString, isPublicReadOnly, id)
}

func SetDeckPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE DECK
		SET PASSWORD_HASH = ?
		WHERE ID = ?
	`
	return execute(sqlString, passwordHash, id)
}

func DeleteDeck(id uuid.UUID) error {
	sqlString := `
		DELETE
		FROM DECK
		WHERE ID = ?
	`
	return execute(sqlString, id)
}
