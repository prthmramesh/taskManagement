package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

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

func loginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	dbUser, err := getUserByUsername(user.Username)
	if err != nil {

		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			slog.Warn("login failed", "reason", "user not found", "username", user.Username)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusUnauthorized)
			slog.Error("login failed", "reason", "internal server error", "username", user.Username, "error", err.Error())
		}

		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		slog.Warn("login failed", "reason", "invalid password", "username", user.Username)
		return
	}

	claims := jwt.MapClaims{
		"user_id": dbUser.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		slog.Error("login failed", "reason", "failed to generate token", "username", user.Username, "error", err.Error())
		return
	}

	slog.Info("User logged in", "username", user.Username, "user_id", user.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"username": dbUser.Username,
		"email":    dbUser.Email,
		"token":    signed,
	})

}
