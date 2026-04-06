package models

import "time"

type User struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Password  string `json:"-"` // ❌ don't expose
	CreatedAt time.Time `json:"created_at"`
}