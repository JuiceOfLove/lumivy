package models

import "time"

type FamilyInvitation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FamilyID  uint      `json:"family_id"`
	Email     string    `gorm:"not null" json:"email"`
	Token     string    `gorm:"unique;not null" json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
