package models

import "time"

type Project struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"default:null" json:"description"`
	OwnerID     string    `gorm:"not null" json:"owner_id"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
}
