package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aquilax/listboard/database"
	"github.com/aquilax/listboard/database/cached"
	"github.com/aquilax/listboard/database/memory"
	"github.com/aquilax/listboard/database/postgres"
	"github.com/aquilax/listboard/database/sqlite"
	"github.com/aquilax/listboard/node"
	"github.com/gorilla/feeds"
	"github.com/julienschmidt/httprouter"
	"github.com/sourcegraph/sitemap"
)

const (
	itemsPerPage  = 100
	statusEnabled = 1

	levelRoot = 1
	levelList = 2
	levelVote = 3
)

type templateIndex = string

var templateMap = map[templateIndex][]string{
	"index": {"layout.html", "index.html"},
	"add":   {"layout.html", "add.html", "form.html"},
	"edit":  {"layout.html", "edit.html", "form.html"},
	"list":  {"layout.html", "list.html", "form.html"},
	"vote":  {"layout.html", "vote.html", "form.html"},
}

type ListBoard struct {
	config        *Config
	m             *Model
	tp            *TransPool
	sg            *SpamGuard
	templateCache map[string]*template.Template
}

type ValidationErrors []string

func NewListBoard() *ListBoard {
	return &ListBoard{}
}

func (l *ListBoard) Run(args []string) {
	var err error

	l.config = NewConfig()

	// setup logger
	if l.config.Environment() != "" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	}

	// load config
	if err = l.config.Load(args); err != nil {
		log.Fatal(err)
	}

	l.sg = NewSpamGuard(l.config.PostBlockExpire)

	// Set up database
	db, err := getDatabaseAdapter(l.config.Database, l.config.CacheDB)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using database: %s cached: %t\n", l.config.Database, l.config.CacheDB)
	if err := db.Open(l.config.Database, l.config.DSN()); err != nil {
		log.Fatal(err)
	}

	l.m = NewModel(db)

	l.tp = NewTransPool(l.config.Translations)
	l.templateCache = getTemplateCache(l.tp, l.config.Servers)

	// pre-load translations
	for i := range l.config.Servers {
		l.tp.Get(l.config.Servers[i].Language)
	}

	r := httprouter.New()

	r.HandlerFunc(http.MethodGet, "/", l.withSession(l.indexHandler))
	r.HandlerFunc(http.MethodGet, "/feed.xml", l.withSession(l.feedHandler))
	r.HandlerFunc(http.MethodGet, "/all.xml", l.withSession(l.feedAllHandler))
	r.HandlerFunc(http.MethodGet, "/sitemap.xml", l.withSession(l.sitemapHandler))

	r.HandlerFunc(http.MethodGet, "/add.html", l.withSession(l.addFormHandler))
	r.HandlerFunc(http.MethodPost, "/add.html", l.withSession(l.addFormHandler))

	r.HandlerFunc(http.MethodGet, "/edit.html", l.withSession(l.editFormHandler))
	r.HandlerFunc(http.MethodPost, "/edit.html", l.withSession(l.editFormHandler))

	r.HandlerFunc(http.MethodGet, "/list/:listID/:slug", l.withSession(l.listHandler))
	r.HandlerFunc(http.MethodPost, "/list/:listID/:slug", l.withSession(l.listHandler))

	r.HandlerFunc(http.MethodGet, "/vote/:itemID/:slug", l.withSession(l.voteHandler))
	r.HandlerFunc(http.MethodPost, "/vote/:itemID/:slug", l.withSession(l.voteHandler))

	// Static assets
	r.NotFound = http.FileServer(http.Dir("./public_html/"))

	shutdownWait := time.Second * 15
	port := l.config.Port()

	srv := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 50,
		Handler:      r,
	}

	go func() {
		log.Printf("starting server at %s", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownWait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)
}

