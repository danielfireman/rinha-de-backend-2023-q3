package main

import (
	"net/http"

	"github.com/alphadose/haxmap"
	"github.com/gofiber/fiber/v2"
	"github.com/rogpeppe/fastuuid"
)

type getPessoa struct {
	rinhadb      *RinhaDB
	cache        *haxmap.Map[string, string]
	apelidoCache *haxmap.Map[string, struct{}]
}

func (gp getPessoa) handler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" || !fastuuid.ValidHex128(id) {
		return fiber.ErrBadRequest
	}

	// verifica primeiro o cache.
	if p, ok := gp.cache.Get(id); ok {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Status(http.StatusOK).
			SendString(p)
	}

	// caso não encontre no cache, busca no rinhadb.
	pessoaStr, apelido, err := gp.rinhadb.Get(id)
	if err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "pessoa não encontrada: %s", id)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	} else {
		// caso encontre, atualiza o cache.
		gp.cache.Set(id, pessoaStr)
		gp.apelidoCache.Set(apelido, struct{}{})
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Status(http.StatusOK).
			SendString(pessoaStr)
	}
}
