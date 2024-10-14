package models

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null" validate:"required,email"`
	Provider string `json:"provider" validate:"required"`
}