func (l ListBoard) withSession(f func(w http.ResponseWriter, r *http.Request, s *Session) error) func(w http.ResponseWriter, r *http.Request) {
	getToken := func(r *http.Request) string {
		return r.Header.Get(l.config.Token)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sc := l.config.getSiteConfig(getToken(r))
		tr := l.tp.Get(sc.Language)
		s := NewSession(sc, tr)
		if err := f(w, r, s); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				log.Printf("404 %d %s %s%s\n", time.Since(start).Nanoseconds(), r.Method, sc.Domain, r.URL)
				return
			}
			httpError, ok := err.(HTTPError)
			if ok {
				http.Error(w, http.StatusText(httpError.Code), httpError.Code)
				log.Printf("%d %d %s %s%s\n", httpError.Code, time.Since(start).Nanoseconds(), r.Method, sc.Domain, r.URL)
				return
			}
			// Default to 500 Internal Server Error
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Printf("500 %d %s %s%s\n", time.Since(start).Nanoseconds(), r.Method, sc.Domain, r.URL)
			return
		}
		log.Printf("200 %d %s %s%s\n", time.Since(start).Nanoseconds(), r.Method, sc.Domain, r.URL)
	}
}

func (l ListBoard) indexHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	page := getPageNumber(r.URL.Query().Get("page"))
	sc := s.SiteConfig()

	s.AddPath("", s.Lang("Home"))
	s.Set("CanonicalURL", s.GetAbsoluteURL("/").String())
	s.Set("Lists", l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, itemsPerPage, (page*itemsPerPage), "updated DESC"))
	s.Set("Pagination", Pagination(PaginationConfig{
		page:  page + 1,
		ipp:   itemsPerPage,
		total: l.m.mustGetTotal(sc.DomainID, node.RootNodeID),
		url:   "?",
		param: "page",
	}))
	return l.renderTemplateKey(w, r, s, sc, "index")
}

func (l ListBoard) addFormHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	sc := s.SiteConfig()

	var errors ValidationErrors
	n := &node.Node{
		ParentID: node.RootNodeID,
		Level:    levelRoot,
	}
	tr := l.tp.Get(sc.Language)

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, node.RootNodeID, levelRoot, tr)
			if len(errors) == 0 {
				// save and redirect
				nodeID, err := l.m.addNode(n)
				if err != nil {
					return &HTTPError{Err: err, Code: http.StatusInternalServerError}
				}
				n.ID = nodeID
				http.Redirect(w, r, s.GetNodeURL(n).String(), http.StatusFound)
				return nil
			}
		}
	}

	s.Set("Errors", errors)
	s.Set("Form", n)
	s.Set("CanonicalURL", s.GetAbsoluteURL("/add.html").String())
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", s.Lang("New list"))
	s.Set("Subtitle", s.Lang("New list"))
	return l.renderTemplateKey(w, r, s, sc, "add")
}

func (l ListBoard) editFormHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	sc := s.SiteConfig()

	var errors ValidationErrors
	var n *node.Node
	var item *node.Node
	var nodeID node.NodeID
	var err error
	nodeID = r.URL.Query().Get("id")

	tr := l.tp.Get(sc.Language)

	item, err = l.m.getNode(sc.DomainID, nodeID)
	if err != nil {
		return err
	}

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, item.ParentID, item.Level, tr)
			if len(errors) == 0 {
				n.ID = item.ID
				// save and redirect
				if err := l.m.editNode(n); err != nil {
					return &HTTPError{Err: err, Code: http.StatusInternalServerError}
				}
				url := s.GetNodeURL(n).String()
				http.Redirect(w, r, url, http.StatusFound)
				return nil
			}
		}
	}

	s.Set("Errors", errors)
	s.Set("Form", item)
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", s.Lang("Edit"))
	s.Set("Subtitle", s.Lang("Edit"))
	return l.renderTemplateKey(w, r, s, sc, "edit")
}

