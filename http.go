package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type httpTester struct {
	srv *http.Server
}

func (h *httpTester) init() {
	h.srv = &http.Server{Addr: ":8080"}
	http.HandleFunc("/greet", func(writer http.ResponseWriter, request *http.Request) {
		if e := request.ParseForm(); e != nil {
			panic(fmt.Sprintf("error parsing http form: %v", e.Error()))
		}
		name := request.Form.Get("name")
		if _, e := fmt.Fprintf(writer, "hello, %s", name); e != nil {
			panic(fmt.Sprintf("error writing http response: %v", e.Error()))
		}
	})
	go func() {
		if e := h.srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			panic(fmt.Sprintf("could not start http server: %v", e.Error()))
		}
	}()
	time.Sleep(time.Second)
}

func (h *httpTester) close() {
	if e := h.srv.Close(); e != nil {
		panic(fmt.Sprintf("could not stop http server: %v", e.Error()))
	}
	time.Sleep(time.Second)
}

func (h *httpTester) doRPC(name string) {
	resp, e := http.PostForm("http://localhost:8080/greet", url.Values{"name": {name}})
	if e != nil {
		panic(fmt.Sprintf("http error: %s", e.Error()))
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		panic(fmt.Sprintf("could not read http response: %v", e.Error()))
	}
	if string(body) != fmt.Sprintf("hello, %s", name) {
		panic(fmt.Sprintf("wrong http answer: %s", string(body)))
	}
}
