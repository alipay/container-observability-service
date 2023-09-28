package api

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type handler interface {
	ParseRequest() error
	ValidRequest() error
	Process() (int, interface{}, error)
	RequestParams() interface{}
}

type fake struct {
	s     *Server
	r     *http.Request
	w     http.ResponseWriter
	query string
}

type handlerFunc func(*Server, http.ResponseWriter, *http.Request) handler

func fakeFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &fake{
		s: s,
		r: r,
		w: w,
	}
}

func (f *fake) RequestParams() interface{} {
	return f.query
}

func (f *fake) ParseRequest() error {
	if f.r.Method == http.MethodGet {
		f.query = f.r.URL.String()
	} else {
		// FIXME
	}
	return nil
}

func (f *fake) ValidRequest() error {
	return nil
}

func (f *fake) Process() (int, interface{}, error) {
	r := rand.Intn(1000)
	time.Sleep(time.Duration(r) * time.Millisecond)
	return http.StatusOK, strconv.Itoa(r) + "\n", nil
}

var _ handler = &fake{}
