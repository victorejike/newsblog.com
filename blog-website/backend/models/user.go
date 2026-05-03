package models

import "time"

type User struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	IsAdmin       bool      `json:"is_admin"`
	Bio           string    `json:"bio,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	WebsiteURL    string    `json:"website_url,omitempty"`
	TwitterHandle string    `json:"twitter_handle,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Bio           string `json:"bio"`
	AvatarURL     string `json:"avatar_url"`
	WebsiteURL    string `json:"website_url"`
	TwitterHandle string `json:"twitter_handle"`
}
