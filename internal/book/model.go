package book

import (
	"time"
)

type Status string

const (
	StatusPlanning Status = "planning"
	StatusReading  Status = "reading"
	StatusRead     Status = "read"
)

type Book struct {
	ID           int        `json:"id"`
	Title        string     `json:"title,omitempty"`
	Author       string     `json:"author,omitempty"`
	Year         *int       `json:"year,omitempty"`
	Status       Status     `json:"status,omitempty"`
	DtAdd        *time.Time `json:"dt_add,omitempty"`
	DtStart      *time.Time `json:"dt_start,omitempty"`
	DtConclusion *time.Time `json:"dt_conclusion,omitempty"`
	LastUpdate   *time.Time `json:"last_update,omitempty"`
	Rating       *float64   `json:"rating,omitempty"`
	Note         string     `json:"note,omitempty"`
}
