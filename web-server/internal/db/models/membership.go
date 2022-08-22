package models

import "time"

//Membership is a model representing a record in the table of Memberships
type Membership struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	GroupID   uint `gorm:"type:bigint;not null"`
	UserID    uint `gorm:"type:bigint;not null"`
}
