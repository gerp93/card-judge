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

	Name      string
	LobbyId   uuid.UUID
	UserId    uuid.UUID
	JoinOrder int
	IsActive  bool
}

type PlayerGameState struct {
	PlayerId uuid.UUID

	WinningStreak     int
	LosingStreak      int
	CreditsSpent      int
	BetOnWin          int
	ExtraResponses    int
	HandSizeAdvantage int
	DiscardAdvantage  bool
	HandicapAdvantage bool
	SpyAdvantage      bool
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
			P.IS_ACTIVE
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return player, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.JoinOrder,
			&player.IsActive,
		); err != nil {
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
			P.IS_ACTIVE
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.LOBBY_ID = ?
			AND P.USER_ID = ?
	`
	rows, err := query(sqlString, lobbyId, userId)
	if err != nil {
		return player, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.JoinOrder,
			&player.IsActive,
		); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}

func GetPlayerGameState(playerId uuid.UUID) (PlayerGameState, error) {
	var state PlayerGameState

	sqlString := `
		SELECT
			PLAYER_ID,
			WINNING_STREAK,
			LOSING_STREAK,
			CREDITS_SPENT,
			BET_ON_WIN,
			EXTRA_RESPONSES,
			HAND_SIZE_ADVANTAGE,
			DISCARD_ADVANTAGE,
			HANDICAP_ADVANTAGE,
			SPY_ADVANTAGE
		FROM CJ_PLAYER_STATE
		WHERE PLAYER_ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return state, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&state.PlayerId,
			&state.WinningStreak,
			&state.LosingStreak,
			&state.CreditsSpent,
			&state.BetOnWin,
			&state.ExtraResponses,
			&state.HandSizeAdvantage,
			&state.DiscardAdvantage,
			&state.HandicapAdvantage,
			&state.SpyAdvantage,
		); err != nil {
			log.Println(err)
			return state, errors.New("failed to scan row in query results")
		}
	}

	return state, nil
}
