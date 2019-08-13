package main

import (
	"context"
	"fmt"
	pb "grpc_with_tls/proto"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	Port = "1202"
)

type SearchServer struct{}

func (SearchServer) Search(ctx context.Context, req *pb.SearchRequest) (resp *pb.SearchResponse, err error) {
	// https://github.com/avinassh/grpc-errors/tree/master/go
	if req.Page > 10 {
		return nil, status.Errorf(codes.NotFound, "page not found")
	}
	return &pb.SearchResponse{
		Code: 0,
		Data: []*pb.SearchResponse_SearchResultItem{
			&pb.SearchResponse_SearchResultItem{
				Id:     11,
				Title:  "A cat is sleeping",
				Detail: "A cat is sleeping in the room",
			},
		},
	}, nil
}

//openssl req -new -newkey rsa:4096 -x509 -sha256 -days 365 -nodes -out server.pem -keyout server.key
func main() {
	cred, err := credentials.NewServerTLSFromFile("../certs/server.pem", "../certs/server.key")
	if err != nil {
		fmt.Println("can't load credentials", err)
		return
	}

	server := grpc.NewServer(grpc.Creds(cred))
	pb.RegisterSearchServiceServer(server, SearchServer{})

	lis, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		fmt.Println("listen error", err)
		return
	}

	server.Serve(lis)
}
