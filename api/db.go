package main

import (
	"fmt"
)

var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrDuplicateKey = fmt.Errorf("duplicate key")
)

type DB interface {
	Create(*Pessoa) error
	Get(string) (*Pessoa, error)
	Search(string) ([]*Pessoa, error)
	Count() (int, error)
}
