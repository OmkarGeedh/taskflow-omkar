package models

import "time"

type Task struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID   string     `gorm:"not null" json:"project_id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `gorm:"default:null" json:"description"`
	Status      string     `gorm:"not null" json:"status"`
	Priority    string     `gorm:"not null" json:"priority"`
	AssigneeID  *string    `gorm:"default:null" json:"assignee_id"`
	DueDate     *time.Time `gorm:"default:null" json:"due_date"`
	CreatorID   string     `gorm:"not null" json:"creator_id"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
}