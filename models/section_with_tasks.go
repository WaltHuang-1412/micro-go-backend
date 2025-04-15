package models

type SectionWithTasks struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Tasks     []Task `json:"tasks"`
}
