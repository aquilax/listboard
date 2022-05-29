package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aquilax/listboard/node"
	"github.com/gorilla/mux"
)

func getTestConfig() *Config {
	c := NewConfig()
	c.Database = "memory"
	c.Dsn = ""
	c.Translations = "./translations/"
	c.Servers = map[string]*SiteConfig{
		"": {
			DomainID: "1",
			Language: "en_US",
		},
	}
	return c
}

func TestNewListBoard(t *testing.T) {
	t.Run("ListBoard is not nil", func(t *testing.T) {
		got := NewListBoard()
		if got == nil {
			t.Errorf("NewListBoard() = %v", got)
		}
	})
}

func TestListBoard(t *testing.T) {
	lb := NewListBoard()
	lb.config = getTestConfig()
	lb.tp = NewTransPool(lb.config.Translations)

	db, _ := getDatabaseAdapter(lb.config.Database)
	db.Open(lb.config.Database, lb.config.Dsn)
	lb.m = NewModel(db)
	listNodeID, _ := lb.m.addNode(&node.Node{
		ID:       "1",
		ParentID: node.RootNodeID,
		Title:    "Test Node",
		DomainID: "1",
	})

	tests := []struct {
		name    string
		req     *http.Request
		handler func(w http.ResponseWriter, r *http.Request, s *Session) error
		assert  func(wr *httptest.ResponseRecorder)
	}{
		{
			"index page loads",
			httptest.NewRequest(http.MethodGet, "/", nil),
			lb.indexHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected 200", wr.Code)
				}

				if !strings.Contains(wr.Body.String(), "powered by") {
					t.Errorf(`response body "%s" does not contain "powered by"`, wr.Body.String())
				}
			},
		},
		{
			"add page loads",
			httptest.NewRequest(http.MethodGet, "/add.html", nil),
			lb.addFormHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected 200", wr.Code)
				}

				if !strings.Contains(wr.Body.String(), "New list") {
					t.Errorf(`response body "%s" does not contain "New list"`, wr.Body.String())
				}
			},
		},
		{
			"list page loads",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/list/"+listNodeID+"/test.html", nil)
				vars := map[string]string{
					"listID": listNodeID,
				}
				return mux.SetURLVars(r, vars)
			}(),
			lb.listHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected 200", wr.Code)
				}

				if !strings.Contains(wr.Body.String(), "Test Node") {
					t.Errorf(`response body "%s" does not contain "Test Node"`, wr.Body.String())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := httptest.NewRecorder()
			lb.withSession(tt.handler)(wr, tt.req)
			tt.assert(wr)
		})
	}

}
