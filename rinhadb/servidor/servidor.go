package main

import (
	"context"
	"strings"
	"sync"

	"github.com/alphadose/haxmap"
	pb "github.com/danielfireman/rinha-de-backend-2023-q3/rinhadb/proto"
	"github.com/rogpeppe/fastuuid"
	"golang.org/x/sync/semaphore"
)

const (
	searchLimit           = 50
	concurrencyLevel      = 2
	initialSizeMapPessoas = 1e5
	initialSizeMapItems   = 1e5
)

type server struct {
	pb.UnimplementedCacheServer

	apelidoMap    *haxmap.Map[string, struct{}]     // mapa usado para detectar apelidos duplicados.
	idMap         *haxmap.Map[string, *pb.Pessoa]   // mapa usado para o Get.
	indice        *haxmap.Map[string, []*pb.Pessoa] // indice invertido, usado para o Search.
	muCriacao     sync.Mutex                        // Lock para tornar atômica a checagem de duplicatas e adição de novas pessoas.
	chanIndexacao chan *pb.Pessoa                   // Canal para indexação de pessoas de forma assíncrona.
	uuidGen       *fastuuid.Generator

	// [PerfNote] Como temos apenas uma thread, não queremos que as diversas goroutines (uma por
	// requisição) fiquem disputando a CPU. Por isso, usamos um semáforo para garantir que apenas uma
	// esteja acordada num determinado momento. Na prática, é como se tivéssemos um worker pool.
	sem *semaphore.Weighted
}

func newServer() *server {
	s := &server{
		apelidoMap:    haxmap.New[string, struct{}](initialSizeMapPessoas),
		idMap:         haxmap.New[string, *pb.Pessoa](initialSizeMapPessoas),
		indice:        haxmap.New[string, []*pb.Pessoa](initialSizeMapItems),
		chanIndexacao: make(chan *pb.Pessoa),
		sem:           semaphore.NewWeighted(concurrencyLevel),
		uuidGen:       fastuuid.MustNewGenerator(),
	}
	// dispara worker de indexação numa nova goroutine.
	// ela será executada de forma assíncrona, não bloqueando o servidor durante
	// a criação de pessoas.
	go func() {
		s.iniciaIndexador()
	}()
	return s
}

func (s *server) Put(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	s.sem.Acquire(ctx, 1)
	defer s.sem.Release(1)

	// [ConcurrencyNote] Checar duplicata e criar uma nova entrada precisa ser executado de forma atômica.
	s.muCriacao.Lock()
	// verifica apelidos duplicados.
	_, ok := s.apelidoMap.Get(in.Pessoa.Apelido)
	if ok {
		s.muCriacao.Unlock()
		return &pb.PutResponse{
			Status: pb.Status_DUPLICATE_KEY,
		}, nil
	}
	// preenchendo mapa de pessoas, o que vai fazer o get e a checagem de duplicatas funcionarem
	// logo após o put.
	pessoa := in.Pessoa
	pessoa.Id = s.uuidGen.Hex128() // it is okay to call it concurrently (as per Next()).
	s.apelidoMap.Set(in.Pessoa.Apelido, struct{}{})
	s.idMap.Set(pessoa.Id, pessoa)
	s.muCriacao.Unlock()

	// dispara a indexação de forma assíncrona.
	go func() {
		s.chanIndexacao <- pessoa
	}()

	return &pb.PutResponse{
		Status: pb.Status_OK,
		Pessoa: pessoa,
	}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.sem.Acquire(ctx, 1)
	defer s.sem.Release(1)

	p, ok := s.idMap.Get(in.Id)
	if !ok { // quando o get não tiver resultados, deve retornar not found.
		return &pb.GetResponse{
			Status: pb.Status_NOT_FOUND,
		}, nil
	}
	return &pb.GetResponse{
		Pessoa: p,
		Status: pb.Status_OK,
	}, nil
}

func (s *server) Search(ctx context.Context, in *pb.SearchRequest) (*pb.SearchResponse, error) {
	s.sem.Acquire(ctx, 1)
	defer s.sem.Release(1)

	p, ok := s.indice.Get(strings.ToLower(in.Term))
	if !ok { // quando a busca não tiver resultados, deve retornar 200.
		return &pb.SearchResponse{
			Pessoas: []*pb.Pessoa{},
			Status:  pb.Status_OK,
		}, nil
	}
	if len(p) > searchLimit {
		p = p[:searchLimit]
	}
	return &pb.SearchResponse{
		Pessoas: p,
		Status:  pb.Status_OK,
	}, nil
}

func (s *server) iniciaIndexador() {
	for {
		p := <-s.chanIndexacao
		// preenchendo índice invertido.
		// coletando lista de termos.
		termos := strings.Split(strings.ToLower(p.Nome), " ")
		termos = append(termos, strings.ToLower(p.Apelido))
		for _, s := range p.Stack {
			termos = append(termos, strings.ToLower(s))
		}

		// associando termos a pessoa.
		for _, t := range termos {
			v, _ := s.indice.Get(t)
			s.indice.Set(t, append(v, p))
		}
	}
}
