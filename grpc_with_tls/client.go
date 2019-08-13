package main

import (
	"context"
	"fmt"
	pb "grpc_with_tls/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	Address = "localhost:1202"
)

func main() {
	// 如果服务端的证书是由其它 CA 签发，只要导入 root.pem 就行
	// 因此要连接多个微服务，如果他们的证书由同一个 CA 签发，客户端只要一个证书就行
	cred, err := credentials.NewClientTLSFromFile("./certs/server.pem", "")
	if err != nil {
		fmt.Println("load credential error", err)
		return
	}
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(cred))
	if err != nil {
		fmt.Println("dial error", err)
		return
	}
	client := pb.NewSearchServiceClient(conn)

	resp, err := client.Search(context.Background(), &pb.SearchRequest{
		Query: "cat",
		Page:  1,
		Count: 10,
	})

	if err != nil {
		fmt.Println("search error", err)
		return
	}

	fmt.Println("result:")
	fmt.Println(resp)

	resp, err = client.Search(context.Background(), &pb.SearchRequest{
		Query: "cat",
		Page:  11,
		Count: 10,
	})

	if err != nil {
		fmt.Println("search error", err)
		return
	}
}
