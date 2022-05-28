package main

import (
	"fmt"
	"html/template"
	"net/http"

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
}

type TemplateData map[string]interface{}

func NewSession(sc *SiteConfig, ln *Language) *Session {
	return &Session{
		td:   NewTemplateData(sc),
		ln:   ln,
		path: []*PathLink{},
	}
}

func NewTemplateData(sc *SiteConfig) TemplateData {
	td := make(TemplateData)
	td.Set("Title", sc.Title)
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
		"slug":     hfSlug,
		"mod":      hfMod,
		"gravatar": hfGravatar,
		"rawHTML": func(value string) template.HTML {
			return template.HTML(fmt.Sprint(value))
		},
		"getUrl": func(n node.Node) string {
			switch n.Level {
			case 0:
				return "/list/" + n.ID + hfSlug(n.Title)
			case 1:
				return "/vote/" + n.ParentID + n.ID + "/vote.html"
			default:
				return "/list/" + n.ID + hfSlug(n.Title)
			}
		},
	}
}

func (s *Session) Lang(text string) string {
	return s.ln.Lang(text)
}

func (s *Session) render(w http.ResponseWriter, r *http.Request, filenames ...string) error {
	t := template.New("layout.html")
	// Add helper functions
	t.Funcs(s.getHelpers())
	// Add path
	s.td.Set("Path", s.path)
	return template.Must(t.ParseFiles(filenames...)).Execute(w, s.td)
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
