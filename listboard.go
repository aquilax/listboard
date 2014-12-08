package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const (
	itemsPerPage = 10
)

type Listboard struct {
	config *Config
	db     *Database
}

type TemplateData map[string]interface{}

var helperFuncs = template.FuncMap{
	"lang":     hfLang,
	"time":     hfTime,
	"slug":     hfSlug,
	"anchor":   hfAnchor,
	"anchorTr": hfAnchorTr,
}

func NewListboard() *Listboard {
	return &Listboard{}
}

func NewTemplateData(sc *SiteConfig) TemplateData {
	return make(TemplateData)
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

	r.HandleFunc("/", http.HandlerFunc(l.indexHandler)).Methods("GET")
	r.HandleFunc("/add.html", http.HandlerFunc(l.addFormHandler)).Methods("GET")
	r.HandleFunc("/list/{listId}/{slug}", http.HandlerFunc(l.listHandler)).Methods("GET")
	r.HandleFunc("/list/vote/{listId}/{itemId}", func(w http.ResponseWriter, r *http.Request) {
		data := NewTemplateData(nil)
		render(&data, w, r, "templates/layout.html", "templates/index.html")
	})
	http.Handle("/", r)

	if err := http.ListenAndServe(l.config.Server, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (l *Listboard) indexHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("hostname")
	page := 0
	var err error
	if len(pageStr) != 0 {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("%s is not a valid page number", pageStr)
			page = 0
		}
	}
	sc := l.db.getSiteConfig("token")
	data := NewTemplateData(sc)
	data["Lists"] = l.db.getChildNodes(0, itemsPerPage, page, "updated")
	render(&data, w, r, "templates/layout.html", "templates/index.html")
}

func (l *Listboard) addFormHandler(w http.ResponseWriter, r *http.Request) {
	sc := l.db.getSiteConfig("token")
	data := NewTemplateData(sc)
	render(&data, w, r, "templates/layout.html", "templates/add.html")
}

func (l *Listboard) listHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listId, err := strconv.Atoi(vars["listId"])
	if err != nil {
		log.Printf("%s is not a valid list number", listId)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	sc := l.db.getSiteConfig("token")
	data := NewTemplateData(sc)
	data["List"] = l.db.getList(listId)
	data["Items"] = l.db.getChildNodes(0, itemsPerPage, 0, "votes")
	render(&data, w, r, "templates/layout.html", "templates/list.html")
}
