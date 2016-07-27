package switcher

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krivokhatko/wschat/pkg/logger"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// User is an middleman between the websocket connection and the switch.
type User struct {
	MsgSwitcher *Switcher

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan *ChatMessage
}

// ReadPump pumps messages from the websocket connection to the switch.
func (c *User) ReadPump() {
	defer func() {
		c.MsgSwitcher.logout <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Log(logger.LevelError, "user", "error: %v", err)
			}
			break
		}
		chat_msg := &ChatMessage{}
		json.Unmarshal([]byte(message), chat_msg)

		logger.Log(logger.LevelNotice, "user", "message read `%s' `%s'\n", string(chat_msg.NickName), string(chat_msg.Text))

		c.MsgSwitcher.spread <- chat_msg
	}
}

// Write writes a message with the given message type and payload.
func (c *User) Write(mt int, payload []byte) error {
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteMessage(mt, payload)
}

// WritePump pumps messages from the switch to the websocket connection.
func (c *User) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The switch closed the channel.
				c.Write(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			n := len(c.Send)
			cm := &ChatMessages{make([]*ChatMessage, 1+n)}
			cm.Messages[0] = message
			for i := 0; i < n; i++ {
				message, ok := <-c.Send
				if !ok {
					c.Write(websocket.CloseMessage, []byte{})
					return
				}
				cm.Messages[i+1] = message
			}

			json_bytes, err := json.Marshal(cm)
			if err != nil {
				return
			}
			w.Write(json_bytes)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
