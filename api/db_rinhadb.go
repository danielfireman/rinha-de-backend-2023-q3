package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/danielfireman/rinha-de-backend-2023-q3/rinhadb/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RinhaDB struct {
	client pb.CacheClient
}

func MustNewRinhaDB() *RinhaDB {
	conn, err := grpc.Dial("host.docker.internal:1313", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("erro ao conectar com servidor de cache: %v", err)
	}
	return &RinhaDB{
		client: pb.NewCacheClient(conn),
	}
}

func (c *RinhaDB) Create(p *Pessoa) error {
	resp, err := c.client.Put(context.TODO(), &pb.PutRequest{Pessoa: &pb.Pessoa{
		Id:      p.ID,
		Apelido: *p.Apelido,
		Nome:    *p.Nome,
		Stack:   p.Stack,
	}})
	if err != nil {
		return fmt.Errorf("error cache put: %w", err)
	}
	switch resp.Status {
	case pb.Status_NOT_FOUND:
		return ErrNotFound
	case pb.Status_ERROR:
		return fmt.Errorf("status error in cache put: %s", resp.Msg)
	case pb.Status_DUPLICATE_KEY:
		return ErrDuplicateKey
	default:
		return nil
	}
}

func (c *RinhaDB) Get(id string) (*Pessoa, error) {
	resp, err := c.client.Get(context.TODO(), &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("error cache get: %w", err)
	}
	switch resp.Status {
	case pb.Status_NOT_FOUND:
		return nil, ErrNotFound
	case pb.Status_ERROR:
		return nil, fmt.Errorf("status error in cache get: %s", resp.Msg)
	}
	p := Pessoa{
		ID:         resp.Pessoa.Id,
		Apelido:    &resp.Pessoa.Apelido,
		Nome:       &resp.Pessoa.Nome,
		Nascimento: &resp.Pessoa.Nascimento,
		Stack:      resp.Pessoa.Stack,
	}
	return &p, nil
}

func (c *RinhaDB) Search(term string) ([]*Pessoa, error) {
	resp, err := c.client.Search(context.TODO(), &pb.SearchRequest{
		Term: term,
	})
	// casos de erro.
	if err != nil {
		return nil, fmt.Errorf("error rinhadb search: %w", err)
	}
	if resp.Status == pb.Status_ERROR {
		return nil, fmt.Errorf("status error in rinhadb get: %s", resp.Msg)
	}
	// caso especial para quando não encontrar o termo.
	if len(resp.Pessoas) == 0 {
		return []*Pessoa{}, nil
	}
	var pessoas []*Pessoa
	for _, p := range resp.Pessoas {
		pessoas = append(pessoas, &Pessoa{
			ID:         p.Id,
			Apelido:    &p.Apelido,
			Nome:       &p.Nome,
			Nascimento: &p.Nascimento,
			Stack:      p.Stack,
		})
	}
	return pessoas, nil
}
