package collaboration

import (
	"log"
	"sync"
)

// Hub maintains a set of active clients and broadcast messages
type Hub struct {
	// RoomID
	RoomId string
	//Registered clients
	Clients map[*Client]bool
	// Inbound message from client
	Broadcast chan Message // -> channel that will pass on Message struct
	//Register requests from clients
	Register chan *Client
	//Unregister requests from clients
	Unregister chan *Client
	//Lock to guard the Clinets mao
	Mutex sync.Mutex
}

// Create a new Hub instance
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Starts the Hub main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.Mutex.Unlock()
			log.Printf("Client Registered: %s", client.userID)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
				log.Printf("Client Unregistered: %s", client.userID)
			}
			h.Mutex.Unlock()

		case message := <-h.Broadcast:
			h.Mutex.Lock()
			for client := range h.Clients {
				select {
				case client.send <- message:
				default:
					// If the clients send buffer is full remove the client
					close(client.send)
					delete(h.Clients, client)
				}
			}
			h.Mutex.Unlock()
		}
	}
}
