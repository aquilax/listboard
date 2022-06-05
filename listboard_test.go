package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aquilax/listboard/node"
	"github.com/julienschmidt/httprouter"
)

func getTestConfig() *Config {
	c := NewConfig()
	c.Database = "memory"
	c.Dsn = ""
	c.Translations = "./translations/"
	c.PostBlockExpire = "0s"
	baseUrl, _ := url.Parse("http://www.example.com")
	c.Servers = map[string]*SiteConfig{
		"": {
			DomainID: "1",
			Language: "en_US",
			BaseUrl:  baseUrl,
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
	lb.sg = NewSpamGuard(lb.config.PostBlockExpire)

	db, _ := getDatabaseAdapter(lb.config.Database, lb.config.CacheDB)
	db.Open(lb.config.Database, lb.config.Dsn)
	lb.m = NewModel(db)
	listNodeID, _ := lb.m.addNode(&node.Node{
		ID:       "1",
		ParentID: node.RootNodeID,
		Title:    "Test Node",
		DomainID: "1",
		Level:    levelRoot,
		TripCode: getTripCode("test"),
	})

	// TODO
	// r.HandlerFunc(http.MethodPost, "/add.html", l.withSession(l.addFormHandler))
	// r.HandlerFunc(http.MethodPost, "/list/:listID/:slug", l.withSession(l.listHandler))
	// r.HandlerFunc(http.MethodGet, "/vote/:itemID/:slug", l.withSession(l.voteHandler))
	// r.HandlerFunc(http.MethodPost, "/vote/:itemID/:slug", l.withSession(l.voteHandler))

	tests := []struct {
		name    string
		req     *http.Request
		handler func(w http.ResponseWriter, r *http.Request, s *Session) error
		assert  func(wr *httptest.ResponseRecorder)
	}{
		{
			"GET /",
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
			"GET /feed.xml",
			httptest.NewRequest(http.MethodGet, "/feed.xml", nil),
			lb.feedHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected %d", wr.Code, http.StatusOK)
				}

				wantContent := "<title>Test Node</title>"
				if !strings.Contains(wr.Body.String(), wantContent) {
					t.Errorf(`response body "%s" does not contain %s`, wr.Body.String(), wantContent)
				}
			},
		},
		{
			"GET /all.xml",
			httptest.NewRequest(http.MethodGet, "/all.xml", nil),
			lb.feedAllHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected %d", wr.Code, http.StatusOK)
				}

				wantContent := "<title>Test Node</title>"
				if !strings.Contains(wr.Body.String(), wantContent) {
					t.Errorf(`response body "%s" does not contain %s`, wr.Body.String(), wantContent)
				}
			},
		},
		{
			"GET /sitemap.xml",
			httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil),
			lb.sitemapHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected %d", wr.Code, http.StatusOK)
				}

				wantContent := "<loc>http://www.example.com/list/" + listNodeID + "/test-node.html</loc>"
				if !strings.Contains(wr.Body.String(), wantContent) {
					t.Errorf(`response body "%s" does not contain %s`, wr.Body.String(), wantContent)
				}
			},
		},
		{
			"GET /add.html",
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
			"GET /list/:listID/:slug",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/list/"+listNodeID+"/test.html", nil)
				ctx := context.WithValue(r.Context(), httprouter.ParamsKey, httprouter.Params{
					{Key: "listID", Value: listNodeID},
				})
				return r.WithContext(ctx)
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
		{
			"GET /edit.html?id=:listID",
			httptest.NewRequest(http.MethodGet, "/edit.html?id="+listNodeID, nil),
			lb.editFormHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusOK {
					t.Errorf("got HTTP status code %d, expected %d", wr.Code, http.StatusOK)
				}

				wantContent := "Test Node"
				if !strings.Contains(wr.Body.String(), wantContent) {
					t.Errorf(`response body "%s" does not contain %s`, wr.Body.String(), wantContent)
				}
			},
		},
		{
			"POST /edit.html?id=:listID",
			func() *http.Request {
				form := url.Values{
					"title":    {"Updated Test Node"},
					"password": {"test"},
					"body":     {"Updated note body"},
				}
				r := httptest.NewRequest(http.MethodPost, "/edit.html?id="+listNodeID, strings.NewReader(form.Encode()))
				r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				return r
			}(),
			lb.editFormHandler,
			func(wr *httptest.ResponseRecorder) {
				if wr.Code != http.StatusFound {
					t.Errorf("got HTTP status code %d, expected %d", wr.Code, http.StatusFound)
				}
				redirectURI := wr.Header().Get("Location")
				wantRedirectURI := "http://www.example.com/list/" + listNodeID + "/updated-test-node.html"
				if redirectURI != wantRedirectURI {
					t.Errorf("got Redirect to %s, expected %s", redirectURI, wantRedirectURI)
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
