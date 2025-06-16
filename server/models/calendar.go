package models

import "time"

type Calendar struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	FamilyID uint `gorm:"not null" json:"family_id"`
	Title     string    `gorm:"size:100" json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
