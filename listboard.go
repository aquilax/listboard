package main

import (
	. "github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sitemap"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	itemsPerPage = 10
)

type Listboard struct {
	config *Config
	db     *Database
}

type TemplateData map[string]interface{}
type ValidationErrors []string

var helperFuncs = template.FuncMap{
	"lang": hfLang,
	"time": hfTime,
	"slug": hfSlug,
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
	r.HandleFunc("/feed.xml", http.HandlerFunc(l.feedHandler)).Methods("GET")
	r.HandleFunc("/all.xml", http.HandlerFunc(l.feedAlllHandler)).Methods("GET")
	r.HandleFunc("/sitemap.xml", http.HandlerFunc(l.sitemapHandler)).Methods("GET")

	r.HandleFunc("/add.html", http.HandlerFunc(l.addFormHandler)).Methods("GET", "POST")
	r.HandleFunc("/list/{listId}/{slug}", http.HandlerFunc(l.listHandler)).Methods("GET", "POST")
	r.HandleFunc("/list/{listId}/{itemId}/vote.html", http.HandlerFunc(l.voteHandler)).Methods("GET", "POST")

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

	var errors ValidationErrors
	var node Node

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			node, errors = validateForm(r, 0)
			if errors == nil {
				// save and redirect
				l.db.addNode(&node)
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}
	}

	data := NewTemplateData(sc)
	data["Errors"] = errors
	data["Form"] = node
	render(&data, w, r, "templates/layout.html", "templates/add.html", "templates/form.html")
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
	data["List"] = l.db.getNode(listId)
	data["Items"] = l.db.getChildNodes(listId, itemsPerPage, 0, "votes")
	render(&data, w, r, "templates/layout.html", "templates/list.html", "templates/form.html")
}

func (l *Listboard) voteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listId, err := strconv.Atoi(vars["listId"])
	if err != nil {
		log.Printf("%s is not a valid list number", listId)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	itemId, err := strconv.Atoi(vars["itemId"])
	if err != nil {
		log.Printf("%s is not a valid item number", listId)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	sc := l.db.getSiteConfig("token")
	data := NewTemplateData(sc)
	data["List"] = l.db.getNode(listId)
	data["Item"] = l.db.getNode(itemId)
	data["Items"] = l.db.getChildNodes(itemId, itemsPerPage, 0, "created")
	render(&data, w, r, "templates/layout.html", "templates/vote.html", "templates/form.html")
}

func (l *Listboard) feedHandler(w http.ResponseWriter, r *http.Request) {
	sc := l.db.getSiteConfig("token")
	feed := &Feed{
		Title:       sc.Title,
		Link:        &Link{Href: "http://" + r.Host + "/"},
		Description: sc.Description,
		Author:      &Author{sc.AuthorName, sc.AuthorEmail},
		Created:     time.Now(),
	}
	nodes := l.db.getChildNodes(0, 20, 0, "created")
	for _, node := range *nodes {
		feed.Items = append(feed.Items, &Item{
			Title:       node.Title,
			Link:        &Link{Href: "http://" + r.Host + "/list/" + strconv.Itoa(node.Id) + "/" + hfSlug(node.Title)},
			Description: node.Rendered,
			Created:     node.Created,
		})
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	err := feed.WriteRss(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *Listboard) feedAlllHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("NOT IMPLEMENTED"))
}

func (l *Listboard) sitemapHandler(w http.ResponseWriter, r *http.Request) {
	nodes := l.db.getChildNodes(0, 1000, 0, "created")
	var urlSet sitemap.URLSet
	for _, node := range *nodes {
		urlSet.URLs = append(urlSet.URLs, sitemap.URL{
			Loc:        "http://" + r.Host + "/list/" + strconv.Itoa(node.Id) + "/" + hfSlug(node.Title),
			LastMod:    &node.Created,
			ChangeFreq: sitemap.Daily,
			Priority:   0.7,
		})
	}
	xml, err := sitemap.Marshal(&urlSet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(xml)
}

func validateForm(r *http.Request, parentId int) (Node, ValidationErrors) {
	node := Node{
		ParentId: parentId,
		Title:    strings.TrimSpace(r.FormValue("title")),
		Vote:     getVote(r.FormValue("vote")),
		Tripcode: tripcode(r.FormValue("password")),
		Body:     r.FormValue("body"),
	}
	errors := ValidationErrors{}
	if len(node.Title) < 3 {
		errors = append(errors, hfLang("Title must be at least 3 characters long"))
	}
	if len(node.Title) < 10 {
		errors = append(errors, hfLang("Please, write something"))
	}
	if len(errors) == 0 {
		node.Rendered = renderText(node.Body)
	}
	return node, errors
}
