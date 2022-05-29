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

	db, _ := getDatabaseAdapter(lb.config.Database)
	db.Open(lb.config.Database, lb.config.Dsn)
	lb.m = NewModel(db)
	listNodeID, _ := lb.m.addNode(&node.Node{
		ID:       "1",
		ParentID: node.RootNodeID,
		Title:    "Test Node",
		DomainID: "1",
		TripCode: getTripCode("test"),
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
			"updating list works",
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
