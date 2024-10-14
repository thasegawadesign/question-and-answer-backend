package models

type Item struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail;references:Email;constraint:OnDelete:CASCADE"`
	UserEmail string `json:"user_email" gorm:"not null"`
}
