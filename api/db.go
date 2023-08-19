package main

import (
	"fmt"
)

var (
	ErrNotFound     = fmt.Errorf("pessoa não encontrada")
	ErrDuplicateKey = fmt.Errorf("apelido duplicado")
)

type DB interface {
	Create(*Pessoa) error
	Get(string) (*Pessoa, error)
	Search(string) ([]*Pessoa, error)
	Count() (int, error)
}
