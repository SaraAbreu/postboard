package store

import (
	"context"
	"errors"

	"github.com/SaraAbreu/postboard/api/internal/models"
)

var ErrNotFound = errors.New("not found")

// PostStore es la interfaz que usan los handlers.
// Gracias a ella podemos usar Postgres en producción y memoria en los tests.
type PostStore interface {
	List(ctx context.Context) ([]models.Post, error)
	Get(ctx context.Context, id int64) (models.Post, error)
	Create(ctx context.Context, in models.PostInput) (models.Post, error)
	Update(ctx context.Context, id int64, in models.PostInput) (models.Post, error)
	Delete(ctx context.Context, id int64) error
}
