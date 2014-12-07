package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	//"github.com/gosimple/slug"
)

type Listboard struct {
	config *Config
	db     *Database
}

func NewListboard() *Listboard {
	return &Listboard{}
}

func (l *Listboard) Run() {
	l.config = NewConfig()
	l.db = NewDatabase(l.config)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "index")
	})
	r.HandleFunc("/list/{listId}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "list")
	})
	r.HandleFunc("/list/vote/{listId}/{itemId}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "vote")
	})
	http.Handle("/", r)
	if err := http.ListenAndServe(l.config.Server, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
