package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rogpeppe/fastuuid"
)

type getPessoa struct {
	rinhadb *RinhaDB
}

func (gp getPessoa) handler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" || !fastuuid.ValidHex128(id) {
		return fiber.ErrBadRequest
	}
	pessoaStr, err := gp.rinhadb.Get(id)
	if err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "pessoa não encontrada: %s", id)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Status(http.StatusOK).
		SendString(pessoaStr)
}
