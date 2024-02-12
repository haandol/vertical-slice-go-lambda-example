package domain

import "github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/dto"

type ClickEvent struct {
	PK        string `json:"PK"`
	SK        string `json:"SK"`
	ID        string `json:"id"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
}

func (d *ClickEvent) DTO() dto.ClickEvent {
	return dto.ClickEvent{
		ID:        d.ID,
		Path:      d.Path,
		CreatedAt: d.CreatedAt,
	}
}
