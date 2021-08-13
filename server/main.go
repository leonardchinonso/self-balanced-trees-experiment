package main

import (
	model "es/model"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	address = "localhost:8000"
)

func main() {
	flag.StringVar(&address, "a", address, "gRPC server address host:port")
	flag.Parse()

	var opts []grpc.ServerOption

	// TODO: Configure TLS here...

	server := grpc.NewServer(opts ...)
	model.RegisterErigonServiceServer(server, &erigonService{})

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to start gRPC server on address %v: %v", address, err))
	}

	if err := server.Serve(lis); err != nil {
		log.Fatal(fmt.Errorf("unable to start gRPC server on address %v: %v", address, err))
	}
}
