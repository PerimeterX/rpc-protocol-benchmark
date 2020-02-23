package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"net"
	"sync"
)

type grpcStreamTester struct {
	s         *grpc.Server
	conn      *grpc.ClientConn
	c         GrpcStreamServiceClient
	lock      sync.Mutex
	callbacks map[string]func(*StreamGreetResponse)
	stream    GrpcStreamService_GreetClient
}

func (g *grpcStreamTester) initServer() {
	g.callbacks = make(map[string]func(*StreamGreetResponse))
	lis, e := net.Listen("tcp", ":8080")
	if e != nil {
		panic(fmt.Sprintf("could not create grpc stream listener: %v", e.Error()))
	}
	g.s = grpc.NewServer()
	service := &grpcStreamService{}
	RegisterGrpcStreamServiceServer(g.s, service)
	go func() {
		if e := g.s.Serve(lis); e != nil {
			panic(fmt.Sprintf("could not start grpc stream server: %v", e.Error()))
		}
	}()
}

func (g *grpcStreamTester) initClient() {
	var e error
	if g.conn, e = grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithBlock()); e != nil {
		panic(fmt.Sprintf("could not start grpc stream client: %v", e.Error()))
	}
	g.c = NewGrpcStreamServiceClient(g.conn)
	if g.stream, e = g.c.Greet(context.Background()); e != nil {
		panic(fmt.Sprintf("count not create grpc client stream: %v", e.Error()))
	}
	go func() {
		for {
			in, e := g.stream.Recv()
			if e != nil {
				return
			}
			g.lock.Lock()
			callback, exists := g.callbacks[in.Id]
			g.lock.Unlock()
			if !exists {
				panic(fmt.Sprintf("grpc stream client callback %s does not exist", in.Id))
			}
			callback(in)
		}
	}()
}

func (g *grpcStreamTester) doRPC(name string) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	g.lock.Lock()
	g.callbacks[name] = func(response *StreamGreetResponse) {
		if response.Greeting != fmt.Sprintf("hello, %s", name) {
			panic(fmt.Sprintf("wrong grpc stream answer: %s", response.Greeting))
		}
		g.lock.Lock()
		defer g.lock.Unlock()
		delete(g.callbacks, name)
		wg.Done()
	}
	g.lock.Unlock()
	if e := g.stream.Send(&StreamGreetRequest{Id: name, Name: name}); e != nil {
		panic(fmt.Sprintf("grpc error: %s", e.Error()))
	}
	wg.Wait()
}

func (g *grpcStreamTester) closeClient() {
	if e := g.stream.CloseSend(); e != nil {
		panic(fmt.Sprintf("could not close grpc stream: %v", e.Error()))
	}
	if e := g.conn.Close(); e != nil {
		panic(fmt.Sprintf("could not close grpc stream client: %v", e.Error()))
	}
}

func (g *grpcStreamTester) closeServer() {
	g.s.Stop()
}

type grpcStreamService struct{}

func (g *grpcStreamService) Greet(stream GrpcStreamService_GreetServer) error {
	for {
		in, e := stream.Recv()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		if e := stream.Send(&StreamGreetResponse{Id: in.Id, Greeting: fmt.Sprintf("hello, %s", in.Name)}); e != nil {
			panic(fmt.Sprintf("grpc stream send response error: %s", e.Error()))
		}
	}
}
