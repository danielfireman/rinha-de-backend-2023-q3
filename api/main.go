package main

import (
	"github.com/labstack/echo/v4"
)

type Pessoa struct {
	ID         string   `json:"id"`
	Apelido    *string  `json:"apelido"`
	Nome       *string  `json:"nome"`
	Nascimento *string  `json:"nascimento"`
	Stack      []string `json:"stack"`
}

func main() {
	e := echo.New()
	e.JSONSerializer = newSerializer() // using a faster JSON serializer.

	e.Logger.SetLevel(1) // 1: INFO, 0: DEBUG

	db := NewInMemDB()
	e.POST("/pessoas", criarPessoa{db}.handler)
	e.GET("/pessoas/:id", getPessoa{db}.handler)
	e.GET("/contagem-pessoas", contarPessoas{db}.handler)
	e.GET("/pessoas", buscaPessoas{db}.handler)
	e.Logger.Fatal(e.Start(":8080"))
}
