package models

type Equipment struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Name              string `json:"name" gorm:"not null"`
	AvailableQuantity int    `json:"available_quantity" gorm:"not null"`
}
