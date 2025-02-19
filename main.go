package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// App
	app := NewApp()

	// Logging
	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})
	slog.SetDefault(slog.New(slogHandler))

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == "GET" {
			app.Index(w, r)
			return
		}

		http.NotFound(w, r)
	})
	http.HandleFunc("POST /create", app.CreateRoom)
	http.HandleFunc("GET /room/join", app.JoinRoom) // this is via a link
	http.HandleFunc("GET /ws/{roomId}", app.Websocket)

	// Server
	slog.Info("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
