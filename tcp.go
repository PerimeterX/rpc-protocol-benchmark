package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

const poolSize = 15

type tcpTester struct {
	lis        net.Listener
	clientPool *tcpConnectionPool
}

func (t *tcpTester) initServer() {
	var e error
	if t.lis, e = net.Listen("tcp", ":8080"); e != nil {
		panic(fmt.Sprintf("could not create tcp listener: %s", e.Error()))
	}
	go func() {
		for {
			conn, e := t.lis.Accept()
			if e != nil {
				if strings.Contains(e.Error(), "use of closed network connection") {
					return
				}
				panic(fmt.Sprintf("could not accept tcp connection: %s", e.Error()))
			}
			reader := bufio.NewReader(conn)
			go func() {
				for {
					data, e := reader.ReadString('\n')
					if e == io.EOF {
						return
					}
					if e != nil {
						panic(fmt.Sprintf("could not read tcp request: %s", e.Error()))
					}
					if _, e := conn.Write([]byte(fmt.Sprintf("hello, %s", data))); e != nil {
						panic(fmt.Sprintf("could not write tcp response: %s", e.Error()))
					}
				}
			}()
		}
	}()
}

func (t *tcpTester) initClient() {
	t.clientPool = newTCPConnectionPool()
}

func (t *tcpTester) doRPC(name string) {
	c := t.clientPool.acquire()
	defer t.clientPool.release(c)
	if _, e := fmt.Fprintf(c, fmt.Sprintf("%s\n", name)); e != nil {
		panic(fmt.Sprintf("could not send tcp request: %s", e.Error()))
	}
	reader := bufio.NewReader(c)
	data, e := reader.ReadString('\n')
	if e != nil {
		panic(fmt.Sprintf("could not read tcp response: %s", e.Error()))
	}
	if data != fmt.Sprintf("hello, %s\n", name) {
		panic(fmt.Sprintf("wrong tcp answer: %s", data))
	}
}

func (t *tcpTester) closeClient() {
	t.clientPool.close()
}

func (t *tcpTester) closeServer() {
	if e := t.lis.Close(); e != nil {
		panic(fmt.Sprintf("could not close tcp listener: %s", e.Error()))
	}
}

type tcpConnectionPool struct {
	ch chan net.Conn
}

func newTCPConnectionPool() *tcpConnectionPool {
	pool := &tcpConnectionPool{ch: make(chan net.Conn, poolSize)}
	for i := 0; i < poolSize; i++ {
		conn, e := net.Dial("tcp", "localhost:8080")
		if e != nil {
			panic(fmt.Sprintf("could not start tcp client: %s", e.Error()))
		}
		pool.ch <- conn
	}
	return pool
}

func (t *tcpConnectionPool) acquire() net.Conn {
	return <-t.ch
}

func (t *tcpConnectionPool) release(c net.Conn) {
	t.ch <- c
}

func (t *tcpConnectionPool) close() {
	for i := 0; i < poolSize; i++ {
		conn := <-t.ch
		if e := conn.Close(); e != nil {
			panic(fmt.Sprintf("could not close tcp client: %s", e.Error()))
		}
	}
}
