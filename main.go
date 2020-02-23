//go:generate protoc --go_out=plugins=grpc:. grpc.proto
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	requestsPerSecond   = 100
	testDurationSeconds = 20
)

type tester interface {
	initServer()
	initClient()
	doRPC(name string)
	closeClient()
	closeServer()
}

var testers = []struct {
	name   string
	tester tester
}{
	{name: "http", tester: &httpTester{}},
	{name: "grpc", tester: &grpcTester{}},
	{name: "grpc_stream", tester: &grpcStreamTester{}},
	{name: "redcon", tester: &redconTester{}},
	{name: "tcp", tester: &tcpTester{}},
}

func main() {
	output := "test,quantile 0.0,quantile 0.5,quantile 0.9,quantile 1.0\n"
	for _, entry := range testers {
		result := doTest(entry.name, entry.tester)
		output += fmt.Sprintf("%s,%s\n", entry.name, strings.Join(result, ","))
	}
	fmt.Println(output)
}

func doTest(name string, t tester) []string {
	fmt.Printf("starting %s\n", name)
	t.initServer()
	t.initClient()
	fmt.Printf("testing %s\n", name)
	c := &calculator{}
	wg := &sync.WaitGroup{}
	totalRequests := testDurationSeconds * requestsPerSecond
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go doRPC(i, t, c, wg)
	}
	wg.Wait()
	fmt.Printf("stopping %s\n", name)
	t.closeClient()
	t.closeServer()
	fmt.Printf("finished %s\n", name)
	return c.quantiles()
}

func doRPC(i int, t tester, c *calculator, wg *sync.WaitGroup) {
	waitTimeSeconds := rand.Intn(testDurationSeconds)
	time.Sleep(time.Duration(waitTimeSeconds) * time.Second)
	timer := c.timer()
	t.doRPC(strconv.Itoa(i))
	timer.done()
	wg.Done()
}
