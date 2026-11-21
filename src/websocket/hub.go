package websocket

import (
	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/database"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Lobby ID
	lobbyId uuid.UUID

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub(lobbyId uuid.UUID) *Hub {
	return &Hub{
		lobbyId:    lobbyId,
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
			if len(h.clients) == 0 {
				_ = database.DeleteLobby(h.lobbyId)
				delete(lobbyHubs, h.lobbyId)
				return
			}
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
	h.broadcastMessage([]byte("<blue>Player Joined</>: <green>" + client.user.Name + "</>"))
	h.broadcastMessage([]byte("refresh"))
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		return
	}
	delete(h.clients, client)
	close(client.send)

	// A single user can briefly hold more than one connection at once: when a
	// client drops and reconnects, the new socket registers while the old,
	// half-open one lingers until its read deadline (pongWait) expires. Player
	// presence is keyed by user, not by connection, so only treat the user as
	// having left once their LAST connection is gone. Otherwise the stale
	// socket closing would mark a still-connected player inactive and broadcast
	// a phantom "Player Left" to their live session.
	for c := range h.clients {
		if c.user.Id == client.user.Id {
			return
		}
	}

	_ = database.SetPlayerInactive(h.lobbyId, client.user.Id)
	h.broadcastMessage([]byte("<red>Player Left</>: <green>" + client.user.Name + "</>"))
	h.broadcastMessage([]byte("refresh"))
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func LobbyBroadcast(lobbyId uuid.UUID, message string) {
	if hub, ok := lobbyHubs[lobbyId]; ok {
		hub.broadcastMessage([]byte(message))
	}
}

func PlayerBroadcast(playerId uuid.UUID, message string) {
	player, err := database.GetPlayer(playerId)
	if err != nil {
		return
	}

	if hub, ok := lobbyHubs[player.LobbyId]; ok {
		for client := range hub.clients {
			if client.user.Id == player.UserId {
				client.send <- []byte(message)
			}
		}
	}
}
