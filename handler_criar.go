package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rogpeppe/fastuuid"
)

const dateLayout = "2006-01-02"

var uuidGen = fastuuid.MustNewGenerator()

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
	p.ID = uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).
	if err := cp.db.Create(p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
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
