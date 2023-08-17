package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alitto/pond"
	"github.com/labstack/echo/v4"
	"github.com/rogpeppe/fastuuid"
)

const dateLayout = "2006-01-02"

var (
	uuidGen = fastuuid.MustNewGenerator()
	pool    = pond.New(5, 50)
)

type criarPessoa struct {
	db DB
}

func (cp criarPessoa) handler(c echo.Context) error {
	p := new(Pessoa)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := cp.validate(p); err != nil {
		return err
	}
	pool.Submit(func() {
		p.ID = uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).
		if err := cp.db.Create(p); err != nil {
			panic(fmt.Errorf("error creating person: %w", err))
		}
	})
	return c.JSON(http.StatusCreated, p)
}

func (c criarPessoa) validate(p *Pessoa) error {
	// verifica campos obrigatórios.
	if p.Nascimento == nil || p.Apelido == nil || p.Nome == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	// verifica se a data está no formato correto.
	if _, err := time.Parse(dateLayout, *p.Nascimento); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return nil
}
