package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aquilax/listboard/database/sqlite"
	"github.com/aquilax/listboard/node"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sitemap"
)

const (
	itemsPerPage  = 100
	statusEnabled = 1
	levelRoot     = iota
	levelList
	levelVote
)

type ListBoard struct {
	config *Config
	m      *Model
	tp     *TransPool
	sg     *SpamGuard
}

type ValidationErrors []string

type appHandler func(http.ResponseWriter, *http.Request) error

func NewListBoard() *ListBoard {
	return &ListBoard{}
}

func (l *ListBoard) Run(args []string) {
	var err error

	if os.Getenv("GO_ENV") != "" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	}

	l.config = NewConfig()
	if err = l.config.Load(args); err != nil {
		panic(err)
	}

	l.sg = NewSpamGuard(l.config.PostBlockExpire)

	db := sqlite.New()
	if err := db.Init(l.config.Database, l.config.Dsn); err != nil {
		log.Fatal(err)
	}

	l.m = NewModel(db)

	l.tp = NewTransPool(l.config.Translations)

	r := mux.NewRouter()
	r.HandleFunc("/", appHandler(l.indexHandler).ServeHTTP).Methods("GET")
	r.HandleFunc("/feed.xml", appHandler(l.feedHandler).ServeHTTP).Methods("GET")
	r.HandleFunc("/all.xml", appHandler(l.feedAllHandler).ServeHTTP).Methods("GET")
	r.HandleFunc("/sitemap.xml", appHandler(l.sitemapHandler).ServeHTTP).Methods("GET")

	r.HandleFunc("/add.html", appHandler(l.addFormHandler).ServeHTTP).Methods("GET", "POST")
	r.HandleFunc("/edit.html", appHandler(l.editFormHandler).ServeHTTP).Methods("GET", "POST")
	r.HandleFunc("/list/{listID}/{slug}", appHandler(l.listHandler).ServeHTTP).Methods("GET", "POST")
	r.HandleFunc("/vote/{itemID}/{slug}", appHandler(l.voteHandler).ServeHTTP).Methods("GET", "POST")

	// Static assets
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public_html")))

	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		port = l.config.Server
	}

	log.Printf("Starting server at %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		httpError, ok := err.(HTTPError)
		if ok {
			http.Error(w, httpError.Message, httpError.Code)
			return
		}
		// Default to 500 Internal Server Error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getPageNumber(pageStr string) int {
	page := 1
	var err error
	if len(pageStr) != 0 {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("%s is not a valid page number", pageStr)
			page = 1
		}
	}
	return page - 1
}

func (l *ListBoard) indexHandler(w http.ResponseWriter, r *http.Request) error {
	page := getPageNumber(r.URL.Query().Get("page"))
	sc := l.config.getSiteConfig(l.getToken(r))
	s := NewSession(sc, l.tp.Get(sc.Language))
	s.AddPath("", s.Lang("Home"))
	s.Set("Lists", l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, itemsPerPage, (page*itemsPerPage), "updated DESC"))
	s.Set("Pagination", Pagination(PaginationConfig{
		page:  page + 1,
		ipp:   itemsPerPage,
		total: l.m.mustGetTotal(sc.DomainID, node.RootNodeID),
		url:   "?",
		param: "page",
	}))
	return s.render(w, r, sc.templatePath("layout.html"), sc.templatePath("index.html"))
}

func (l *ListBoard) addFormHandler(w http.ResponseWriter, r *http.Request) error {
	sc := l.config.getSiteConfig(l.getToken(r))

	var errors ValidationErrors
	var n node.Node
	tr := l.tp.Get(sc.Language)
	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, node.RootNodeID, levelRoot, tr)
			if len(errors) == 0 {
				// save and redirect
				id, err := l.m.addNode(&n)
				if err != nil {
					return &HTTPError{Err: err, Code: http.StatusInternalServerError}
				}
				url := "/list/" + strconv.Itoa(id) + "/" + hfSlug(n.Title)
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)
	s.Set("Errors", errors)
	s.Set("Form", n)
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", s.Lang("New list"))
	s.Set("Subtitle", s.Lang("New list"))
	return s.render(w, r, sc.templatePath("layout.html"), sc.templatePath("add.html"), sc.templatePath("form.html"))
}

func (l *ListBoard) editFormHandler(w http.ResponseWriter, r *http.Request) error {
	sc := l.config.getSiteConfig(l.getToken(r))

	var errors ValidationErrors
	var n node.Node
	var item *node.Node
	var nodeID node.NodeID
	var err error
	nodeID = r.URL.Query().Get("id")

	tr := l.tp.Get(sc.Language)
	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			level := 0
			parentID := node.RootNodeID
			if level, err = strconv.Atoi(r.FormValue("level")); err != nil {
				level = 0
			}
			parentID = r.FormValue("parent_id")
			n, errors = l.validateForm(r, sc.DomainID, parentID, level, tr)
			if len(errors) == 0 {
				n.ID = nodeID
				// save and redirect
				if err := l.m.editNode(&n); err != nil {
					return &HTTPError{Err: err, Code: http.StatusInternalServerError}
				}
				url := getUrl("http://"+r.Host, n)
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
	}

	item, err = l.m.getNode(sc.DomainID, nodeID)
	if err != nil {
		return err
	}

	s := NewSession(sc, tr)
	s.Set("Errors", errors)
	s.Set("Form", item)
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("", s.Lang("Edit"))
	s.Set("Subtitle", s.Lang("Edit"))
	return s.render(w, r, sc.templatePath("layout.html"), sc.templatePath("edit.html"), sc.templatePath("form.html"))
}

