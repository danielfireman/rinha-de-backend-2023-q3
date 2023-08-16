package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type contarPessoas struct {
	db DB
}

func (cp contarPessoas) handler(c echo.Context) error {
	n, err := cp.db.Count()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, fmt.Sprintf("%d", n))
}
