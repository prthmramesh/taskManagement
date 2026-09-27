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

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	mux := http.NewServeMux()

	initDB()
	defer db.Close()

	mux.HandleFunc("POST /register", registerUser)

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

func registerUser(w http.ResponseWriter, r *http.Request) {

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" || user.Email == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	hashedPassword := passwordHash(user.Password)
	if hashedPassword == nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)

	err = insertUser(user.Username, user.Password, user.Email)
	if err != nil {

		if pq, ok := err.(*pq.Error); ok && pq.Code == "23505" {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			slog.Warn("insert user failed", "reason", "username or email already exists", "user_name", user.Username, "email", user.Email)
			return
		}

		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		slog.Error("insert user failed", "user_name", user.Username, "error", err.Error())
		return
	}

	slog.Info("User registered", "username", user.Username)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"username": user.Username,
		"email":    user.Email,
	})
}

func passwordHash(password string) []byte {

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		slog.Error("Failed to hash password", "error", err.Error())
		return nil
	}
	return hashedPassword
}
