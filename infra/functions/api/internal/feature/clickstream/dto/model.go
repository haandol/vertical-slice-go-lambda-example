package dto

type ClickEvent struct {
	ID        string `json:"id"`
	Path      string `json:"path" binding:"required" validate:"required"`
	CreatedAt string `json:"createdAt"`
}
