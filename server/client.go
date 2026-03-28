package server

import (
	"log"
	"sync"

	"coa-server/game"

	"github.com/gorilla/websocket"
)

const sendBufferSize = 64

// Client represents a single WebSocket connection.
type Client struct {
	conn        *websocket.Conn
	send        chan []byte
	once        sync.Once
	PlayerIndex game.PlayerIndex
	PlayerID    string
}

func NewClient(conn *websocket.Conn, playerIndex game.PlayerIndex, playerID string) *Client {
	return &Client{
		conn:        conn,
		send:        make(chan []byte, sendBufferSize),
		PlayerIndex: playerIndex,
		PlayerID:    playerID,
	}
}

// Send queues a message for delivery to this client.
func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		log.Printf("client %s send buffer full, dropping message", c.PlayerID)
	}
}

// WritePump drains the send channel and writes to the WebSocket connection.
// Run this in its own goroutine.
func (c *Client) WritePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("write error for client %s: %v", c.PlayerID, err)
			return
		}
	}
}

// Close shuts down the send channel once.
func (c *Client) Close() {
	c.once.Do(func() { close(c.send) })
}
