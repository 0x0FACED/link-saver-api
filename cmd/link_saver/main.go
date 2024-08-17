package main

import (
	"log"
	"net"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/server"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	cfg, err := config.Load()
	if err != nil {
		panic("cant load config, panic!")
	}
	srv := server.New(cfg)
	gen.RegisterLinkServiceServer(s, srv)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
