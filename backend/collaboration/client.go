package collaboration

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Client -> A single websocket connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan Message
	userID   string
	userName string
}

// ReadPump -> listens for incoming messages from Websocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	// Set the limit and deadline
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)

		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}

		msg.SenderID = c.userID
		msg.Timestamp = time.Now()
		msg.SenderName = c.userName

		// If this is a chant msg then save the message
		if msg.Type == MessageTypeChat {
			go SaveChatMessage(msg, msg.RoomID)
		}
		c.hub.Broadcast <- msg

	}
}

// WritePump -> send outgoing messages from Websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// hub closes the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Error writing JSON: %v", err)
				return
			}

		case <-ticker.C:
			// Send a ping to maintain connection liveliness
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
