package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type RoomId = string

type Room struct {
	ID   string
	host *websocket.Conn

	rw      sync.RWMutex
	members []*websocket.Conn
}

func NewRoom() *Room {
	id := randomString(4)
	return &Room{
		ID: id,
	}
}

func (r *Room) HasHost() bool {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.host != nil
}

func (r *Room) SetHost(conn *websocket.Conn) {
	r.rw.Lock()
	defer r.rw.Unlock()

	if r.host != nil {
		panic("host already set")
	}
	if len(r.members) > 0 {
		panic(fmt.Sprintf("room %s has no host but has members already", r.ID))
	}
	r.host = conn
}

func (r *Room) Join(conn *websocket.Conn) {
	r.rw.Lock()
	defer r.rw.Unlock()
	r.members = append(r.members, conn)
}
