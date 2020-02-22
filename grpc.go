package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type grpcTester struct {
	s    *grpc.Server
	conn *grpc.ClientConn
	c    GrpcServiceClient
}

func (g *grpcTester) init() {
	lis, e := net.Listen("tcp", ":8080")
	if e != nil {
		panic(fmt.Sprintf("could not create grpc listener: %v", e.Error()))
	}
	g.s = grpc.NewServer()
	service := &grpcService{}
	RegisterGrpcServiceServer(g.s, service)
	go func() {
		if e := g.s.Serve(lis); e != nil {
			panic(fmt.Sprintf("could not start grpc server: %v", e.Error()))
		}
	}()
	g.conn, e = grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithBlock())
	if e != nil {
		panic(fmt.Sprintf("could not start grpc client: %v", e.Error()))
	}
	g.c = NewGrpcServiceClient(g.conn)
}

func (g *grpcTester) close() {
	if e := g.conn.Close(); e != nil {
		panic(fmt.Sprintf("could not close grpc client: %v", e.Error()))
	}
	g.s.Stop()
}

func (g *grpcTester) doRPC(name string) {
	r, e := g.c.Greet(context.Background(), &GreetRequest{Name: name})
	if e != nil {
		panic(fmt.Sprintf("grpc error: %s", e.Error()))
	}
	if r.Greeting != fmt.Sprintf("hello, %s", name) {
		panic(fmt.Sprintf("wrong grpc answer: %s", r.Greeting))
	}
}

type grpcService struct{}

func (g *grpcService) Greet(_ context.Context, req *GreetRequest) (*GreetResponse, error) {
	return &GreetResponse{Greeting: fmt.Sprintf("hello, %s", req.Name)}, nil
}
