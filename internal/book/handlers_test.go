package book_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MarlonSdS/MyBookListAPI/internal/book"
)

// fake que implementa book.Store só o suficiente pra compilar;
// se Create for chamado, o teste deveria falhar - por isso o t.Fatal.
type fakeStore struct {
	t *testing.T
}

func (f *fakeStore) Create(ctx context.Context, b *book.Book) error {
	f.t.Fatal("Create não deveria ter sido chamado")
	return nil
}
func (f *fakeStore) GetByID(ctx context.Context, id int) (*book.Book, error) {
	panic("not implemented")
}
func (f *fakeStore) List(ctx context.Context, filters book.Filters) ([]book.Book, error) {
	panic("not implemented")
}
func (f *fakeStore) Update(ctx context.Context, b *book.Book) (*book.Book, error) {
	panic("not implemented")
}
func (f *fakeStore) Delete(ctx context.Context, id int) error { panic("not implemented") }

func TestCreateHandler_MissingTitle(t *testing.T) {
	store := &fakeStore{t: t}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"author": "Autor Sem Título"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] == "" {
		t.Error("expected an error message in response")
	}
}
