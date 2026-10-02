package book_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MarlonSdS/MyBookListAPI/internal/book"
)

// errStoreFailure simula um erro genérico de banco/infra (não ErrNotFound),
// usado nos testes que esperam 500.
var errStoreFailure = errors.New("simulated store failure")

// fakeStore implementa book.Store com funções configuráveis por teste.
// Se um método for chamado sem a função correspondente ter sido setada,
// o panic denuncia que o handler fez uma chamada inesperada.
type fakeStore struct {
	createFunc  func(ctx context.Context, b *book.Book) error
	getByIDFunc func(ctx context.Context, id int) (*book.Book, error)
	listFunc    func(ctx context.Context, f book.Filters) ([]book.Book, error)
	updateFunc  func(ctx context.Context, b *book.Book) (*book.Book, error)
	deleteFunc  func(ctx context.Context, id int) error
}

func (f *fakeStore) Create(ctx context.Context, b *book.Book) error {
	if f.createFunc == nil {
		panic("createFunc not set")
	}
	return f.createFunc(ctx, b)
}

func (f *fakeStore) GetByID(ctx context.Context, id int) (*book.Book, error) {
	if f.getByIDFunc == nil {
		panic("getByIDFunc not set")
	}
	return f.getByIDFunc(ctx, id)
}

func (f *fakeStore) List(ctx context.Context, filters book.Filters) ([]book.Book, error) {
	if f.listFunc == nil {
		panic("listFunc not set")
	}
	return f.listFunc(ctx, filters)
}

func (f *fakeStore) Update(ctx context.Context, b *book.Book) (*book.Book, error) {
	if f.updateFunc == nil {
		panic("updateFunc not set")
	}
	return f.updateFunc(ctx, b)
}

func (f *fakeStore) Delete(ctx context.Context, id int) error {
	if f.deleteFunc == nil {
		panic("deleteFunc not set")
	}
	return f.deleteFunc(ctx, id)
}

// decodeError lê {"error": "..."} do corpo da resposta.
func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return resp["error"]
}

// ---------- CreateHandler ----------

func TestCreateHandler_Success(t *testing.T) {
	var created *book.Book
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			b.ID = 1 // simula o que o Postgres faria via RETURNING
			created = b
			return nil
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"title":"O Hobbit","author":"J.R.R. Tolkien"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if created == nil || created.Title != "O Hobbit" {
		t.Fatalf("expected Create to be called with the decoded book")
	}

	var got book.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("expected response ID=1, got %d", got.ID)
	}
}

func TestCreateHandler_MissingTitle(t *testing.T) {
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			t.Fatal("Create não deveria ter sido chamado")
			return nil
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"author":"Autor Sem Título"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if decodeError(t, rec) == "" {
		t.Error("expected an error message in response")
	}
}

