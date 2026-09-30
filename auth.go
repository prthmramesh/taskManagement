package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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

func initJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtSecret = []byte(secret)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			slog.Warn("Missing or invalid Authorization header", "header", "auth_failed")
			return
		}

		tokenStr := strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		}, jwt.WithExpirationRequired())

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			slog.Warn("auth failed", "reason", err.Error())
			return
		}
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			slog.Warn("auth failed", "reason", "token not valid")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			slog.Warn("Invalid token claims", "error", "user_id not found")
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			slog.Warn("Invalid token claims", "error", "user_id not found")
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, int(userID))

		slog.Info("request authenticated", "user_id", int(userID), "method", r.Method, "path", r.URL.Path)

		next(w, r.WithContext(ctx))

	}
}
