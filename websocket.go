package main

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type Client struct {
	ws *websocket.Conn

	read chan WSMessage
	write chan WSMessage
	closed chan struct{} 
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{ws: conn}
}

func (c *Client) Close() error {
	return c.ws.Close()
}

func (c *Client) Run() {
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) readLoop() {
	defer close(c.read)
	for {
		var msg WSMessage
		err := c.ws.ReadJSON(&msg)
		if websocket.IsCloseError(err) {
			slog.Info("connection closed while reading", "error", err)
			c.closed <- struct{}{}
			return
		} else if err != nil {
			slog.Error("error reading message", "error", err)
		}
		c.read <- msg
	}
}

func (c *Client) writeLoop() {
	defer close(c.write)
	for msg := range c.write {
		err := c.ws.WriteJSON(msg)
		if websocket.IsCloseError(err) {
			slog.Info("connection closed while writing", "error", err)
			return
		} else if err != nil {
			slog.Error("error writing message", "error", err, "message", msg)
		}
	}
}

type WSMessageType string

const (
	WSMessageTypeError WSMessageType = "error"
	WSMessageTypeStart WSMessageType = "start"
)

type WSMessage interface {
	Type() WSMessageType
}

type WSMessageBase struct {
	Type_ WSMessageType `json:"type"`
}

func (m WSMessageBase) Type() WSMessageType {
	return m.Type_
}

type WSMessageError struct {
	WSMessageBase
	Error string `json:"error"`
}

func NewWSMessageError(err error) WSMessageError {
	return WSMessageError{
		WSMessageBase: WSMessageBase{Type_: WSMessageTypeError},
		Error: err.Error(),
	}
}

type WSMessageStart struct {
	WSMessageBase
}

func NewWSMessageStart() WSMessageStart {
	return WSMessageStart{
		WSMessageBase: WSMessageBase{Type_: WSMessageTypeStart},
	}
}
