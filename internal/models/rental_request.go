package models

import (
	"time"
)

type RentalRequest struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	EquipmentID uint      `json:"equipment_id" gorm:"not null"`
	FromDate    time.Time `json:"from_date" gorm:"not null"`
	ToDate      time.Time `json:"to_date" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
