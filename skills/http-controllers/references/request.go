package documents

import (
	"fmt"

	"<module>/internal/shared"
)

// CreateRequest — DTO транспорта. JSON-теги живут только здесь.
type CreateRequest struct {
	Title string `json:"title" example:"Акт осмотра"`
	Kind  string `json:"kind" example:"inspection"`
}

func (r CreateRequest) toInput() (shared.CreateDocumentInput, error) {
	if r.Title == "" {
		return shared.CreateDocumentInput{}, fmt.Errorf("title is required")
	}
	if r.Kind == "" {
		return shared.CreateDocumentInput{}, fmt.Errorf("kind is required")
	}
	return shared.CreateDocumentInput{
		Title: r.Title,
		Kind:  r.Kind,
	}, nil
}
