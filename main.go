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
	http.DefaultServeMux = routes(app)

	// Server
	slog.Info("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func routes(app *App) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == "GET" {
			app.Index(w, r)
			return
		}

		http.NotFound(w, r)
	})
	mux.HandleFunc("POST /create", app.CreateRoom)
	mux.HandleFunc("GET /room/join", app.JoinRoom) // this is via a link
	mux.HandleFunc("GET /ws/{roomId}", app.Websocket)
	return mux
}