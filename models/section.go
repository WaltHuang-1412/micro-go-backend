package models

import "time"

type UpdateSectionInput struct {
	Title string `json:"title" binding:"required"`
}

type CreateSectionInput struct {
	Title string `json:"title" binding:"required"`
}

type Section struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
