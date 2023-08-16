package main

import (
	"strings"

	"golang.org/x/exp/slices"
)

type MemDB struct {
	data map[string]*Pessoa
}

func MustNewMemDB() *MemDB {
	return &MemDB{
		data: make(map[string]*Pessoa),
	}
}

func (db *MemDB) Create(p *Pessoa) error {
	db.data[p.ID] = p
	return nil
}

func (db *MemDB) Get(id string) (*Pessoa, error) {
	p, ok := db.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// NOTE: Essa função não está implementada corretamente. Segundo a documentação ela deve ignorar o case e fazer match parcial.
func (db *MemDB) Search(termo string) ([]*Pessoa, error) {
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

func (db *MemDB) Count() (int, error) {
	return len(db.data), nil
}
