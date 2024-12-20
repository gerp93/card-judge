package database

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id            uuid.UUID
	CreatedOnDate time.Time

	Name           string
	LobbyId        uuid.UUID
	UserId         uuid.UUID
	JoinOrder      int
	IsActive       bool
	WinningStreak  int
	LosingStreak   int
	CreditsSpent   int
	BetOnWin       int
	ExtraResponses int
}

func GetPlayer(playerId uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			P.ID,
			P.CREATED_ON_DATE,
			U.NAME,
			P.LOBBY_ID,
			P.USER_ID,
			P.JOIN_ORDER,
			P.IS_ACTIVE,
			P.WINNING_STREAK,
			P.LOSING_STREAK,
			P.CREDITS_SPENT,
			P.BET_ON_WIN,
			P.EXTRA_RESPONSES
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.JoinOrder,
			&player.IsActive,
			&player.WinningStreak,
			&player.LosingStreak,
			&player.CreditsSpent,
			&player.BetOnWin,
			&player.ExtraResponses); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}

func GetLobbyUserPlayer(lobbyId uuid.UUID, userId uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			P.ID,
			P.CREATED_ON_DATE,
			U.NAME,
			P.LOBBY_ID,
			P.USER_ID,
			P.JOIN_ORDER,
			P.IS_ACTIVE,
			P.WINNING_STREAK,
			P.LOSING_STREAK,
			P.CREDITS_SPENT,
			P.BET_ON_WIN,
			P.EXTRA_RESPONSES
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.LOBBY_ID = ?
			AND P.USER_ID = ?
	`
	rows, err := query(sqlString, lobbyId, userId)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.JoinOrder,
			&player.IsActive,
			&player.WinningStreak,
			&player.LosingStreak,
			&player.CreditsSpent,
			&player.BetOnWin,
			&player.ExtraResponses); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}
