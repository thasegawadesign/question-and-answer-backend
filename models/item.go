package models

type Item struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	User     User   `json:"user" gorm:"foreignKey:Email;references:Email;constraint:OnDelete:CASCADE"`
	Email    string `json:"email" gorm:"not null"`
}