func TestCreateHandler_MissingAuthor(t *testing.T) {
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			t.Fatal("Create não deveria ter sido chamado")
			return nil
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"title":"Livro Sem Autor"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateHandler_InvalidStatus(t *testing.T) {
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			t.Fatal("Create não deveria ter sido chamado")
			return nil
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"title":"X","author":"Y","status":"lendo"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateHandler_InvalidBody(t *testing.T) {
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			t.Fatal("Create não deveria ter sido chamado")
			return nil
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateHandler_StoreError(t *testing.T) {
	store := &fakeStore{
		createFunc: func(ctx context.Context, b *book.Book) error {
			return errStoreFailure
		},
	}
	handler := book.CreateHandler(store)

	body := strings.NewReader(`{"title":"X","author":"Y"}`)
	req := httptest.NewRequest(http.MethodPost, "/book", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// ---------- GetByIDHandler ----------

func TestGetByIDHandler_Success(t *testing.T) {
	want := &book.Book{ID: 1, Title: "O Hobbit", Author: "J.R.R. Tolkien"}

	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			if id != 1 {
				t.Errorf("expected id=1, got %d", id)
			}
			return want, nil
		},
	}
	handler := book.GetByIDHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/book/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got book.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Title != want.Title {
		t.Errorf("expected title %q, got %q", want.Title, got.Title)
	}
}

func TestGetByIDHandler_NotFound(t *testing.T) {
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return nil, book.ErrNotFound
		},
	}
	handler := book.GetByIDHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/book/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetByIDHandler_InvalidID(t *testing.T) {
	store := &fakeStore{} // nenhum método deveria ser chamado
	handler := book.GetByIDHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/book/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetByIDHandler_StoreError(t *testing.T) {
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return nil, errStoreFailure
		},
	}
	handler := book.GetByIDHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/book/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// ---------- DeleteHandler ----------

func TestDeleteHandler_Success(t *testing.T) {
	called := false
	store := &fakeStore{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 1 {
				t.Errorf("expected id=1, got %d", id)
			}
			return nil
		},
	}
	handler := book.DeleteHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/book/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
	if !called {
		t.Error("expected Delete to be called")
	}
}

func TestDeleteHandler_NotFound(t *testing.T) {
	store := &fakeStore{
		deleteFunc: func(ctx context.Context, id int) error {
			return book.ErrNotFound
		},
	}
	handler := book.DeleteHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/book/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDeleteHandler_InvalidID(t *testing.T) {
	store := &fakeStore{}
	handler := book.DeleteHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/book/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteHandler_StoreError(t *testing.T) {
	store := &fakeStore{
		deleteFunc: func(ctx context.Context, id int) error {
			return errStoreFailure
		},
	}
	handler := book.DeleteHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/book/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// ---------- UpdateHandler ----------

func TestUpdateHandler_Success(t *testing.T) {
	existing := &book.Book{ID: 1, Title: "Título Antigo", Author: "Autor Antigo"}

	var passedToUpdate *book.Book
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return existing, nil
		},
		updateFunc: func(ctx context.Context, b *book.Book) (*book.Book, error) {
			passedToUpdate = b
			return b, nil
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"title":"Título Novo"}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if passedToUpdate == nil {
		t.Fatal("expected Update to be called")
	}
	// campo enviado foi alterado...
	if passedToUpdate.Title != "Título Novo" {
		t.Errorf("expected title to be updated, got %q", passedToUpdate.Title)
	}
	// ...campo não enviado permaneceu intacto (o teste central desta função)
	if passedToUpdate.Author != "Autor Antigo" {
		t.Errorf("expected author to remain unchanged, got %q", passedToUpdate.Author)
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return nil, book.ErrNotFound
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"title":"X"}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/999", body)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateHandler_InvalidID(t *testing.T) {
	store := &fakeStore{}
	handler := book.UpdateHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/book/abc", strings.NewReader(`{}`))
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateHandler_InvalidBody(t *testing.T) {
	store := &fakeStore{}
	handler := book.UpdateHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/book/1", strings.NewReader(`not json`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateHandler_EmptyTitle(t *testing.T) {
	existing := &book.Book{ID: 1, Title: "Título Antigo", Author: "Autor Antigo"}
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return existing, nil
		},
		updateFunc: func(ctx context.Context, b *book.Book) (*book.Book, error) {
			t.Fatal("Update não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"title":""}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateHandler_EmptyAuthor(t *testing.T) {
	existing := &book.Book{ID: 1, Title: "Título Antigo", Author: "Autor Antigo"}
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return existing, nil
		},
		updateFunc: func(ctx context.Context, b *book.Book) (*book.Book, error) {
			t.Fatal("Update não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"author":""}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateHandler_InvalidStatus(t *testing.T) {
	existing := &book.Book{ID: 1, Title: "T", Author: "A"}
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return existing, nil
		},
		updateFunc: func(ctx context.Context, b *book.Book) (*book.Book, error) {
			t.Fatal("Update não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"status":"xpto"}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateHandler_StoreError(t *testing.T) {
	existing := &book.Book{ID: 1, Title: "T", Author: "A"}
	store := &fakeStore{
		getByIDFunc: func(ctx context.Context, id int) (*book.Book, error) {
			return existing, nil
		},
		updateFunc: func(ctx context.Context, b *book.Book) (*book.Book, error) {
			return nil, errStoreFailure
		},
	}
	handler := book.UpdateHandler(store)

	body := strings.NewReader(`{"title":"Novo"}`)
	req := httptest.NewRequest(http.MethodPatch, "/book/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// ---------- ListHandler ----------

func TestListHandler_NoFilters(t *testing.T) {
	want := []book.Book{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}

	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			if f.Title != nil || f.Author != nil || f.Status != nil || f.Year != nil {
				t.Error("expected empty filters")
			}
			return want, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got []book.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 books, got %d", len(got))
	}
}

func TestListHandler_WithFilters(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			if f.Title == nil || *f.Title != "Hobbit" {
				t.Errorf("expected title filter = Hobbit, got %v", f.Title)
			}
			if f.Author == nil || *f.Author != "Tolkien" {
				t.Errorf("expected author filter = Tolkien, got %v", f.Author)
			}
			if f.Status == nil || *f.Status != book.StatusRead {
				t.Errorf("expected status filter = read, got %v", f.Status)
			}
			if f.Year == nil || *f.Year != 1937 {
				t.Errorf("expected year filter = 1937, got %v", f.Year)
			}
			return []book.Book{}, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet,
		"/books?title=Hobbit&author=Tolkien&status=read&year=1937", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListHandler_DateRangeFilter(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			if f.DtAdd.From == nil || f.DtAdd.To == nil {
				t.Fatal("expected DtAdd.From and DtAdd.To to be set")
			}
			if f.DtAdd.From.Format("2006-01-02") != "2026-01-01" {
				t.Errorf("unexpected DtAdd.From: %v", f.DtAdd.From)
			}
			if f.DtAdd.To.Format("2006-01-02") != "2026-01-31" {
				t.Errorf("unexpected DtAdd.To: %v", f.DtAdd.To)
			}
			return []book.Book{}, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet,
		"/books?from_dt_add=2026-01-01&to_dt_add=2026-01-31", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListHandler_InvalidYear(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			t.Fatal("List não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/books?year=abc", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestListHandler_InvalidStatus(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			t.Fatal("List não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/books?status=xpto", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestListHandler_InvalidDate(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			t.Fatal("List não deveria ter sido chamado")
			return nil, nil
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/books?from_dt_add=não-é-data", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if got := decodeError(t, rec); got == "" {
		t.Error("expected an error message naming the invalid field")
	}
}

func TestListHandler_StoreError(t *testing.T) {
	store := &fakeStore{
		listFunc: func(ctx context.Context, f book.Filters) ([]book.Book, error) {
			return nil, errStoreFailure
		},
	}
	handler := book.ListHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
