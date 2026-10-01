package book

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type createBookRequest struct {
	Title        string     `json:"title"`
	Author       string     `json:"author"`
	Year         *int       `json:"year,omitempty"`
	Status       Status     `json:"status,omitempty"`
	DtStart      *time.Time `json:"dt_start,omitempty"`
	DtConclusion *time.Time `json:"dt_conclusion,omitempty"`
	Rating       *float64   `json:"rating,omitempty"`
	Note         string     `json:"note,omitempty"`
}

type patchBookRequest struct {
	Title        *string    `json:"title,omitempty"`
	Author       *string    `json:"author,omitempty"`
	Year         *int       `json:"year,omitempty"`
	Status       *Status    `json:"status,omitempty"`
	DtStart      *time.Time `json:"dt_start,omitempty"`
	DtConclusion *time.Time `json:"dt_conclusion,omitempty"`
	Rating       *float64   `json:"rating,omitempty"`
	Note         *string    `json:"note,omitempty"`
}

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJson(w, status, map[string]string{"error": message})
}

func CreateHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createBookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Author == "" || req.Title == "" {
			writeError(w, http.StatusBadRequest, "title and author are required")
			return
		}

		b := &Book{
			Title:        req.Title,
			Author:       req.Author,
			Year:         req.Year,
			Status:       req.Status,
			DtStart:      req.DtStart,
			DtConclusion: req.DtConclusion,
			Rating:       req.Rating,
			Note:         req.Note,
		}

		if err := store.Create(r.Context(), b); err != nil {
			writeError(w, http.StatusInternalServerError, "could not create book")
			return
		}

		writeJson(w, http.StatusCreated, b)
	}
}

func UpdateHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req patchBookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		existing, err := store.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "book not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not fetch book")
			return
		}

		if req.Title != nil {
			if *req.Title == "" {
				writeError(w, http.StatusBadRequest, "title cannot be empty")
				return
			}
			existing.Title = *req.Title
		}
		if req.Author != nil {
			if *req.Author == "" {
				writeError(w, http.StatusBadRequest, "author cannot be empty")
				return
			}
			existing.Author = *req.Author
		}
		if req.Year != nil {
			existing.Year = req.Year
		}
		if req.Status != nil {
			if !req.Status.Valid() {
				writeError(w, http.StatusBadRequest, "invalid status")
				return
			}
			existing.Status = *req.Status
		}
		if req.DtStart != nil {
			existing.DtStart = req.DtStart
		}
		if req.DtConclusion != nil {
			existing.DtConclusion = req.DtConclusion
		}
		if req.Rating != nil {
			existing.Rating = req.Rating
		}
		if req.Note != nil {
			existing.Note = *req.Note
		}

		updated, err := store.Update(r.Context(), existing)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not update book")
			return
		}
		writeJson(w, http.StatusOK, updated)
	}
}

func DeleteHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := store.Delete(r.Context(), id); err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "book not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not delete book")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetByIDHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		book, err := store.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "book not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not fetch book")
			return
		}
		writeJson(w, http.StatusOK, book)
	}
}

func ListHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var f Filters

		if title := q.Get("title"); title != "" {
			f.Title = &title
		}
		if author := q.Get("author"); author != "" {
			f.Author = &author
		}
		if status := q.Get("status"); status != "" {
			s := Status(status)
			if !s.Valid() {
				writeError(w, http.StatusBadRequest, "invalid status")
				return
			}
			f.Status = &s
		}
		if v := q.Get("year"); v != "" {
			year, err := strconv.Atoi(v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid year filter")
				return
			}
			f.Year = &year
		}
		err := getAllDates(q, &f)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		books, err := store.List(r.Context(), f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not fetch books")
			return
		}
		writeJson(w, http.StatusOK, books)
	}
}

func getAllDates(q url.Values, f *Filters) error {
	var err error
	f.DtAdd.From, err = parseDateFilter(q, "from_dt_add")
	if err != nil {
		return fmt.Errorf("invalid from_dt_add: %w", err)
	}
	f.DtAdd.To, err = parseDateFilter(q, "to_dt_add")
	if err != nil {
		return fmt.Errorf("invalid to_dt_add: %w", err)
	}
	f.DtStart.From, err = parseDateFilter(q, "from_dt_start")
	if err != nil {
		return fmt.Errorf("invalid from_dt_start: %w", err)
	}
	f.DtStart.To, err = parseDateFilter(q, "to_dt_start")
	if err != nil {
		return fmt.Errorf("invalid to_dt_start: %w", err)
	}
	f.DtConclusion.From, err = parseDateFilter(q, "from_dt_conclusion")
	if err != nil {
		return fmt.Errorf("invalid from_dt_conclusion: %w", err)
	}
	f.DtConclusion.To, err = parseDateFilter(q, "to_dt_conclusion")
	if err != nil {
		return fmt.Errorf("invalid to_dt_conclusion: %w", err)
	}
	return nil
}

func parseDateFilter(q url.Values, key string) (*time.Time, error) {
	v := q.Get(key)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
