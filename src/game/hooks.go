package game

import (
	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/database"
)

// CardJudge implements gameshell.Game — the card-judge game's lifecycle hooks.
// Each hook is a thin wrapper over a game stored procedure, matching the DB
// layer's CALL SP_... convention.
type CardJudge struct{}

func (CardJudge) OnRoomCreated(lobbyId uuid.UUID) error {
	return database.InitLobbyGame(lobbyId)
}

func (CardJudge) OnPlayerJoined(playerId uuid.UUID) error {
	return database.InitPlayerGame(playerId)
}

func (CardJudge) OnRoomEmpty(lobbyId uuid.UUID) error {
	return database.CleanupLobbyGame(lobbyId)
}
