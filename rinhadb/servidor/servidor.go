package main

import (
	"context"
	"fmt"
	"io"
	"strings"

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
	pb.UnimplementedRinhaDBServer

	idMap         *haxmap.Map[string, string]   // mapa usado para o Get.
	indice        *haxmap.Map[string, []string] // indice invertido, usado para o Search.
	chanIndexacao chan *pb.Pessoa               // Canal para indexação de pessoas de forma assíncrona.
	uuidGen       *fastuuid.Generator
	clientStreams map[string]pb.RinhaDB_NewPersonServer

	// [PerfNote] Como temos apenas uma thread, não queremos que as diversas goroutines (uma por
	// requisição) fiquem disputando a CPU. Por isso, usamos um semáforo para garantir que apenas uma
	// esteja acordada num determinado momento. Na prática, é como se tivéssemos um worker pool.
	sem *semaphore.Weighted
}

func newServer() *server {
	s := &server{
		idMap:         haxmap.New[string, string](initialSizeMapPessoas),
		indice:        haxmap.New[string, []string](initialSizeMapItems),
		clientStreams: make(map[string]pb.RinhaDB_NewPersonServer),
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

func (s *server) NewPerson(stream pb.RinhaDB_NewPersonServer) error {
	var sID pb.StreamID
	if err := stream.RecvMsg(&sID); err != nil { // ignorando mensagem inicial
		return fmt.Errorf("err ao receber primeira mensagem no stream: %w", err)
	}
	id := sID.Id
	s.clientStreams[id] = stream
	for {
		pessoa, err := stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("o fluxo de novas pessoas não deve ser fechado")
		}
		if err != nil {
			return fmt.Errorf("erro ao receber mensagem do stream de novas pessoas: %w", err)
		}

		// atualiza bancos de dados com valor pré-processado (pronto para ser retornado).
		s.idMap.Set(pessoa.Id, pessoa2Str(pessoa))

		// enviando a atualização para as demais instâncias.
		for k, stream := range s.clientStreams {
			if k != id {
				if err := stream.Send(pessoa); err != nil {
					return fmt.Errorf("erro ao enviar pessoa no stream de novas pessoas: %w", err)
				}
			}
		}

		if err := stream.Send(pessoa); err != nil {
			return fmt.Errorf("erro ao confirmar envio de pessoa: %w", err)
		}

		// dispara a indexação de forma assíncrona.
		go func() {
			s.chanIndexacao <- pessoa
		}()
	}
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

func (s *server) Search(req *pb.SearchRequest, stream pb.RinhaDB_SearchServer) error { //Search(ctx context.Context, in *pb.SearchRequest) (*pb.SearchResponse, error) {
	s.sem.Acquire(context.TODO(), 1)
	defer s.sem.Release(1)

	resCount := 0
	term := strings.ToLower(req.Term)
	var sendError error
	s.indice.ForEach(func(k string, pessoas []string) bool {
		if strings.Contains(k, term) {
			for _, p := range pessoas {
				if err := stream.Send(&pb.SearchResponse{Pessoa: p}); err != nil {
					sendError = err
					return false
				}
			}
		}
		resCount++
		return resCount < searchLimit
	})
	return sendError
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
			s.indice.Set(t, append(v, pessoa2Str(p)))
		}
	}
}

func pessoa2Str(p *pb.Pessoa) string {
	return fmt.Sprintf(`{"id": "%s", "apelido": "%s", "nome": "%s", "nascimento": "%s", "stack": %s}`,
		p.Id,
		p.Apelido,
		p.Nome,
		p.Nascimento,
		`["`+strings.Join(p.Stack, `", "`)+`"]`)
}
