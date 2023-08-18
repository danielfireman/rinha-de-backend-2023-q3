package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rogpeppe/fastuuid"
)

var (
	uuidGen = fastuuid.MustNewGenerator()
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
			if err := rinhaDB.Create(p); err != nil {
				log.Printf("error creating person %s at rinhadb: %w", p.ID, err)
			}
			if err := mongoDB.Create(p); err != nil {
				log.Printf("error creating person %s at mongodb: %w", p.ID, err)
			}
		}
	}()
	return &criarPessoa{
		db:          mongoDB,
		rinhadb:     rinhaDB,
		chanCriacao: c,
	}
}

func (cp criarPessoa) handler(c echo.Context) error {
	p := new(Pessoa)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("error binding pessoa: %w", err))
	}
	// validação
	if p.Nome == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo nome não preenchido"))
	}
	if p.Apelido == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo apelido não preenchido"))
	}
	if p.Nascimento == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo nascimento não preenchido"))
	}

	// checando duplicidade de apelido.
	if isDup, err := cp.rinhadb.ChecaDuplicata(*p.Apelido); err == nil {
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error checking duplicate: %w", err))
		}
		if isDup {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error apelido duplicado"))
		}
	}
	// preenchimento do ID
	p.ID = uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).

	// envia evento para o worker que atualiza os bancos de dados.
	// somente executar quando a requisição for válida.
	go func() {
		cp.chanCriacao <- p
	}()
	return c.JSON(http.StatusCreated, p)
}
