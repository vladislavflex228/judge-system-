package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	HashPassword string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewUser(username, email, hash_password string, created_at time.Time) *User {
	return &User{Username: username, Email: email, HashPassword: hash_password, CreatedAt: created_at}
}
