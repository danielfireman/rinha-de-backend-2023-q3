package main

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
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

type InMemDB struct {
	data map[string]*Pessoa
}

func NewInMemDB() *InMemDB {
	return &InMemDB{
		data: make(map[string]*Pessoa),
	}
}

func (db *InMemDB) Create(p *Pessoa) error {
	db.data[p.ID] = p
	return nil
}

func (db *InMemDB) Get(id string) (*Pessoa, error) {
	p, ok := db.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// NOTE: Essa função não está implementada corretamente. Segundo a documentação ela deve ignorar o case e fazer match parcial.
func (db *InMemDB) Search(termo string) ([]*Pessoa, error) {
	results := make([]*Pessoa, 0) // quando não encontrar matches deve retornar slice vazio.
	for _, p := range db.data {
		if strings.Contains(*p.Nome, termo) ||
			strings.Contains(*p.Apelido, termo) ||
			slices.Contains(p.Stack, termo) {
			results = append(results, p)
		}
	}
	return results, nil
}

func (db *InMemDB) Count() (int, error) {
	return len(db.data), nil
}
