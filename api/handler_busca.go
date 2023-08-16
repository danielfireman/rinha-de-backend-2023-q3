package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type buscaPessoas struct {
	db DB
}

func (bp buscaPessoas) handler(c echo.Context) error {
	termo := c.QueryParam("t")
	if termo == "" {
		return echo.ErrBadRequest
	}
	pessoas, err := bp.db.Search(termo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pessoas)
}
