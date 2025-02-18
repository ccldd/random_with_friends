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
	http.HandleFunc("GET /", app.Index)
	http.HandleFunc("POST /create", app.CreateRoom)
	http.HandleFunc("GET /room/{roomId}", app.GetRoom) // this is via a link
	http.HandleFunc("POST /room", app.PostRoom) // this is via the form
	http.HandleFunc("GET /ws/{roomId}", app.Websocket)

	// Server
	slog.Info("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
