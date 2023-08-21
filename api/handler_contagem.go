package main

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type contarPessoas struct {
	db DB
}

func (cp contarPessoas) handler(c *fiber.Ctx) error {
	n, err := cp.db.Count()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).
		SendString(fmt.Sprintf("%d", n))
}
