CREATE TYPE book_status AS ENUM ('planning', 'reading', 'read');

CREATE TABLE IF NOT EXISTS books (
    id             SERIAL PRIMARY KEY,
    title          TEXT NOT NULL,
    author         TEXT NOT NULL,
    year           INTEGER,
    status         book_status NOT NULL DEFAULT 'planning',
    dt_add         TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_start       TIMESTAMPTZ,
    dt_conclusion  TIMESTAMPTZ,
    last_update    TIMESTAMPTZ NOT NULL DEFAULT now(),
    rating         REAL CHECK (rating >= 0 AND rating <= 10),
    note           TEXT
);
-- Banco de testes
CREATE DATABASE booklist_test;
\c booklist_test

CREATE TYPE book_status AS ENUM ('planning', 'reading', 'read');

CREATE TABLE IF NOT EXISTS books (
    id             SERIAL PRIMARY KEY,
    title          TEXT NOT NULL,
    author         TEXT NOT NULL,
    year           INTEGER,
    status         book_status NOT NULL DEFAULT 'planning',
    dt_add         TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_start       TIMESTAMPTZ,
    dt_conclusion  TIMESTAMPTZ,
    last_update    TIMESTAMPTZ NOT NULL DEFAULT now(),
    rating         REAL CHECK (rating >= 0 AND rating <= 10),
    note           TEXT
);