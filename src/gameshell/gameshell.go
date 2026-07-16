package gameshell

import "github.com/google/uuid"

// Game is implemented by a game to receive lifecycle hooks from the platform
// shell when framework-owned rooms and participants are created. The framework
// must never import a game; a game registers its implementation at startup and
// the shell invokes it through this interface (dependency inversion).
type Game interface {
	// OnRoomCreated runs after a base LOBBY row is inserted, letting the game
	// bootstrap its own per-room state.
	OnRoomCreated(lobbyId uuid.UUID) error
	// OnPlayerJoined runs after a new base PLAYER row is inserted, letting the
	// game bootstrap its own per-player state.
	OnPlayerJoined(playerId uuid.UUID) error
	// OnPlayerActive runs after an existing inactive player becomes active
	// again (rejoins the room).
	OnPlayerActive(playerId uuid.UUID) error
	// OnPlayerInactive runs after an active player becomes inactive (leaves
	// the room).
	OnPlayerInactive(playerId uuid.UUID) error
	// OnRoomEmpty runs when the last client disconnects from a room, letting
	// the game clean up its own per-room state before the shell deletes the
	// base LOBBY row.
	OnRoomEmpty(lobbyId uuid.UUID) error
}

var registeredGame Game

// Register sets the game whose hooks the shell will invoke. Called once at
// startup from main.
func Register(g Game) {
	registeredGame = g
}

// Registered returns the registered game, or nil if none has been registered.
func Registered() Game {
	return registeredGame
}
