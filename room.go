package main

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type RoomId = string

type Room struct {
	ID   string
	host *Client
	isRunning atomic.Bool

	rw      sync.RWMutex
	members []*Client

	onRoomClose func()
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

func (r *Room) SetHost(client *Client) {
	r.rw.Lock()
	defer r.rw.Unlock()

	if r.host != nil {
		panic("host already set")
	}
	if len(r.members) > 0 {
		panic(fmt.Sprintf("room %s has no host but has members already", r.ID))
	}
	r.host = client
}

func (r *Room) Join(client *Client) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.host.outgoing <- NewWSMessageJoin(client.name)
	for _, member := range r.members {
		member.outgoing <- NewWSMessageJoin(client.name)
	}

	r.members = append(r.members, client)
}

func (r *Room) SetOnRoomClose(fn func()) {
	r.onRoomClose = fn
}

func (r *Room) Run() {
	if r.isRunning.Load() {
		panic(fmt.Sprintf("room %s is already running", r.ID))
	}

	r.waitHostStart()
}

func (r *Room) Close() {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.host.Close()

	for _, client := range r.members {
		client.ws.WriteJSON(NewWSMessageError(errors.New("host disconnected")))
		client.Close()
	}

	if r.onRoomClose != nil {
		r.onRoomClose()
	}
}

func (r *Room) waitHostStart() {
	logger := slog.With("roomId", r.ID)
	logger.Info("waiting for host to start random generation")
	for msg := range r.host.incoming {
		if _, ok := msg.(WSMessageStart); ok {
			logger.Info("host has started random generation")
			break
		}
	}
}