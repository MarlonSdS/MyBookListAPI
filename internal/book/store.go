package book

import (
	"context"
	"errors"
	"time"
)

// Erro de livro não encontrado
var ErrNotFound = errors.New("book not found")

type Store interface {
	Create(ctx context.Context, b *Book) error
	GetByID(ctx context.Context, id int) (*Book, error)
	List(ctx context.Context, filters Filters) ([]Book, error)
	Update(ctx context.Context, id int, b *Book) error
	Delete(ctx context.Context, id int) error
}

type DateRange struct {
	From *time.Time
	To   *time.Time
}

type Filters struct {
	Title        *string
	Author       *string
	Status       *Status
	Year         *int
	Rating       *float64
	DtAdd        DateRange
	DtStart      DateRange
	DtConclusion DateRange
}