func (l ListBoard) listHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	params := httprouter.ParamsFromContext(r.Context())
	listID := node.NodeID(params.ByName("listID"))

	sc := s.SiteConfig()
	tr := l.tp.Get(sc.Language)

	var errors ValidationErrors
	list, err := l.m.getNode(sc.DomainID, listID)
	if err != nil {
		return HTTPError{Err: err, Code: http.StatusNotFound, Message: "404 Page not found"}
	}

	n := &node.Node{
		ParentID: list.ID,
		Level:    levelList,
	}

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, listID, levelList, tr)
			if len(errors) == 0 {
				// save and redirect
				nodeID, err := l.m.addNode(n)
				if err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				n.ID = nodeID
				http.Redirect(w, r, s.GetNodeURL(n).String(), http.StatusFound)
				return nil
			}
		}
	}

	page := getPageNumber(r.URL.Query().Get("page"))

	s.Set("Errors", errors)
	s.Set("CanonicalURL", s.GetNodeURL(list).String())
	s.Set("Form", n)
	s.Set("List", list)
	s.Set("Items", l.m.mustGetChildNodes(sc.DomainID, listID, itemsPerPage, (page*itemsPerPage), "vote DESC, created"))
	s.Set("FormTitle", s.Lang("New suggestion"))
	s.Set("Subtitle", list.Title)
	s.Set("Description", list.Title)
	s.Set("Pagination", Pagination(PaginationConfig{
		page:  page + 1,
		ipp:   itemsPerPage,
		total: l.m.mustGetTotal(sc.DomainID, listID),
		url:   "?",
		param: "page",
	}))

	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", list.Title)
	return l.renderTemplateKey(w, r, s, sc, "list")
}

func (l ListBoard) voteHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	params := httprouter.ParamsFromContext(r.Context())
	itemID := node.NodeID(params.ByName("itemID"))

	sc := s.SiteConfig()
	tr := l.tp.Get(sc.Language)
	var errors ValidationErrors

	item, err := l.m.getNode(sc.DomainID, itemID)
	if err != nil {
		return err
	}

	n := &node.Node{
		ParentID: item.ID,
		Title:    s.Lang("Re") + ": " + item.Title,
		Level:    levelVote,
	}

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, item.ID, levelVote, tr)
			if len(errors) == 0 {
				newNodeID, err := l.m.addNode(n)
				if err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				if err := l.m.Vote(sc.DomainID, n.Vote, newNodeID, item.ID, item.ParentID); err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				http.Redirect(w, r, r.URL.String()+"#I"+newNodeID, http.StatusFound)
				return nil
			}
		}
	}
	s.Set("Subtitle", item.Title)
	s.Set("Description", item.Title)
	s.Set("ShowVote", true)
	s.Set("Errors", errors)
	list, err := l.m.getNode(sc.DomainID, item.ParentID)
	if err != nil {
		return err
	}
	s.Set("List", list)
	s.Set("Item", item)
	s.Set("CanonicalURL", s.GetNodeURL(list).String())
	s.Set("Form", n)
	s.Set("Items", l.m.mustGetChildNodes(sc.DomainID, itemID, itemsPerPage, 0, "created DESC"))
	s.Set("FormTitle", s.Lang("New vote"))
	s.AddPath("/", s.Lang("Home"))
	s.AddPath(s.GetNodeURL(list).String(), list.Title)
	s.AddPath("", item.Title)
	return l.renderTemplateKey(w, r, s, sc, "vote")
}

func (l ListBoard) feed(w http.ResponseWriter, sc *SiteConfig, baseURL string, s *Session, nodes *node.NodeList) error {
	feed := &feeds.Feed{
		Title:       sc.Title,
		Link:        &feeds.Link{Href: s.GetAbsoluteURL("/").String()},
		Description: sc.Description,
		Author:      &feeds.Author{Name: sc.AuthorName, Email: sc.AuthorEmail},
		Created:     time.Now(),
	}
	for _, n := range *nodes {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       n.Title,
			Link:        &feeds.Link{Href: s.GetNodeURL(&n).String()},
			Description: string(n.Rendered),
			Created:     n.Created,
		})
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	return feed.WriteRss(w)
}

func (l ListBoard) feedHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	sc := s.SiteConfig()
	nodes := l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, 20, 0, "created DESC")
	baseUrl := sc.Domain
	return l.feed(w, sc, baseUrl, s, nodes)
}

