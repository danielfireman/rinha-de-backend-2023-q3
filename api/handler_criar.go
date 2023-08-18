package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rogpeppe/fastuuid"
)

var (
	uuidGen = fastuuid.MustNewGenerator()
)

type criarPessoa struct {
	db              DB
	rinhadb         *RinhaDB
	chanMongoUpdate chan *Pessoa
}

func newCriarPessoa(db DB, rinhaDB *RinhaDB) *criarPessoa {
	c := make(chan *Pessoa)
	// inicializa worker que atualiza o mongodb de forma assíncrona.
	go func() {
		for {
			p := <-c
			if err := db.Create(p); err != nil {
				panic(err)
			}
		}
	}()
	return &criarPessoa{
		db:              db,
		rinhadb:         rinhaDB,
		chanMongoUpdate: c,
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
	// preenchimento do ID
	p.ID = uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).

	// atualizando o rinhadb de forma síncrona.
	if err := cp.rinhadb.Create(p); err != nil {
		if err == ErrDuplicateKey {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error apelido duplicado"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error creating person: %w", err))
	}

	// envia evento para o worker que atualiza o mongo.
	// somente executar quando a requisição for válida.
	go func() {
		cp.chanMongoUpdate <- p
	}()
	return c.JSON(http.StatusCreated, p)
}
