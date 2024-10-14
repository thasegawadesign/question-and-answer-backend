package models

type Item struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Question  string `json:"question" validate:"required"`
	Answer    string `json:"answer" validate:"required"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail;references:Email;constraint:OnDelete:CASCADE" validate:"omitempty"`
	UserEmail string `json:"user_email" gorm:"not null" validate:"required,email"`
}
