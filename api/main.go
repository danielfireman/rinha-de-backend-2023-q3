package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 1 * time.Minute,
	}))

	db := MustNewMongoDB()
	cache := MustNewCache()
	e.POST("/pessoas", criarPessoa{db, cache}.handler)
	e.GET("/pessoas", func(c echo.Context) error { return echo.ErrNotFound })
	e.GET("/pessoas", buscaPessoas{cache}.handler)
	e.GET("/pessoas/:id", getPessoa{cache}.handler)
	e.GET("/contagem-pessoas", contarPessoas{db}.handler)

	e.Logger.Fatal(e.Start(":8080"))
}
