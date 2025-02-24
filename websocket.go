package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var lastClientId atomic.Uint32

type Client struct {
	id   uint
	name string
	ws   *websocket.Conn

	incoming chan WSMessage
	outgoing chan WSMessage
	closed   chan struct{}
}

func NewClient(name string, conn *websocket.Conn) *Client {
	id := uint(lastClientId.Add(1))
	return &Client{
		id:       id,
		name:     name,
		ws:       conn,
		incoming: make(chan WSMessage, 10),
		outgoing: make(chan WSMessage, 10),
		closed:   make(chan struct{}, 10),
	}
}

func (c *Client) Close() error {
	return c.ws.Close()
}

func (c *Client) Run() {
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) readLoop() {
	defer close(c.incoming)
	logger := slog.With("client", c.id)
	for {
		msgType, bytes, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logger.Info("connection closed while reading", "error", err)
				c.closed <- struct{}{}
				break
			}
			logger.Error("error reading message", "error", err)
			continue
		}

		switch msgType {
		case websocket.CloseMessage:
			logger.Info("connection closed by client")
			c.closed <- struct{}{}
			continue
		case websocket.PingMessage:
			err := c.ws.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				logger.Error("error sending pong message", "error", err)
				continue
			}
			continue
		case websocket.PongMessage:
			err := c.ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				logger.Error("error sending pong message", "error", err)
				continue
			}
			continue
		case websocket.BinaryMessage:
		case websocket.TextMessage:
			msg, err := UnmarshalWSMessage(bytes)
			if err != nil {
				logger.Error("error unmarshalling message", "error", err)
				continue
			}
			c.incoming <- msg
		}
	}
}

func (c *Client) writeLoop() {
	defer close(c.outgoing)
	logger := slog.With("client", c.id)
	for msg := range c.outgoing {
		if err := c.ws.WriteJSON(msg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logger.Info("connection closed while writing", "error", err)
				c.closed <- struct{}{}
				break
			}

			logger.Error("error writing message", "error", err, "message", msg)
		}
	}
}

type WSMessageType string

const (
	WSMessageTypeError WSMessageType = "error"
	WSMessageTypeStart WSMessageType = "start"
	WSMessageTypeJoin  WSMessageType = "join"
)

type WSMessage interface {
	GetType() WSMessageType
}

type WSMessageBase struct {
	Type WSMessageType `json:"type"`
}

func (m WSMessageBase) GetType() WSMessageType {
	return m.Type
}

type WSMessageError struct {
	WSMessageBase
	Error string `json:"error"`
}

func NewWSMessageError(err error) WSMessageError {
	return WSMessageError{
		WSMessageBase: WSMessageBase{Type: WSMessageTypeError},
		Error:         err.Error(),
	}
}

type WSMessageStart struct {
	WSMessageBase
}

func NewWSMessageStart() WSMessageStart {
	return WSMessageStart{
		WSMessageBase: WSMessageBase{Type: WSMessageTypeStart},
	}
}

type WSMessageJoin struct {
	WSMessageBase
	Name string `json:"name"`
}

func NewWSMessageJoin(name string) WSMessageJoin {
	return WSMessageJoin{
		WSMessageBase: WSMessageBase{Type: WSMessageTypeJoin},
		Name:          name,
	}
}

func UnmarshalWSMessage(data []byte) (WSMessage, error) {
	var msg WSMessageBase
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case WSMessageTypeError:
		var m WSMessageError
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return m, nil
	case WSMessageTypeStart:
		var m WSMessageStart
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return m, nil
	case WSMessageTypeJoin:
		var m WSMessageJoin
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unknown message type %q", msg.Type)
	}
}
