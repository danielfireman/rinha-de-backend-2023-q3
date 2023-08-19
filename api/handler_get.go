package main

import (
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/labstack/echo/v4"
)

type getPessoa struct {
	rinhadb *RinhaDB
	cache   *ristretto.Cache
}

func (gp getPessoa) handler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.ErrBadRequest
	}

	// verifica primeiro o cache.
	if p, ok := gp.cache.Get(id); ok {
		return c.JSON(http.StatusOK, p)
	}

	// caso não encontre, busca no rinhadb.
	p, err := gp.rinhadb.Get(id)
	switch err {
	case nil:

		// caso encontre, atualiza o cache.
		gp.cache.Set(id, p, 1)
		gp.cache.Set(p.Apelido, p, 1)
		return c.JSON(http.StatusOK, p)

	case ErrNotFound:
		return echo.ErrNotFound
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
