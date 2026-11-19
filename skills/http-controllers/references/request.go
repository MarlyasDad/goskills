package documents

import (
	"<module>/internal/shared"
	"<module>/internal/transport/<appname>/http/responses"
)

// CreateRequest — DTO транспорта. JSON-теги живут только здесь.
type CreateRequest struct {
	Title string `json:"title" example:"Акт осмотра"`
	Kind  string `json:"kind" example:"inspection"`
}

// validate проверяет формат полей. Бизнес-валидация — в сервисе или домене.
func (r CreateRequest) validate() responses.ValidationErrors {
	var errs responses.ValidationErrors

	if r.Title == "" {
		errs = append(errs, responses.ValidationError{
			Message: "title: value is required",
			Code:    responses.CodeUnassigned,
		})
	}
	if r.Kind == "" {
		errs = append(errs, responses.ValidationError{
			Message: "kind: value is required",
			Code:    responses.CodeUnassigned,
		})
	}

	return errs
}

func (r CreateRequest) toInput() shared.CreateDocumentInput {
	return shared.CreateDocumentInput{
		Title: r.Title,
		Kind:  r.Kind,
	}
}
