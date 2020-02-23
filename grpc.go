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

func (g *grpcTester) initServer() {
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
}

func (g *grpcTester) initClient() {
	var e error
	g.conn, e = grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithBlock())
	if e != nil {
		panic(fmt.Sprintf("could not start grpc client: %v", e.Error()))
	}
	g.c = NewGrpcServiceClient(g.conn)
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

func (g *grpcTester) closeClient() {
	if e := g.conn.Close(); e != nil {
		panic(fmt.Sprintf("could not close grpc client: %v", e.Error()))
	}
}

func (g *grpcTester) closeServer() {
	g.s.Stop()
}

type grpcService struct{}

func (g *grpcService) Greet(_ context.Context, req *GreetRequest) (*GreetResponse, error) {
	return &GreetResponse{Greeting: fmt.Sprintf("hello, %s", req.Name)}, nil
}
