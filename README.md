# MyBookListAPI
## Uma lista de livros já lidos

#### Ideia principal do projeto

Uma API que guarda em um banco de dados livros que o usuário planeja ler, está lendo ou já leu.

Esta API permitirá registrar livros com título, nome do autor, ano de publicação, status (planejando, lendo, lido), datas de adição, de início da leitura e conclusão, última atualização, uma nota de 0 a 10 e uma nota de texto opcional caso o usuário queira dizer algo sobre o livro. Apenas título e nome do autor serão obrigatórios. As datas serão todas automáticas.

## Tabelas do banco de dados

#### Tabela book

| Id | title | author | year | status | dt_add | dt_start | dt_conclusion | last_update | rating | note |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| PK, int, nn, auto | text, nn | text, nn | int | string, (planning, reading, read) | time, nn, auto | time | time | time, auto | float | text |

## Rotas da API

#### Criar livro

`POST /book` Recebe dados de um livro e os armazena, criando um novo item na tabela.

Dados obrigatórios: `tittle:string, author: string`

Dados opcionais: `year: int, status: string, dt_start: time, dt_conclusion: time, rating: float, note: string`

#### Atualizar livro

`PATCH /book/{id}` Recebe campos a serem atualizados e identifica o livro pelo id passado no path.

Dado pathvalue: `id: int` devolver erro caso o id seja inválido ou não existente.

Dados da requisição: `tittle: string, author: string, year: int, status: string, dt_add: time, dt_start: time, dt_conclusion: time, rating: float, note: string` todos opcionais.

#### Deletar livro

`DELETE /book/{id}` Deleta um livro pelo id passado na url.

Dado pathvalue: `id: int` devolver erro caso o id seja inválido ou não existente.

#### Listar todos os livros

`GET /books` Retorna todos os livros armazenados.

#### Listar livros com filtros

`GET /books?tittle=&author=&year=&status=&dt_add=&dt_start=&dt_conclusion=&rating=` 

Transforma cada campo GET passado na url em parâmetro de busca no banco e retorna os livros baseados nestes filtros.

#### Pegar um livro específico

`GET /book/{id}` Retorna o livro com aquele id. Retorna erro em caso de Id inválido ou inexistente.

## Persistência de dados

#### PostgreSQL e Docker

O banco de dados será PostgreSQL e rodará dentro de um container docker.