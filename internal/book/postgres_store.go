package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(ctx context.Context, b *Book) error {
	if b.Status == "" {
		b.Status = StatusPlanning
	}
	const query = `
        INSERT INTO books (title, author, year, status, dt_start, dt_conclusion, rating, note)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, dt_add, last_update`
	err := s.db.QueryRowContext(ctx, query, b.Title, b.Author, b.Year, b.Status, b.DtStart,
		b.DtConclusion, b.Rating, b.Note).Scan(&b.ID, &b.DtAdd, &b.LastUpdate)

	if err != nil {
		return fmt.Errorf("creating book: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id int) (*Book, error) {
	//o handler já deve ter feito a validação do id, então vamos usá-lo diretamente
	const query = `
	SELECT id, title, author, year, status, dt_add, dt_start, dt_conclusion, last_update, rating, note 
	FROM books WHERE id = $1`

	book := &Book{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
		&book.Status,
		&book.DtAdd,
		&book.DtStart,
		&book.DtConclusion,
		&book.LastUpdate,
		&book.Rating,
		&book.Note,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting book by id: %w", err)
	}

	return book, nil
}

func (s *PostgresStore) List(ctx context.Context, f Filters) ([]Book, error) {
	query := `SELECT id, title, author, year, status, dt_add, dt_start, dt_conclusion, last_update, rating, note FROM books`

	var conditions []string
	var args []any
	argPos := 1

	if f.Title != nil {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argPos))
		args = append(args, "%"+*f.Title+"%")
		argPos++
	}
	if f.Author != nil {
		conditions = append(conditions, fmt.Sprintf("author ILIKE $%d", argPos))
		args = append(args, "%"+*f.Author+"%")
		argPos++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *f.Status)
		argPos++
	}
	if f.Year != nil {
		conditions = append(conditions, fmt.Sprintf("year = $%d", argPos))
		args = append(args, *f.Year)
		argPos++
	}

	if f.DtStart.From != nil {
		conditions = append(conditions, fmt.Sprintf("dt_start::date >= $%d", argPos))
		args = append(args, f.DtStart.From.Format("2006-01-02"))
		argPos++
	}
	if f.DtStart.To != nil {
		conditions = append(conditions, fmt.Sprintf("dt_start::date <= $%d", argPos))
		args = append(args, f.DtStart.To.Format("2006-01-02"))
		argPos++
	}

	if f.DtAdd.From != nil {
		conditions = append(conditions, fmt.Sprintf("dt_add::date >= $%d", argPos))
		args = append(args, f.DtAdd.From.Format("2006-01-02"))
		argPos++
	}
	if f.DtAdd.To != nil {
		conditions = append(conditions, fmt.Sprintf("dt_add::date <= $%d", argPos))
		args = append(args, f.DtAdd.To.Format("2006-01-02"))
		argPos++
	}

	if f.DtConclusion.From != nil {
		conditions = append(conditions, fmt.Sprintf("dt_conclusion::date >= $%d", argPos))
		args = append(args, f.DtConclusion.From.Format("2006-01-02"))
		argPos++
	}
	if f.DtConclusion.To != nil {
		conditions = append(conditions, fmt.Sprintf("dt_conclusion::date <= $%d", argPos))
		args = append(args, f.DtConclusion.To.Format("2006-01-02"))
		argPos++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing books: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Status,
			&b.DtAdd, &b.DtStart, &b.DtConclusion, &b.LastUpdate, &b.Rating, &b.Note); err != nil {
			return nil, fmt.Errorf("scanning book row: %w", err)
		}
		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating book rows: %w", err)
	}

	return books, nil
}

func (s *PostgresStore) Update(ctx context.Context, b *Book) (*Book, error) {
	const query = `
		UPDATE books
		SET title = $1, author = $2, year = $3, status = $4, dt_start = $5, dt_conclusion = $6,
		rating = $7, note = $8, last_update = now()
		WHERE id = $9 RETURNING last_update`
	err := s.db.QueryRowContext(ctx, query, b.Title, b.Author, b.Year, b.Status, b.DtStart, b.DtConclusion, b.Rating,
		b.Note, b.ID).Scan(&b.LastUpdate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("updating book: %w", err)
	}

	return b, nil
}

func (s *PostgresStore) Delete(ctx context.Context, id int) error {
	const query = "DELETE FROM books WHERE id = $1"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting book: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
