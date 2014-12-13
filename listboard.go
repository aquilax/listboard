package main

import (
	. "github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sitemap"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	itemsPerPage  = 10
	statusEnabled = 1
	levelRoot     = iota
	levelList
	levelVote
)

type Listboard struct {
	config *Config
	m      *Model
	tp     *TransPool
}

type ValidationErrors []string

func NewListboard() *Listboard {
	return &Listboard{}
}

func (l *Listboard) Run(args []string) {
	var err error
	l.config = NewConfig()
	if err = l.config.Load(args); err != nil {
		panic(err)
	}
	l.m = NewModel(l.config)
	if err = l.m.Init(l.config); err != nil {
		panic(err)
	}
	l.tp = NewTransPool(l.config.Translations)

	r := mux.NewRouter()

	r.HandleFunc("/", http.HandlerFunc(l.indexHandler)).Methods("GET")
	r.HandleFunc("/feed.xml", http.HandlerFunc(l.feedHandler)).Methods("GET")
	r.HandleFunc("/all.xml", http.HandlerFunc(l.feedAlllHandler)).Methods("GET")
	r.HandleFunc("/sitemap.xml", http.HandlerFunc(l.sitemapHandler)).Methods("GET")

	r.HandleFunc("/add.html", http.HandlerFunc(l.addFormHandler)).Methods("GET", "POST")
	r.HandleFunc("/list/{listId}/{slug}", http.HandlerFunc(l.listHandler)).Methods("GET", "POST")
	r.HandleFunc("/vote/{itemId}/{slug}", http.HandlerFunc(l.voteHandler)).Methods("GET", "POST")

	// Static assets
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public_html")))

	http.Handle("/", r)

	log.Printf("Starting server at %s", l.config.Server)
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
	sc := l.config.getSiteConfig("token")
	s := NewSession(sc, l.tp.Get(sc.Language))

	s.Set("Lists", l.m.mustGetChildNodes(0, itemsPerPage, page, "updated"))
	s.render(w, r, "templates/layout.html", "templates/index.html")
}

func (l *Listboard) addFormHandler(w http.ResponseWriter, r *http.Request) {
	sc := l.config.getSiteConfig("token")

	var errors ValidationErrors
	var node Node
	tr := l.tp.Get(sc.Language)
	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			node, errors = l.validateForm(r, sc.DomainId, 0, levelRoot, tr)
			if len(errors) == 0 {
				// save and redirect
				id, err := l.m.addNode(&node)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					panic(err)
					return
				}
				url := "/list/" + strconv.Itoa(id) + "/" + hfSlug(node.Title)
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)
	s.Set("Errors", errors)
	s.Set("Form", node)
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", s.Lang("New list"))
	s.Set("Subtitle", s.Lang("New list"))
	s.render(w, r, "templates/layout.html", "templates/add.html", "templates/form.html")
}

func (l *Listboard) listHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listId, err := strconv.Atoi(vars["listId"])
	if err != nil {
		log.Printf("%d is not a valid list number", listId)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	sc := l.config.getSiteConfig("token")
	tr := l.tp.Get(sc.Language)

	var errors ValidationErrors
	var node Node

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			node, errors = l.validateForm(r, sc.DomainId, listId, levelList, tr)
			if len(errors) == 0 {
				// save and redirect
				id, err := l.m.addNode(&node)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					panic(err)
					return
				}
				url := "/list/" + strconv.Itoa(listId) + "/" + hfSlug(node.Title) + "#I" + strconv.Itoa(id)
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)

	s.Set("Errors", errors)
	s.Set("Form", node)
	list := l.m.mustGetNode(listId)
	s.Set("List", list)
	s.Set("Items", l.m.mustGetChildNodes(listId, itemsPerPage, 0, "vote DESC, created"))
	s.Set("FormTitle", s.Lang("New suggestion"))
	s.Set("Subtitle", list.Title)
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", list.Title)
	s.render(w, r, "templates/layout.html", "templates/list.html", "templates/form.html")
}

