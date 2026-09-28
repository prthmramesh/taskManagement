package main

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	ID       int    `json:"id"`
}

var jwtSecret []byte

type contextKey string

const userIDKey contextKey = "user_id"
