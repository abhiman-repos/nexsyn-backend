package models

import "time"

type GoogleAuth struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Provider   string `json:"provider"`    // "google"
	ProviderID string `json:"provider_id"` // Google user ID

	Email string `json:"email"`

	AccessToken  string `json:"-"` // ❌ don't expose
	RefreshToken string `json:"-"` // ❌ don't expose

	Expiry int64 `json:"expiry"`

	CreatedAt time.Time
	UpdatedAt time.Time
}