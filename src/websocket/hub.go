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

	// Messages targeted at a single user (by user id).
	direct chan directMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// directMessage is a payload destined for every client belonging to a single
// user, delivered through the hub goroutine so the clients map is never
// touched concurrently.
type directMessage struct {
	userId  uuid.UUID
	payload []byte
}

func newHub(lobbyId uuid.UUID) *Hub {
	return &Hub{
		lobbyId:    lobbyId,
		broadcast:  make(chan []byte),
		direct:     make(chan directMessage),
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
		case dm := <-h.direct:
			h.directMessage(dm)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
	h.broadcastMessage([]byte("<blue>Player Joined</>: <green>" + client.user.Name + "</>"))
	h.broadcastMessage([]byte("refresh"))
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		_ = database.SetPlayerInactive(h.lobbyId, client.user.Id)
	}
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

func (h *Hub) directMessage(dm directMessage) {
	for client := range h.clients {
		if client.user.Id != dm.userId {
			continue
		}
		select {
		case client.send <- dm.payload:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func LobbyBroadcast(lobbyId uuid.UUID, message string) {
	if hub, ok := lobbyHubs[lobbyId]; ok {
		hub.broadcast <- []byte(message)
	}
}

func PlayerBroadcast(playerId uuid.UUID, message string) {
	player, err := database.GetPlayer(playerId)
	if err != nil {
		return
	}

	if hub, ok := lobbyHubs[player.LobbyId]; ok {
		hub.direct <- directMessage{
			userId:  player.UserId,
			payload: []byte(message),
		}
	}
}
