package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alphadose/haxmap"
	pb "github.com/danielfireman/rinha-de-backend-2023-q3/rinhadb/proto"
	"github.com/rogpeppe/fastuuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type novaPessoa struct {
	pessoa  *pb.Pessoa
	confirm chan struct{}
}

type RinhaDB struct {
	client         pb.RinhaDBClient
	cache          *haxmap.Map[string, string]
	apelidoCache   *haxmap.Map[string, struct{}]
	uuidGen        *fastuuid.Generator
	novaPessoaChan chan *novaPessoa
}

func MustNewRinhaDB() *RinhaDB {
	conn, err := grpc.Dial("host.docker.internal:1313",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("erro ao conectar com servidor de cache: %v", err)
	}
	client := pb.NewRinhaDBClient(conn)
	newPersonStream, err := client.NewPerson(context.TODO())
	if err != nil {
		log.Fatalf("erro ao criar o stream com servidor de cache: %v", err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("erro ao obter o hostname: %v", err)
	}
	if err := newPersonStream.SendMsg(&pb.StreamID{Id: hostname}); err != nil {
		log.Fatalf("erro ao enviar mensagem inicial no stream de novas pessoas: %v", err)
	}

	// criando estruturas de dados para serem cache.
	// TODO: usar algo mais rubusto, por exemplo couchdb's ristretto.
	apelidoCache := haxmap.New[string, struct{}](1e5)
	cache := haxmap.New[string, string](1e5)
	novaPessoaChan := make(chan *novaPessoa)
	novaPessoaMap := haxmap.New[string, *novaPessoa](1e5)

	// disparando worker que fica escutando o channel de novas pessoas e enviando
	// as atualizações para o servidor. Isso é importante pois não devemos ter
	// mensagens sendo enviadas por diversas goroutines num mesmo stream.
	go func() {
		for {
			novaPessoa := <-novaPessoaChan
			novaPessoaMap.Set(novaPessoa.pessoa.Id, novaPessoa)
			if err := newPersonStream.Send(novaPessoa.pessoa); err != nil {
				log.Fatalf("erro avisando servidor sobre nova pessoa: %s", err)
			}
		}
	}()

	// disparando worker que fica escutando o stream de novas pessoas
	// e atualizando os caches. Isso é importante pois não devemos ter
	// mensagens sendo recebidas por diversas goroutines num mesmo stream.
	go func() {
		for {
			pessoa, err := newPersonStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("erro ao receber mensagem do stream de novas pessoas: %v", err)
			}

			// Atualiza os caches quando uma nova pessoa chega.
			cache.Set(pessoa.Id, pessoa2Str(pessoa))
			apelidoCache.Set(pessoa.Apelido, struct{}{})

			// se a pessoa foi criada nessa máquina, libera a chamada
			if np, ok := novaPessoaMap.Get(pessoa.Id); ok {
				np.confirm <- struct{}{}
			}
		}
	}()
	return &RinhaDB{
		client:         client,
		cache:          cache,
		apelidoCache:   apelidoCache,
		uuidGen:        fastuuid.MustNewGenerator(),
		novaPessoaChan: novaPessoaChan,
	}
}

func (c *RinhaDB) Create(p *Pessoa) (string, error) {
	// primeiro checa apelido no cache, evitando um RT no rinhadb para verificar.
	if _, ok := c.apelidoCache.Get(*p.Apelido); ok {
		return "", ErrDuplicateKey
	}
	// cria id.
	p.ID = c.uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).
	pessoa := &pb.Pessoa{
		Id:         p.ID,
		Apelido:    *p.Apelido,
		Nome:       *p.Nome,
		Nascimento: *p.Nascimento,
		Stack:      p.Stack,
	}
	np := &novaPessoa{
		pessoa:  pessoa,
		confirm: make(chan struct{}),
	}
	c.novaPessoaChan <- np

	// aguarda a confirmação que todas as máquinas receberam a chamada
	<-np.confirm
	return p.ID, nil
}

func (c *RinhaDB) Get(id string) (string, error) {
	// verifica cache primeiro.
	p, ok := c.cache.Get(id)
	if ok {
		return p, nil
	}
	// caso não esteja no cache, faz a chamada remota.
	resp, err := c.client.Get(context.TODO(), &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return "", fmt.Errorf("error cache get: %w", err)
	}
	switch resp.Status {
	case pb.Status_NOT_FOUND:
		return "", ErrNotFound
	case pb.Status_ERROR:
		return "", fmt.Errorf("status error in cache get: %s", resp.Msg)
	}
	return resp.Pessoa, nil
}

func (c *RinhaDB) Search(term string) (string, error) {
	stream, err := c.client.Search(context.TODO(), &pb.SearchRequest{
		Term: term,
	})
	if err != nil {
		return "", fmt.Errorf("error rinhadb search: %w", err)
	}
	pessoas := []string{}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error on searching term %s: %w", term, err)
		}
		pessoas = append(pessoas, resp.Pessoa)
	}
	return fmt.Sprintf("[%s]", strings.Join(pessoas, ",")), nil
}

func pessoa2Str(p *pb.Pessoa) string {
	return fmt.Sprintf(`{"id": "%s", "apelido": "%s", "nome": "%s", "nascimento": "%s", "stack": %s}`,
		p.Id,
		p.Apelido,
		p.Nome,
		p.Nascimento,
		`["`+strings.Join(p.Stack, `", "`)+`"]`)
}
