package book_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/MarlonSdS/MyBookListAPI/internal/book"
	"github.com/MarlonSdS/MyBookListAPI/internal/config"
)

func TestMain(m *testing.M) {
	_, currentFile, _, _ := runtime.Caller(0)
	// currentFile = .../MyBookListAPI/internal/book/postgres_store_test.go
	rootDir := filepath.Join(filepath.Dir(currentFile), "..", "..")
	envPath := filepath.Join(rootDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		panic("erro ao carregar .env para os testes: " + err.Error())
	}

	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.TestDSN())
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("pinging test db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
func TestCreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}

	err := store.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() {
		store.Delete(ctx, b.ID)
	})

	if b.ID == 0 {
		t.Error("expected ID to be set after Create")
	}
	if b.Status != book.StatusPlanning {
		t.Errorf("expected default status %q, got %q", book.StatusPlanning, b.Status)
	}

	got, err := store.GetByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != b.Title {
		t.Errorf("expected title %q, got %q", b.Title, got.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	_, err := store.GetByID(ctx, 999999999)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, book.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestListWithFilters(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b1 := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}
	b2 := &book.Book{
		Title:  "Encontro com Rama",
		Author: "Arthur C. Clark",
	}
	err := store.Create(ctx, b1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	err = store.Create(ctx, b2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() {
		store.Delete(ctx, b1.ID)
		store.Delete(ctx, b2.ID)
	})
	title := "Rama"
	from := time.Now().AddDate(0, 0, -1) // ontem
	to := time.Now().AddDate(0, 0, 1)    // amanhã
	f := book.Filters{
		Title: &title,
		DtAdd: book.DateRange{From: &from, To: &to},
	}
	books, err := store.List(ctx, f)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	} else {
		if len(books) != 1 {
			t.Errorf("Expected just one result, found: %d", len(books))
			return
		}
		if len(books) == 0 {
			t.Fatalf("Nothing found")
			return
		}
		if books[0].Title != "Encontro com Rama" {
			t.Fatalf("Unexpected Result")
		}
	}

	//só pra ver como ficou os resultados
	//fmt.Println(books)
}

func TestListWithNoFilters(t *testing.T) {
	// a ideia é retornar todas as linhas na tabela
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b1 := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}
	b2 := &book.Book{
		Title:  "Encontro com Rama",
		Author: "Arthur C. Clark",
	}
	err := store.Create(ctx, b1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	err = store.Create(ctx, b2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() {
		store.Delete(ctx, b1.ID)
		store.Delete(ctx, b2.ID)
	})
	f := book.Filters{}
	books, err := store.List(ctx, f)
	if err != nil {
		t.Fatalf("List() error = %v", err)
		return
	}
	if len(books) != 2 {
		t.Errorf("Expected 2 results, found %d", len(books))
	}
	if len(books) == 0 {
		t.Errorf("Nothing found")
	}
	if books[0].Title != "O Hobbit" || books[1].Title != "Encontro com Rama" {
		t.Errorf("Unexpected Results")
	}
	//apenas pra ver como ficou
	//fmt.Println(books)
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}

	err := store.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() {
		store.Delete(ctx, b.ID)
	})
	previousTitle := b.Title
	previousAuthor := b.Author
	b.Title = "O Hobbit - Volume com Prefácio"
	b.Author = "Tolkien"
	previousLastUpdate := b.LastUpdate
	updatedBook, err := store.Update(ctx, b)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			t.Errorf("Book not found: %v", err)
			return
		}
		t.Errorf("Update() error = %v", err)
		return
	}
	if updatedBook.LastUpdate == previousLastUpdate {
		t.Errorf("last_update remains the same")
	}
	if updatedBook.Title == previousTitle && updatedBook.Author == previousAuthor {
		t.Errorf("Title and author not updated to %s and %s", b.Title, b.Author)
	}
}

func TestUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}

	err := store.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	prevID := b.ID
	t.Cleanup(func() {
		store.Delete(ctx, prevID)
	})
	b.ID = b.ID + 1
	updatedBook, err := store.Update(ctx, b)
	if err != nil {
		if !errors.Is(err, book.ErrNotFound) {
			t.Errorf("Expected Error Not Found, received: %v", err)
			return
		}
	} else {
		t.Errorf("Expected Error Not Found, but a book was returned: %v", updatedBook)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}

	err := store.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	prevID := b.ID
	t.Cleanup(func() {
		store.Delete(ctx, prevID)
	})
	err = store.Delete(ctx, prevID)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			t.Errorf("Book not found: %v", err)
			return
		}
		t.Errorf("Delete() error = %v", err)
	}
	_, err = store.GetByID(ctx, prevID)
	if err == nil {
		t.Fatalf("No error while trying to get the book, meaning the book was not deleted")
	} else {
		if !errors.Is(err, book.ErrNotFound) {
			t.Fatalf("Another type of error that is not ErrNotFound: %v", err)
		}
	}
}

func TestDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	store := book.NewPostgresStore(db)
	ctx := context.Background()

	b := &book.Book{
		Title:  "O Hobbit",
		Author: "J.R.R. Tolkien",
	}

	err := store.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	prevID := b.ID
	t.Cleanup(func() {
		store.Delete(ctx, prevID)
	})
	err = store.Delete(ctx, prevID+5555)
	if err == nil {
		t.Fatal("No error returned")
	}
	if !errors.Is(err, book.ErrNotFound) {
		t.Fatalf("Expected Not Found, another type of erro ocurred: %v", err)
	}
}
