package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

type App struct {
	tmpl     *template.Template
	upgrader websocket.Upgrader

	rw    sync.RWMutex
	rooms map[RoomId]*Room
}

// NewApp creates a new instance of App
func NewApp() *App {
	return &App{
		upgrader: websocket.Upgrader{},
		rooms:    make(map[RoomId]*Room),
		rw:       sync.RWMutex{},
	}
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	tmpl, err := a.tmpl.ParseFiles("templates/layout.html", filepath.Join("templates", name))
	if err != nil {
		slog.Error("error parsing template: %w", "error", err, "template", name)
	}
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("error rendering template: %w", "error", err)
	}
}

func (a *App) doesRoomExists(roomId RoomId) bool {
	a.rw.RLock()
	defer a.rw.RUnlock()
	_, ok := a.rooms[roomId]
	return ok
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	a.render(w, "index.html", nil)
}

func (a *App) JoinRoom(w http.ResponseWriter, r *http.Request) {
	// If roomId in path, then it must be from a link
	roomId := r.URL.Query().Get("roomId")
	if roomId == "" {
		http.Error(w, "missing room ID", http.StatusBadRequest)
		return
	}
	if !a.doesRoomExists(roomId) {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	data := make(map[string]any)
	data["RoomId"] = roomId
	a.render(w, "room.html", data)
}

func (a *App) CreateRoom(w http.ResponseWriter, r *http.Request) {
	a.rw.Lock()
	room := NewRoom()
	a.rooms[room.ID] = room
	a.rw.Unlock()
	slog.Info("room created", "roomId", room.ID)

	http.Redirect(w, r, fmt.Sprintf("/room/join?roomId=%s", room.ID), http.StatusSeeOther)
}

func (a *App) PostRoom(w http.ResponseWriter, r *http.Request) {
	roomId := r.Form.Get("roomId")
	if roomId == "" {
		http.Error(w, "missing room ID", http.StatusBadRequest)
		return
	}

	redirectUrl := fmt.Sprintf("/room/%s", roomId)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func (a *App) Websocket(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")
	if roomId == "" {
		http.Error(w, "missing room ID", http.StatusBadRequest)
		return
	}

	// Check if room exists
	var room *Room
	var ok bool
	{
		a.rw.RLock()
		defer a.rw.RUnlock()
		room, ok = a.rooms[roomId]
		if !ok || room == nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
	}

	// Upgrade to websocket
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading websocket", "error", err)
		return
	}

	client := NewClient(ws)
	client.Run()

	// If first in the room, then it is the host
	if !room.HasHost() {
		a.handleHostConnect(room, client)
		room.SetOnRoomClose(func() { a.handleHostDisconnect(room) })
	} else {
		a.handleMemberConnect(room, client)
	}
}

func (a *App) handleHostConnect(room *Room, client *Client) {
	room.SetHost(client)
	slog.Info("host connected", "roomId", room.ID)
	go room.Run()
}

func (a *App) handleHostDisconnect(room *Room) {
	slog.Info("host disconnected", "roomId", room.ID, "membersCount", len(room.members))

	// Remove the room
	a.rw.Lock()
	room.Close()
	delete(a.rooms, room.ID)
	a.rw.Unlock()
}

func (a *App) handleMemberConnect(room *Room, client *Client) {
	room.Join(client)
	slog.Info("member connected", "roomId", room.ID, "membersCount", len(room.members))
}
