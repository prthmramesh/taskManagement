package main

import (
	"log"
	"log/slog"
	"os"

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
