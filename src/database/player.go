package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

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
