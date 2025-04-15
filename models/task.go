package models

type Task struct {
	ID          int64  `json:"id"`
	SectionID   int64  `json:"section_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsCompleted bool   `json:"is_completed"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateTaskInput struct {
	SectionID   int64  `json:"section_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	IsCompleted bool   `json:"is_completed"`
}

type UpdateTaskInput struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsCompleted bool   `json:"is_completed"`
}
