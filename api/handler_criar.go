package main

import (
	"fmt"
	"net/http"

	"github.com/alphadose/haxmap"
	"github.com/labstack/echo/v4"
)

type criarPessoa struct {
	db           DB
	rinhadb      *RinhaDB
	cache        *haxmap.Map[string, string]
	apelidoCache *haxmap.Map[string, struct{}]

	// canal para enviar chamadas remotas para criação de pessoa de forma assíncrona.
	chanCriacao chan *Pessoa
}

func newCriarPessoa(mongoDB DB, rinhaDB *RinhaDB, cache *haxmap.Map[string, string], apelidoCache *haxmap.Map[string, struct{}]) *criarPessoa {
	c := make(chan *Pessoa)
	// inicializa worker que atualiza o mongodb de forma assíncrona.
	go func() {
		for {
			p := <-c
			if err := mongoDB.Create(p); err != nil {
				panic(fmt.Errorf("error creating person %s at mongodb: %w", p.ID, err))
			}
		}
	}()
	return &criarPessoa{
		db:           mongoDB,
		rinhadb:      rinhaDB,
		cache:        cache,
		apelidoCache: apelidoCache,
		chanCriacao:  c,
	}
}

func (cp criarPessoa) handler(c echo.Context) error {
	p := new(Pessoa)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("error binding pessoa: %w", err))
	}
	// validação
	if p.Nome == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo nome não preenchido"))
	}
	if p.Apelido == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo apelido não preenchido"))
	}
	if p.Nascimento == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error campo nascimento não preenchido"))
	}

	// primeiro checa apelido no cache, evitando um RT no rinhadb para verificar.
	if _, ok := cp.apelidoCache.Get(*p.Apelido); ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error apelido duplicado %s", *p.Apelido))
	}

	// cria pessoa no rinhadb.
	if pID, pStr, err := cp.rinhadb.Create(p); err == nil {
		// adiciona pessoa no cache.
		cp.cache.Set(p.ID, pStr)
		cp.apelidoCache.Set(*p.Apelido, struct{}{})

		// envia evento para o worker que atualiza os bancos de dados.
		// somente executar quando a requisição for válida.
		go func() {
			cp.chanCriacao <- p
		}()

		c.Response().Header().Set("Location", fmt.Sprintf("/pessoas/%s", pID))
		return c.NoContent(http.StatusCreated)

	} else {
		if err == ErrDuplicateKey {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("error apelido duplicado %s", *p.Apelido))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error criando pessoa: %w", err))
	}
}
