package book

import (
	"encoding/json"
	"net/http"
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
