package main

import (
	"html/template"
	"net/http"
)

type Session struct {
	td TemplateData
	ln *Language
}

type TemplateData map[string]interface{}

func NewSession(sc *SiteConfig, ln *Language) *Session {
	return &Session{
		td: NewTemplateData(sc),
		ln: ln,
	}
}

func NewTemplateData(sc *SiteConfig) TemplateData {
	td := make(TemplateData)
	td.Set("Title", sc.Title)
	td.Set("Description", sc.Description)
	td.Set("ShowVote", false)
	td.Set("Css", sc.Css)
	return td
}

func (s *Session) getHelpers() template.FuncMap {
	return template.FuncMap{
		"lang": s.Lang,
		"time": hfTime,
		"slug": hfSlug,
	}
}

func (s *Session) Lang(text string) string {
	return s.ln.Lang(text)
}

func (s *Session) render(w http.ResponseWriter, r *http.Request, filenames ...string) {
	t := template.New("layout.html")
	t.Funcs(s.getHelpers())
	if err := template.Must(t.ParseFiles(filenames...)).Execute(w, s.td); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (td TemplateData) Set(name string, value interface{}) {
	td[name] = value
}

func (s *Session) Set(name string, value interface{}) {
	s.td.Set(name, value)
}
