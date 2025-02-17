package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

type App struct {
	tmpl *template.Template
	upgrader websocket.Upgrader

	rw sync.RWMutex
	rooms map[RoomId]*Room
}

var app App

func main() {
	// App
	app = App{
		upgrader: websocket.Upgrader{},
		rooms: make(map[RoomId]*Room),
	}

	// Logging
	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(slogHandler))

	// Templates
	app.tmpl = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

	// Routes
	http.HandleFunc("GET /", render("index.html", nil))
	http.HandleFunc("POST /create", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade to websocket
		ws, err := app.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("error upgrading websocket: %w", "error", err)
			return
		}

		// Create room
		app.rw.Lock()
		room := NewRoom(ws)
		app.rooms[room.ID] = room
		app.rw.Unlock()
		slog.Info("room created", "roomId", room.ID)	
	})
	http.HandleFunc("GET /join/{roomId}", func(w http.ResponseWriter, r *http.Request) {
		roomId := r.PathValue("roomId")
		if roomId == "" {
			http.Error(w, "missing room ID", http.StatusBadRequest)
			return
		}

		joinRoom(w, r, roomId)
	})
	http.HandleFunc("GET /room/{roomId}", func(w http.ResponseWriter, r *http.Request) {
		roomId := r.Form.Get("roomId")
		if roomId == "" {
			http.Error(w, "missing room ID", http.StatusBadRequest)
			return
		}

		joinRoom(w, r, roomId)
	})

	// Server
	slog.Info("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func joinRoom(w http.ResponseWriter, r *http.Request, roomId string) {
		// Check if room exists
		var room *Room
		var ok bool
		{
			app.rw.RLock()
			defer app.rw.RUnlock()
			room, ok = app.rooms[roomId];
			if !ok || room == nil {
				http.Error(w, "room not found", http.StatusNotFound)
				return
			}
		}

		// Upgrade to websocket
		ws, err := app.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("error upgrading websocket: %w", "error", err)
			return
		}

		// Join room
		app.rw.Lock()
		defer app.rw.Unlock()
		room.members = append(room.members, ws)
}

func render(name string, data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := app.tmpl.ExecuteTemplate(w, name, data); err != nil {
			slog.Error("error rendering template: %w", "error", err)
		}
	}
}
