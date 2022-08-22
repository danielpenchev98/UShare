package models

import "time"

//User is a model representing a record in the table of Users
type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string `gorm:"type:varchar(20);not null"`
	Password  string `gorm:"type:varchar(256);not null"`
}
