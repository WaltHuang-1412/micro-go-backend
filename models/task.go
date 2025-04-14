package models

type Task struct {
	ID          int    `json:"id"`
	SectionID   int    `json:"section_id"`
	Content     string `json:"content"`
	IsCompleted bool   `json:"is_completed"`
	SortOrder   int    `json:"sort_order"`
}
