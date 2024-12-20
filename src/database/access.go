package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func UserHasLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_HAS_LOBBY_ACCESS (?, ?)"
	rows, err := query(sqlString, userId, lobbyId)
	if err != nil {
		return false, err
	}

	hasAccess := false
	for rows.Next() {
		if err := rows.Scan(&hasAccess); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return hasAccess, nil
}

func AddUserLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_LOBBY (USER_ID, LOBBY_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, lobbyId)
}

func UserHasDeckAccess(userId uuid.UUID, deckId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_HAS_DECK_ACCESS (?, ?)"
	rows, err := query(sqlString, userId, deckId)
	if err != nil {
		return false, err
	}

	hasAccess := false
	for rows.Next() {
		if err := rows.Scan(&hasAccess); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return hasAccess, nil
}

func AddUserDeckAccess(userId uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_DECK (USER_ID, DECK_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, deckId)
}
