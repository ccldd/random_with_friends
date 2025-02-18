package main

import (
	"errors"
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

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	a.render(w, "index.html", nil)
}

func (a *App) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")
	if roomId == "" {
		http.Error(w, "missing room ID", http.StatusBadRequest)
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

	http.Redirect(w, r, fmt.Sprintf("/room/%s", room.ID), http.StatusSeeOther)
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

	// If first in the room, then it is the host
	if !room.HasHost() {
		a.handleHostConnect(room, ws)
	} else {
		a.handleMemberConnect(room, ws)
	}
}

func (a *App) handleHostConnect(room *Room, ws *websocket.Conn) {
	room.SetHost(ws)
	ws.SetCloseHandler(func(code int, text string) error {
		return a.handleHostDisconnect(code, text, room)
	})
}

func (a *App) handleHostDisconnect(code int, text string, room *Room) error {
	slog.Info("host disconnected", "roomId", room.ID, "membersCount", len(room.members))

	// Remove the room
	a.rw.Lock()
	delete(a.rooms, room.ID)
	a.rw.Unlock()

	// Let all members know that the room has been closed
	errs := make([]error, 0)
	room.rw.RLock()
	defer room.rw.RUnlock()
	for _, conn := range room.members {
		if err := conn.WriteJSON(NewWSMessageError(errors.New("host disconnected"))); err != nil {
			errs = append(errs, err)
		}
		if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, text)); err != nil {
			errs = append(errs, err)
		}
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *App) handleMemberConnect(room *Room, ws *websocket.Conn) {
	room.Join(ws)
	ws.SetCloseHandler(func(code int, text string) error {
		return a.handleMemberDisconnect(code, text, room, ws)
	})
}

func (a *App) handleMemberDisconnect(code int, text string, room *Room, ws *websocket.Conn) error {
	panic("unimplemented")
}
