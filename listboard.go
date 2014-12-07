package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"html/template"
	//"github.com/gosimple/slug"
)

type Listboard struct {
	config *Config
	db     *Database
}

type TemplateData map[string]interface{}

var helperFuncs = template.FuncMap{}

func NewListboard() *Listboard {
	return &Listboard{}
}

func NewTemplateData() *TemplateData {
	return &TemplateData{}
}

func render(data *TemplateData, w http.ResponseWriter, r *http.Request, filenames ...string) {
	t := template.New("layout.html")
	t.Funcs(helperFuncs)
	if err := template.Must(t.ParseFiles(filenames...)).Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (l *Listboard) Run() {
	l.config = NewConfig()
	l.db = NewDatabase(l.config)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := NewTemplateData()
		render(data, w, r, "templates/layout.html", "templates/index.html")
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
