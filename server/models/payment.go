package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	PaymentID string `gorm:"uniqueIndex;not null" json:"payment_id"`
	FamilyID  uint   `gorm:"not null" json:"family_id"`
	UserID    uint   `json:"user_id"`
	Amount string `json:"amount"`
	Status string `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
