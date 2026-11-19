package documents

import (
	"context"

	"<module>/internal/shared"
)

// Creator — операция, которую контроллер вызывает у сервиса.
// Интерфейс объявлен здесь, у потребителя.
type Creator interface {
	Create(ctx context.Context, input shared.CreateDocumentInput) (shared.CreateDocumentOutput, error)
}
