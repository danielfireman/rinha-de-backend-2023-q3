package main

import (
	"fmt"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type DB interface {
	Create(*Pessoa) error
	Get(string) (*Pessoa, error)
	Search(string) ([]*Pessoa, error)
	Count() (int, error)
}
