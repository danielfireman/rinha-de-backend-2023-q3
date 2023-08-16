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
	p, err := bp.db.Search(termo)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, p)
	case ErrNotFound:
		return echo.ErrNotFound
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
