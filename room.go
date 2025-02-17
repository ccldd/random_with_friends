package main

import (
	"github.com/gorilla/websocket"
)

type RoomId = string

type Room struct {
	ID   string
	host *websocket.Conn
	members []*websocket.Conn
}

func NewRoom(host *websocket.Conn) *Room {
	id := randomString(4)
	return &Room{
		ID: id,
		host: host,
	}
}
