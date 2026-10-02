package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	mux := http.NewServeMux()

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using existing environment variables")
	}

	initDB()
	defer db.Close()

	initJWT()

	mux.HandleFunc("POST /register", registerUser)

	mux.HandleFunc("POST /login", loginUser)

	mux.HandleFunc("GET /me", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(userIDKey)
		json.NewEncoder(w).Encode(map[string]any{"user_id": id})
	}))

	mux.HandleFunc("POST /tasks", authMiddleware(createTask))
	mux.HandleFunc("GET /tasks", authMiddleware(getTasks))

	mux.HandleFunc("GET /tasks/{id}", authMiddleware(getTaskByID))

	fmt.Println("Server listening on :8080")

	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan // blocks here until Ctrl+C

	slog.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err.Error())
	} else {
		slog.Info("server shut down cleanly")
	}

}
