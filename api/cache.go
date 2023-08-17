package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/danielfireman/rinha-de-backend-2023-q3/cache/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Cache struct {
	client pb.CacheClient
}

func MustNewCache() *Cache {
	conn, err := grpc.Dial("host.docker.internal:1313", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("erro ao conectar com servidor de cache: %v", err)
	}
	return &Cache{
		client: pb.NewCacheClient(conn),
	}
}

func (c *Cache) Create(p *Pessoa) error {
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
	default:
		return nil
	}
}

func (c *Cache) Get(id string) (*Pessoa, error) {
	resp, err := c.client.Get(context.TODO(), &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("error cache get: %w", err)
	}
	p := Pessoa{
		ID:         resp.Pessoa.Id,
		Apelido:    &resp.Pessoa.Apelido,
		Nome:       &resp.Pessoa.Nome,
		Nascimento: &resp.Pessoa.Nascimento,
		Stack:      resp.Pessoa.Stack,
	}
	switch resp.Status {
	case pb.Status_NOT_FOUND:
		return nil, ErrNotFound
	case pb.Status_ERROR:
		return nil, fmt.Errorf("status error in cache get: %s", resp.Msg)
	case pb.Status_DUPLICATE_KEY:
		return nil, ErrDuplicateKey
	default:
		return &p, nil
	}
}

func (c *Cache) Search(term string) ([]*Pessoa, error) {
	resp, err := c.client.Search(context.TODO(), &pb.SearchRequest{
		Term: term,
	})
	if err != nil {
		return nil, fmt.Errorf("error cache search: %w", err)
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
	switch resp.Status {
	case pb.Status_NOT_FOUND:
		return nil, ErrNotFound
	case pb.Status_ERROR:
		return nil, fmt.Errorf("status error in cache get: %s", resp.Msg)
	default:
		return pessoas, nil
	}
}
