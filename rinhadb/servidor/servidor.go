package main

import (
	"context"
	"strings"

	pb "github.com/danielfireman/rinha-de-backend-2023-q3/rinhadb/proto"
)

const (
	searchLimit = 50
)

type server struct {
	pb.UnimplementedCacheServer

	// [PerfNote] Não estou usando locks para acessar esses mapas pois o servidor está
	// configurado para rodar com 1 core.
	apelidoMap map[string]struct{}     // mapa comum, usado para detectar apelidos duplicados.
	idMap      map[string]*pb.Pessoa   // mapa comum, usado para o Get.
	indice     map[string][]*pb.Pessoa // indice invertido, usado para o Search.
}

func newServer() *server {
	return &server{
		apelidoMap: make(map[string]struct{}),
		idMap:      make(map[string]*pb.Pessoa),
		indice:     make(map[string][]*pb.Pessoa),
	}
}

func (s *server) Put(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	_, ok := s.apelidoMap[in.Pessoa.Apelido]
	if ok {
		return &pb.PutResponse{
			Status: pb.Status_DUPLICATE_KEY,
		}, nil
	}

	// preenchendo mapa.
	pessoa := in.Pessoa
	s.apelidoMap[in.Pessoa.Apelido] = struct{}{}
	s.idMap[pessoa.Id] = pessoa

	// preenchendo índice invertido.
	// coletando lista de termos.
	termos := strings.Split(strings.ToLower(pessoa.Nome), " ")
	termos = append(termos, strings.ToLower(pessoa.Apelido))
	for _, s := range pessoa.Stack {
		termos = append(termos, strings.ToLower(s))
	}

	// associando termos a pessoa.
	for _, t := range termos {
		s.indice[t] = append(s.indice[t], pessoa)
	}
	return &pb.PutResponse{
		Status: pb.Status_OK,
		Pessoa: pessoa,
	}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	p, ok := s.idMap[in.Id]
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
	p, ok := s.indice[strings.ToLower(in.Term)]
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