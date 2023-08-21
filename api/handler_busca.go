package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type buscaPessoas struct {
	rinhadb *RinhaDB
}

func (bp buscaPessoas) handler(c *fiber.Ctx) error {
	termo := c.Query("t")
	if termo == "" {
		return fiber.ErrBadRequest
	}
	p, err := bp.rinhadb.Search(termo)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Status(http.StatusOK)
	c.SendString(p)
	return nil
}
