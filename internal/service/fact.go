package service

import (
	"time"
)

type Fact struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"-"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
}
