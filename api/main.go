package main

import (
	"time"

	"github.com/alphadose/haxmap"
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

	cache := haxmap.New[string, string](1e5)

	// [PerfNote] Criando um RPC Stub do rinha DB por tipo de chamada, pois todas
	// são estressadas. Perftip vinda de https://grpc.io/docs/guides/performance/
	e.POST("/pessoas", newCriarPessoa(db, MustNewRinhaDB(), cache).handler)
	e.GET("/pessoas", func(c echo.Context) error { return echo.ErrNotFound })
	e.GET("/pessoas", buscaPessoas{MustNewRinhaDB()}.handler)
	e.GET("/pessoas/:id", getPessoa{MustNewRinhaDB(), cache}.handler)
	e.GET("/contagem-pessoas", contarPessoas{db}.handler)

	e.Logger.Fatal(e.Start(":8080"))
}