func (l *ListBoard) listHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	listID := node.NodeID(vars["listID"])

	sc := l.config.getSiteConfig(l.getToken(r))
	tr := l.tp.Get(sc.Language)

	var errors ValidationErrors
	var n node.Node

	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, listID, levelList, tr)
			if len(errors) == 0 {
				// save and redirect
				id, err := l.m.addNode(&n)
				if err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				url := "/list/" + listID + "/" + hfSlug(n.Title) + "#I" + strconv.Itoa(id)
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)

	s.Set("Errors", errors)
	s.Set("Form", n)
	list, err := l.m.getNode(sc.DomainID, listID)
	if err != nil {
		return err
	}
	page := getPageNumber(r.URL.Query().Get("page"))
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
	return s.render(w, r, sc.templatePath("layout.html"), sc.templatePath("list.html"), sc.templatePath("form.html"))
}

func (l *ListBoard) voteHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	itemID := node.NodeID(vars["itemID"])

	sc := l.config.getSiteConfig(l.getToken(r))
	tr := l.tp.Get(sc.Language)
	var errors ValidationErrors
	var n node.Node
	item, err := l.m.getNode(sc.DomainID, itemID)
	if err != nil {
		return err
	}
	if r.Method == "POST" {
		if !inHoneypot(r.FormValue("name")) {
			n, errors = l.validateForm(r, sc.DomainID, itemID, levelVote, tr)
			if len(errors) == 0 {
				id, err := l.m.addNode(&n)
				if err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				if err := l.m.Vote(sc.DomainID, n.Vote, id, itemID, item.ParentID); err != nil {
					return &HTTPError{
						Err:     err,
						Message: err.Error(),
						Code:    http.StatusInternalServerError,
					}
				}
				http.Redirect(w, r, r.URL.String()+"#I"+strconv.Itoa(id), http.StatusFound)
			}
		}
	}
	s := NewSession(sc, tr)
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
	if len(n.Title) == 0 {
		n.Title = s.Lang("Re") + ": " + item.Title
	}
	s.Set("Form", n)
	s.Set("Items", l.m.mustGetChildNodes(sc.DomainID, itemID, itemsPerPage, 0, "created DESC"))
	s.Set("FormTitle", s.Lang("New vote"))
	s.AddPath("/", s.Lang("Home"))
	s.AddPath("/list/"+list.ID+"/"+hfSlug(list.Title), list.Title)
	s.AddPath("", item.Title)
	return s.render(w, r, sc.templatePath("layout.html"), sc.templatePath("vote.html"), sc.templatePath("form.html"))
}

func (l *ListBoard) feed(w http.ResponseWriter, sc *SiteConfig, baseURL string, nodes *node.NodeList) error {
	feed := &feeds.Feed{
		Title:       sc.Title,
		Link:        &feeds.Link{Href: baseURL},
		Description: sc.Description,
		Author:      &feeds.Author{Name: sc.AuthorName, Email: sc.AuthorEmail},
		Created:     time.Now(),
	}
	for _, node := range *nodes {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       node.Title,
			Link:        &feeds.Link{Href: getUrl(baseURL, node)},
			Description: string(node.Rendered),
			Created:     node.Created,
		})
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	return feed.WriteRss(w)
}

func (l *ListBoard) feedHandler(w http.ResponseWriter, r *http.Request) error {
	sc := l.config.getSiteConfig(l.getToken(r))
	nodes := l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, 20, 0, "created DESC")
	baseUrl := "http://" + r.Host
	return l.feed(w, sc, baseUrl, nodes)
}

func (l *ListBoard) feedAllHandler(w http.ResponseWriter, r *http.Request) error {
	sc := l.config.getSiteConfig(l.getToken(r))
	nodes := l.m.mustGetAllNodes(sc.DomainID, 20, 0, "created DESC")
	baseUrl := "http://" + r.Host
	return l.feed(w, sc, baseUrl, nodes)
}

func (l *ListBoard) sitemapHandler(w http.ResponseWriter, r *http.Request) error {
	sc := l.config.getSiteConfig(l.getToken(r))
	nodes := l.m.mustGetChildNodes(sc.DomainID, node.RootNodeID, 1000, 0, "created")
	var urlSet sitemap.URLSet
	for _, n := range *nodes {
		urlSet.URLs = append(urlSet.URLs, sitemap.URL{
			Loc:        "http://" + r.Host + "/list/" + n.ID + "/" + hfSlug(n.Title),
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

func (l *ListBoard) validateForm(r *http.Request, domainID node.DomainID, parentID node.NodeID, level int, ln *Language) (node.Node, ValidationErrors) {
	n := node.Node{
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

func (l *ListBoard) getToken(r *http.Request) string {
	return r.Header.Get(l.config.Token)
}

func getUrl(baseURL string, n node.Node) string {
	switch n.Level {
	case levelRoot:
		return baseURL + "/list/" + n.ID + "/" + hfSlug(n.Title)
	case levelList:
		return baseURL + "/list/" + n.ParentID + "/item#I" + n.ID
	default:
		return baseURL + "/vote/" + n.ParentID + "/item#I" + n.ID
	}
}
