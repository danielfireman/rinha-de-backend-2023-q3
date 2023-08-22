package main

import (
	"time"

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
	rinhaDB := MustNewRinhaDB()

	app.Post("/pessoas", newCriarPessoa(db, rinhaDB).handler)
	app.Get("/pessoas/:id", getPessoa{rinhaDB}.handler)
	app.Get("/pessoas", buscaPessoas{rinhaDB}.handler)
	app.Get("/contagem-pessoas", contarPessoas{db}.handler)

	panic(app.Listen(":8080"))
}
