package got

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fmtPanic(a ...interface{}) {
	panic(fmt.Sprint(a...))
}

func assert(condition bool, a ...interface{}) {
	if !condition {
		fmtPanic(a...)
	}
}

func nilOrPanic(err error, a ...interface{}) {
	assert(err == nil, append(a, " Error: ", err)...)
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "GET", "Expected GET method")
		w.WriteHeader(200)
		w.Write([]byte("Hello back"))
	}))
	defer server.Close()

	g := New()
	req := g.NewRequest("GET", server.URL, nil)
	res, err := req.Send()
	nilOrPanic(err, "Test request failed")
	assert(string(res.Body) == "Hello back", "Wrong body recieved!")
	assert(res.StatusCode == 200, "non 200")
	assert(res.Attempts == 1, "More than one attempt")
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "POST", "Expected POST method")
		body, err := ioutil.ReadAll(r.Body)
		nilOrPanic(err, "Failed to read body")
		assert(string(body) == "Hello World", "Wrong body")
		w.WriteHeader(200)
		w.Write([]byte("Hello back"))
	}))
	defer server.Close()

	g := New()
	req := g.NewRequest("POST", server.URL, []byte("Hello World"))
	res, err := req.Send()
	nilOrPanic(err, "Test request failed")
	assert(string(res.Body) == "Hello back", "Wrong body recieved!")
	assert(res.StatusCode == 200, "non 200")
	assert(res.Attempts == 1, "More than one attempt")
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "PUT", "Expected PUT method")
		body, err := ioutil.ReadAll(r.Body)
		nilOrPanic(err, "Failed to read body")
		assert(string(body) == "Hello World", "Wrong body")
		w.WriteHeader(200)
		w.Write([]byte("Hello back"))
	}))
	defer server.Close()

	g := New()
	req := g.NewRequest("PUT", server.URL, []byte("Hello World"))
	res, err := req.Send()
	nilOrPanic(err, "Test request failed")
	assert(string(res.Body) == "Hello back", "Wrong body recieved!")
	assert(res.StatusCode == 200, "non 200")
	assert(res.Attempts == 1, "More than one attempt")
}

func TestRetries(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "POST", "Expected POST method")
		body, err := ioutil.ReadAll(r.Body)
		nilOrPanic(err, "Failed to read body")
		assert(string(body) == "Hello World", "Wrong body")

		calls++
		if calls > 3 {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(500)
		}
		w.Write([]byte("Hello back"))
	}))
	defer server.Close()

	g := New()
	req := g.NewRequest("POST", server.URL, []byte("Hello World"))
	res, err := req.Send()
	nilOrPanic(err, "Test request failed")
	assert(string(res.Body) == "Hello back", "Wrong body recieved!")
	assert(res.StatusCode == 200, "non 200")
	assert(res.Attempts == 4, "not 4 attempt")
}

func TestMaxSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "GET", "Expected GET method")
		w.WriteHeader(200)
		w.Write([]byte("1234"))
	}))
	defer server.Close()

	g := New()
	g.MaxSize = 4
	req := g.NewRequest("GET", server.URL, nil)
	res, err := req.Send()
	nilOrPanic(err, "Test request failed")
	assert(string(res.Body) == "1234", "Wrong body recieved!")
	assert(res.StatusCode == 200, "non 200")
	assert(res.Attempts == 1, "More than one attempt")
}

func TestMaxSizeTooBig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert(r.Method == "GET", "Expected GET method")
		w.WriteHeader(200)
		w.Write([]byte("1234"))
	}))
	defer server.Close()

	g := New()
	g.MaxSize = 3
	req := g.NewRequest("GET", server.URL, nil)
	res, err := req.Send()
	assert(err == ErrResponseTooLarge, "Expected ErrResponseTooLarge")
	assert(res.Body == nil, "Expected nil body")
}
