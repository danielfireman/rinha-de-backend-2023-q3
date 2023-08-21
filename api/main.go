package main

import (
	"time"

	"github.com/alphadose/haxmap"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

type Pessoa struct {
	ID         string   `json:"id"`
	Apelido    *string  `json:"apelido"`
	Nome       *string  `json:"nome"`
	Nascimento *string  `json:"nascimento"`
	Stack      []string `json:"stack"`
}

func main() {

	// [PerfNote] Usando sonic para json marshal e unmarshal.
	app := fiber.New(fiber.Config{
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		JSONEncoder:  sonic.ConfigFastest.Marshal,
		JSONDecoder:  sonic.ConfigFastest.Unmarshal,
	})
	db := MustNewMongoDB()

	cache := haxmap.New[string, string](1e5)
	apelidoCache := haxmap.New[string, struct{}](1e5)

	// [PerfNote] Criando um RPC Stub do rinha DB por tipo de chamada, pois todas
	// são estressadas. Perftip vinda de https://grpc.io/docs/guides/performance/
	app.Post("/pessoas", newCriarPessoa(db, MustNewRinhaDB(), cache, apelidoCache).handler)
	app.Get("/pessoas/:id", getPessoa{MustNewRinhaDB(), cache, apelidoCache}.handler)
	app.Get("/pessoas", buscaPessoas{MustNewRinhaDB()}.handler)
	app.Get("/contagem-pessoas", contarPessoas{db}.handler)

	panic(app.Listen(":8080"))
}
