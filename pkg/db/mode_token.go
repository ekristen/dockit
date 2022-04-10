package db

import "time"

type Token struct {
	ID           int64 `gorm:"primaryKey;autoIncrement:false" json:"id"`
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	IssuedAt     *time.Time
	ExpiresAt    *time.Time
	UserID       int64
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
