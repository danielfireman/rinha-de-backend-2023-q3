package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type getPessoa struct {
	db DB
}

func (gp getPessoa) handler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.ErrBadRequest
	}
	p, err := gp.db.Get(id)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, p)
	case ErrNotFound:
		return echo.ErrNotFound
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
