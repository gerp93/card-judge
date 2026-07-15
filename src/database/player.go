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

	Name              string
	LobbyId           uuid.UUID
	UserId            uuid.UUID
	JoinOrder         int
	IsActive          bool
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
			P.IS_ACTIVE,
			CJPS.WINNING_STREAK,
			CJPS.LOSING_STREAK,
			CJPS.CREDITS_SPENT,
			CJPS.BET_ON_WIN,
			CJPS.EXTRA_RESPONSES,
			CJPS.HAND_SIZE_ADVANTAGE,
			CJPS.DISCARD_ADVANTAGE,
			CJPS.HANDICAP_ADVANTAGE,
			CJPS.SPY_ADVANTAGE
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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
			&player.WinningStreak,
			&player.LosingStreak,
			&player.CreditsSpent,
			&player.BetOnWin,
			&player.ExtraResponses,
			&player.HandSizeAdvantage,
			&player.DiscardAdvantage,
			&player.HandicapAdvantage,
			&player.SpyAdvantage,
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
			P.IS_ACTIVE,
			CJPS.WINNING_STREAK,
			CJPS.LOSING_STREAK,
			CJPS.CREDITS_SPENT,
			CJPS.BET_ON_WIN,
			CJPS.EXTRA_RESPONSES,
			CJPS.HAND_SIZE_ADVANTAGE,
			CJPS.DISCARD_ADVANTAGE,
			CJPS.HANDICAP_ADVANTAGE,
			CJPS.SPY_ADVANTAGE
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			INNER JOIN CJ_PLAYER_STATE AS CJPS ON CJPS.PLAYER_ID = P.ID
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
			&player.WinningStreak,
			&player.LosingStreak,
			&player.CreditsSpent,
			&player.BetOnWin,
			&player.ExtraResponses,
			&player.HandSizeAdvantage,
			&player.DiscardAdvantage,
			&player.HandicapAdvantage,
			&player.SpyAdvantage,
		); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}
