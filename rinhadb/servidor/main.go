package main

import (
	"log"
	"net"

	pb "github.com/danielfireman/rinha-de-backend-2023-q3/rinhadb/proto"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":1313")
	if err != nil {
		log.Fatalf("falha ao ouvir porta 1313: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRinhaDBServer(s, newServer())
	log.Printf("servidor ouvindo %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("falha ao servir: %v", err)
	}
}