func (l *Listboard) voteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemId, err := strconv.Atoi(vars["itemId"])
	if err != nil {
		log.Printf("%d is not a valid item number", itemId)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	sc := l.config.getSiteConfig("token")
	tr := l.tp.Get(sc.Language)
	var errors ValidationErrors
	var node Node
	item := l.m.mustGetNode(itemId)

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			node, errors = l.validateForm(r, sc.DomainId, itemId, levelVote, tr)
			if len(errors) == 0 {
				id, err := l.m.addNode(&node)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					panic(err)
					return
				}
				if err := l.m.Vote(node.Vote, id, itemId, item.ParentId); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					panic(err)
					return
				}
				http.Redirect(w, r, r.URL.String()+"#I"+strconv.Itoa(id), http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)
	s.Set("Subtitle", item.Title)
	s.Set("ShowVote", true)
	s.Set("Errors", errors)
	list := l.m.mustGetNode(item.ParentId)
	s.Set("List", list)
	s.Set("Item", item)
	if len(node.Title) == 0 {
		node.Title = s.Lang("Re") + ": " + item.Title
	}
	s.Set("Form", node)
	s.Set("Items", l.m.mustGetChildNodes(itemId, itemsPerPage, 0, "created"))
	s.Set("FormTitle", s.Lang("New vote"))
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("/list/"+strconv.Itoa(list.Id)+"/"+hfSlug(list.Title), list.Title)
	s.AddPath("", item.Title)
	s.render(w, r, "templates/layout.html", "templates/vote.html", "templates/form.html")
}

func (l *Listboard) feed(w http.ResponseWriter, baseURL string, nodes *NodeList) {
	sc := l.config.getSiteConfig("token")
	feed := &Feed{
		Title:       sc.Title,
		Link:        &Link{Href: baseURL},
		Description: sc.Description,
		Author:      &Author{sc.AuthorName, sc.AuthorEmail},
		Created:     time.Now(),
	}
	for _, node := range *nodes {
		feed.Items = append(feed.Items, &Item{
			Title:       node.Title,
			Link:        &Link{Href: baseURL + "/list/" + strconv.Itoa(node.Id) + "/" + hfSlug(node.Title)},
			Description: string(node.Rendered),
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

func (l *Listboard) feedHandler(w http.ResponseWriter, r *http.Request) {
	nodes := l.m.mustGetChildNodes(0, 20, 0, "created")
	baseUrl := "http://" + r.Host
	l.feed(w, baseUrl, nodes)
}

func (l *Listboard) feedAlllHandler(w http.ResponseWriter, r *http.Request) {
	nodes := l.m.mustGetAllNodes(20, 0, "created")
	baseUrl := "http://" + r.Host
	l.feed(w, baseUrl, nodes)
}

func (l *Listboard) sitemapHandler(w http.ResponseWriter, r *http.Request) {
	nodes := l.m.mustGetChildNodes(0, 1000, 0, "created")
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

func (l *Listboard) validateForm(r *http.Request, domainId, parentId, level int, ln *Language) (Node, ValidationErrors) {
	node := Node{
		ParentId: parentId,
		DomainId: domainId,
		Title:    strings.TrimSpace(r.FormValue("title")),
		Vote:     getVote(r.FormValue("vote")),
		Tripcode: getTripcode(r.FormValue("password")),
		Body:     r.FormValue("body"),
		Status:   statusEnabled,
		Level:    level,
	}
	errors := ValidationErrors{}
	if len(node.Title) < 3 {
		errors = append(errors, ln.Lang("Title must be at least 3 characters long"))
	}
	if len(node.Body) < 10 {
		errors = append(errors, ln.Lang("Please, write something"))
	} else {
		node.Rendered = renderText(node.Body)
		// Check again after the rendering
		if len(node.Rendered) < 10 {
			errors = append(errors, ln.Lang("Please, write something"))
		}
	}
	return node, errors
}