func (l ListBoard) feedAllHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	sc := s.SiteConfig()
	nodes := l.m.mustGetAllNodes(sc.DomainID, 20, 0, "created DESC")
	baseUrl := sc.Domain
	return l.feed(w, sc, baseUrl, s, nodes)
}

func (l ListBoard) sitemapHandler(w http.ResponseWriter, r *http.Request, s *Session) error {
	sc := s.SiteConfig()
	nodes := l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, 1000, 0, "created")
	var urlSet sitemap.URLSet
	for _, n := range *nodes {
		urlSet.URLs = append(urlSet.URLs, sitemap.URL{
			Loc:        s.GetNodeURL(&n).String(),
			LastMod:    &n.Created,
			ChangeFreq: sitemap.Daily,
			Priority:   0.7,
		})
	}
	xml, err := sitemap.Marshal(&urlSet)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/xml")
	_, err = w.Write(xml)
	return err
}

func (l ListBoard) validateForm(r *http.Request, domainID node.DomainID, parentID node.NodeID, level int, ln *Language) (*node.Node, ValidationErrors) {
	n := &node.Node{
		ParentID: parentID,
		DomainID: domainID,
		Title:    strings.TrimSpace(r.FormValue("title")),
		Vote:     getVote(r.FormValue("vote")),
		TripCode: getTripCode(r.FormValue("password")),
		Body:     r.FormValue("body"),
		Status:   statusEnabled,
		Level:    level,
	}
	errors := ValidationErrors{}
	if !l.sg.CanPost(r.RemoteAddr) {
		errors = append(errors, ln.Lang("Please wait before posting again"))
		// Exit fast on flood
		return n, errors
	}
	if len(n.Title) < 3 {
		errors = append(errors, ln.Lang("Title must be at least 3 characters long"))
	}
	if len(n.Body) < 10 {
		errors = append(errors, ln.Lang("Please, write something"))
	} else {
		n.Rendered = renderText(n.Body)
		// Check again after the rendering
		if len(n.Rendered) < 10 {
			errors = append(errors, ln.Lang("Please, write something"))
		}
	}
	return n, errors
}

func getTemplateCache(tp *TransPool, sites map[string]*SiteConfig) map[string]*template.Template {
	templateCache := make(map[string]*template.Template)

	for _, sc := range sites {
		s := NewSession(sc, tp.Get(sc.Language))
		for name, fileNames := range templateMap {
			key := sc.getTemplateCacheKey(name)
			if _, found := templateCache[key]; !found {
				layout := template.New("layout.html")
				layout.Funcs(s.getHelpers())
				files := make([]string, len(fileNames))
				for i, fileName := range fileNames {
					files[i] = sc.templatePath(fileName)
				}
				templateCache[key] = template.Must(layout.ParseFiles(files...))
			}
		}
	}
	return templateCache
}

func getDatabaseAdapter(db string, useCache bool) (database.Database, error) {
	dbAdapter, err := func(db string) (database.Database, error) {
		if db == "sqlite" || db == "sqlite3" {
			return sqlite.New(), nil
		}
		if db == "postgres" {
			return postgres.New(), nil
		}
		if db == "memory" {
			return memory.New(), nil
		}
		return nil, fmt.Errorf("database %s is not supported", db)
	}(db)

	if useCache {
		return cached.New(dbAdapter), err
	}
	return dbAdapter, err
}

func (l ListBoard) renderTemplateKey(w http.ResponseWriter, r *http.Request, s *Session, sc *SiteConfig, key templateIndex) error {
	if t, found := l.templateCache[sc.getTemplateCacheKey(key)]; found {
		return s.renderTemplate(w, r, t)
	}
	if templates, found := templateMap[key]; found {
		templatePaths := make([]string, len(templates))
		for i := range templates {
			templatePaths[i] = sc.templatePath(templates[i])
		}
		return s.render(w, r, templatePaths...)
	}
	return fmt.Errorf("templates not found for key %s", key)
}
