package models

import "time"

//FileInfo is a model representing the most important info for a file
type FileInfo struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Name      string `gorm:"type:varchar(256);not null"`
	OwnerID   uint   `gorm:"type:Integer;not null"`
	GroupID   uint   `gorm:"type:Integer;not null"`
}
