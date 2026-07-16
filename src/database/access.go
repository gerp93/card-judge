package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func UserHasDeckAccess(userId uuid.UUID, deckId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_HAS_DECK_ACCESS (?, ?)"
	rows, err := query(sqlString, userId, deckId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

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
		INSERT INTO USER_ACCESS_DECK(USER_ID, DECK_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, deckId)
}
