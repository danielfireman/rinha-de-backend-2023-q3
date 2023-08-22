package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type criarPessoa struct {
	db      DB
	rinhadb *RinhaDB

	// canal para enviar chamadas remotas para criação de pessoa de forma assíncrona.
	chanCriacao chan *Pessoa
}

func newCriarPessoa(mongoDB DB, rinhaDB *RinhaDB) *criarPessoa {
	c := make(chan *Pessoa)
	// inicializa worker que atualiza o mongodb de forma assíncrona.
	go func() {
		for {
			p := <-c
			if err := mongoDB.Create(p); err != nil {
				panic(fmt.Errorf("error creating person %s at mongodb: %w", p.ID, err))
			}
		}
	}()
	return &criarPessoa{
		db:          mongoDB,
		rinhadb:     rinhaDB,
		chanCriacao: c,
	}
}

func (cp criarPessoa) handler(c *fiber.Ctx) error {
	p := new(Pessoa)
	if err := c.BodyParser(p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("error binding pessoa: %s", err))
	}
	// validação
	if p.Nome == nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "error campo nome não preenchido")
	}
	if p.Apelido == nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "error campo apelido não preenchido")
	}
	if p.Nascimento == nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "error campo nascimento não preenchido")
	}
	// cria pessoa no rinhadb.
	var pID string
	var err error
	if pID, err = cp.rinhadb.Create(p); err != nil {
		if err == ErrDuplicateKey {
			return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintf("error apelido duplicado %s", *p.Apelido))
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error criando pessoa: %s", err))
	}
	// envia evento para o worker que atualiza os bancos de dados.
	// somente executar quando a requisição for válida.
	go func() {
		cp.chanCriacao <- p
	}()
	c.Set("Location", fmt.Sprintf("/pessoas/%s", pID))
	return c.SendStatus(fiber.StatusCreated)
}
