package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/MarlonSdS/MyBookListAPI/internal/book"
	"github.com/MarlonSdS/MyBookListAPI/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Conectado ao banco com sucesso")

	store := book.NewPostgresStore(db)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /book", book.CreateHandler(store))
	mux.HandleFunc("GET /book/{id}", book.GetByIDHandler(store))
	mux.HandleFunc("PATCH /book/{id}", book.UpdateHandler(store))
	mux.HandleFunc("DELETE /book/{id}", book.DeleteHandler(store))
	mux.HandleFunc("GET /books", book.ListHandler(store))

	log.Println("servidor rodando na porta 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
