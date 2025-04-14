package models

type CreateSectionInput struct {
	Title string `json:"title" binding:"required"`
}

type Section struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Sort      int    `json:"sort"`
	CreatedAt int64  `json:"created_at"` // ✅ timestamp 格式（Unix 秒數）
	UpdatedAt int64  `json:"updated_at"`
}
