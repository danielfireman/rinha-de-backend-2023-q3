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
	db DB
}

func (cp criarPessoa) handler(c echo.Context) error {
	p := new(Pessoa)
	if err := c.Bind(p); err != nil {
		return echo.ErrBadRequest
	}
	if p.Nascimento == nil || p.Apelido == nil || p.Nome == nil {
		return echo.ErrUnprocessableEntity
	}
	p.ID = uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).
	if err := cp.db.Create(p); err != nil {
		if err == ErrDuplicateKey {
			return echo.ErrUnprocessableEntity
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error creating person: %w", err))
	}
	return c.JSON(http.StatusCreated, p)
}
