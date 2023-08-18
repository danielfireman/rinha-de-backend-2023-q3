package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type buscaPessoas struct {
	rinhadb *RinhaDB
}

func (bp buscaPessoas) handler(c echo.Context) error {
	termo := c.QueryParam("t")
	if termo == "" {
		return echo.ErrBadRequest
	}
	p, err := bp.rinhadb.Search(termo)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, p)
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
