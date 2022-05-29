package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/aquilax/listboard/node"
)

type PathLink struct {
	Url   string
	Label string
}

type Session struct {
	td   TemplateData
	ln   *Language
	path []*PathLink
	slug func(s string) string
	base *url.URL
	sc   *SiteConfig
}

type TemplateData map[string]interface{}

func NewSession(sc *SiteConfig, ln *Language) *Session {
	return &Session{
		td:   NewTemplateData(sc),
		ln:   ln,
		path: []*PathLink{},
		slug: getLanguageSlug(sc.Language),
		base: sc.BaseUrl,
		sc:   sc,
	}
}

func NewTemplateData(sc *SiteConfig) TemplateData {
	td := make(TemplateData)
	td.Set("Title", sc.Title)
	td.Set("Language", sc.Language)
	td.Set("Description", sc.Description)
	td.Set("ShowVote", false)
	td.Set("Css", sc.Css)
	td.Set("FormTitle", "")
	td.Set("Analytics", sc.Analytics)
	td.Set("Domain", sc.Domain)
	td.Set("PostHeader", template.HTML(sc.PostHeader))
	td.Set("PreFooter", template.HTML(sc.PreFooter))
	return td
}

func (s *Session) getHelpers() template.FuncMap {
	return template.FuncMap{
		"lang":     s.Lang,
		"time":     hfTime,
		"slug":     s.Slug,
		"mod":      hfMod,
		"gravatar": hfGravatar,
		"rawHTML": func(value string) template.HTML {
			return template.HTML(fmt.Sprint(value))
		},
	}
}

func (s Session) Lang(text string) string {
	return s.ln.Lang(text)
}

func (s Session) Slug(text string) string {
	return s.slug(text)
}

func (s *Session) render(w http.ResponseWriter, r *http.Request, filenames ...string) error {
	t := template.New("layout.html")
	// Add helper functions
	t.Funcs(s.getHelpers())
	// Add path
	s.td.Set("Path", s.path)
	return template.Must(t.ParseFiles(filenames...)).Execute(w, s.td)
}

func (s *Session) renderTemplate(w http.ResponseWriter, r *http.Request, t *template.Template) error {
	// Add path
	s.td.Set("Path", s.path)
	return t.Execute(w, s.td)
}

func (td TemplateData) Set(name string, value interface{}) {
	td[name] = value
}

func (s *Session) Set(name string, value interface{}) {
	s.td.Set(name, value)
}

func (s *Session) AddPath(url, label string) {
	s.path = append(s.path, &PathLink{url, label})
}

func (s Session) GetNodeURL(n node.Node) *url.URL {
	relativeURL := func(n node.Node) string {
		switch n.Level {
		case levelRoot:
			return "/list/" + n.ID + "/" + s.Slug(n.Title) + ".html"
		case levelList:
			return "/vote/" + n.ID + "/" + s.Slug(n.Title) + ".html#post"
		default:
			return "/vote/" + n.ParentID + "/item.html#I" + n.ID
		}
	}(n)
	return s.GetAbsoluteURL(relativeURL)
}

func (s Session) GetAbsoluteURL(relativeURL string) *url.URL {
	u, err := url.Parse(relativeURL)
	if err != nil {
		log.Printf("invalid url: %s\n", relativeURL)
	}

	return s.base.ResolveReference(u)
}

func (s Session) SiteConfig() *SiteConfig {
	return s.sc
}
