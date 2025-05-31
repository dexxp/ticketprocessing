package models

import "time"

type RequestStatusLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RequestID uint      `json:"request_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"not null"`
	Timestamp time.Time `json:"timestamp" gorm:"not null"`
	Comment   string    `json:"comment"`
}
