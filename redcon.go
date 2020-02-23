package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/tidwall/redcon"
	"net"
	"strings"
)

type redconTester struct {
	lis    net.Listener
	client *redis.Client
}

func (r *redconTester) initServer() {
	var e error
	if r.lis, e = net.Listen("tcp", ":8080"); e != nil {
		panic(fmt.Sprintf("coult not start redcon listener: %s", e.Error()))
	}
	go func() {
		if e := redcon.Serve(r.lis, handler, accept, closed); e != nil {
			panic(fmt.Sprintf("could not start redcon server: %s", e.Error()))
		}
	}()
}

func (r *redconTester) initClient() {
	r.client = redis.NewClient(&redis.Options{Addr: "localhost:8080"})
}

func (r *redconTester) doRPC(name string) {
	res, e := r.client.Do("greet", name).Result()
	if e != nil {
		panic(fmt.Sprintf("error sending redcon request: %s", e.Error()))
	}
	data, ok := res.(string)
	if !ok {
		panic(fmt.Sprintf("unexpected redcon response type"))
	}
	if data != fmt.Sprintf("hello, %s", name) {
		panic(fmt.Sprintf("wrong redcon answer: %s", data))
	}
}

func (r *redconTester) closeClient() {
	if e := r.client.Close(); e != nil {
		panic(fmt.Sprintf("could not close redcon client: %s", e.Error()))
	}
}

func (r *redconTester) closeServer() {
	if e := r.lis.Close(); e != nil {
		panic(fmt.Sprintf("could not close redcon server: %s", e.Error()))
	}
}

func handler(conn redcon.Conn, cmd redcon.Command) {
	cmdName := strings.ToLower(string(cmd.Args[0]))
	if cmdName != "greet" {
		conn.WriteError(fmt.Sprintf("invalid cmd %s", cmdName))
		return
	}
	name := strings.ToLower(string(cmd.Args[1]))
	conn.WriteString(fmt.Sprintf("hello, %s", name))
}

func accept(_ redcon.Conn) bool {
	return true
}

func closed(_ redcon.Conn, _ error) {}
