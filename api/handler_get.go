package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type getPessoa struct {
	db DB
}

func (gp getPessoa) handler(c echo.Context) error {
	p, err := gp.db.Get(c.Param("id"))
	switch err {
	case nil:
		return c.JSON(http.StatusOK, p)
	case ErrNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
