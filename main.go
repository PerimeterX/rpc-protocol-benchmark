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

const testSize = 10000

type tester interface {
	init()
	close()
	doRPC(name string)
}

var testers = map[string]tester{
	"http":         &httpTester{},
	"grpc":         &grpcTester{},
	"grpc_stream":  &grpcStreamTester{},
	"redconTester": &redconTester{},
}

func main() {
	output := "test,quantile 0.0,quantile 0.5,quantile 0.9,quantile 1.0\n"
	for name, t := range testers {
		result := doTest(name, t)
		output += fmt.Sprintf("%s,%s\n", name, strings.Join(result, ","))
	}
	fmt.Println(output)
}

func doTest(name string, t tester) []string {
	fmt.Printf("starting %s\n", name)
	t.init()
	fmt.Printf("testing %s\n", name)
	c := &calculator{}
	wg := &sync.WaitGroup{}
	for i := 0; i < testSize; i++ {
		wg.Add(1)
		go doRPC(i, t, c, wg)
	}
	wg.Wait()
	fmt.Printf("stopping %s\n", name)
	t.close()
	fmt.Printf("finished %s\n", name)
	return c.quantiles()
}

func doRPC(i int, t tester, c *calculator, wg *sync.WaitGroup) {
	delayMS := rand.Intn(5000)
	time.Sleep(time.Duration(delayMS) * time.Millisecond)
	timer := c.timer()
	t.doRPC(strconv.Itoa(i))
	timer.done()
	wg.Done()
}
