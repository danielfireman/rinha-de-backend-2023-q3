package main

import (
	"time"

	"github.com/dgraph-io/ristretto"
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

	// Configurações de cache vindas de https://github.com/dgraph-io/ristretto
	// Só diminui o tamanho máximo do cache.
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,    // number of keys to track frequency of (10M).
		MaxCost:     1 << 5, // maximum cost of cache (0,15GB).
		BufferItems: 64,     // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}

	// [PerfNote] Criando um RPC Stub do rinha DB por tipo de chamada, pois todas
	// são estressadas. Perftip vinda de https://grpc.io/docs/guides/performance/
	e.POST("/pessoas", newCriarPessoa(db, MustNewRinhaDB(), cache).handler)
	e.GET("/pessoas", func(c echo.Context) error { return echo.ErrNotFound })
	e.GET("/pessoas", buscaPessoas{MustNewRinhaDB()}.handler)
	e.GET("/pessoas/:id", getPessoa{MustNewRinhaDB(), cache}.handler)
	e.GET("/contagem-pessoas", contarPessoas{db}.handler)

	e.Logger.Fatal(e.Start(":8080"))
}
