package main

import (
	"net/http"

	"github.com/alphadose/haxmap"
	"github.com/labstack/echo/v4"
	"github.com/rogpeppe/fastuuid"
)

type getPessoa struct {
	rinhadb      *RinhaDB
	cache        *haxmap.Map[string, string]
	apelidoCache *haxmap.Map[string, struct{}]
}

func (gp getPessoa) handler(c echo.Context) error {
	id := c.Param("id")
	if id == "" || !fastuuid.ValidHex128(id) {
		return echo.ErrBadRequest
	}

	// verifica primeiro o cache.
	if p, ok := gp.cache.Get(id); ok {
		return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, []byte(p))
	}

	// caso não encontre no cache, busca no rinhadb.
	pessoaStr, apelido, err := gp.rinhadb.Get(id)
	if err != nil {
		if err == ErrNotFound {
			return echo.ErrNotFound
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else {
		// caso encontre, atualiza o cache.
		gp.cache.Set(id, pessoaStr)
		gp.apelidoCache.Set(apelido, struct{}{})
		return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, []byte(pessoaStr))
	}
}
