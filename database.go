package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"time"
)

type Model struct {
	db *sqlx.DB
}

type SiteConfig struct {
	DomainId    int
	Css         string
	Title       string
	Description string
	AuthorName  string
	AuthorEmail string
}

type Node struct {
	Id           int    `db:"id"`
	ParentId     int    `db:"parent_id"`
	DomainId     int    `db:"domain_id"`
	Title        string `db:"title"`
	Vote         int    `db:"vote"`
	Tripcode     string `db:"tripcode"`
	Body         string `db:"body"`
	Rendered     string `db:"rendered"`
	RenderedHTML template.HTML
	Status       int       `db:"status"`
	Created      time.Time `db:"created"`
	Updated      time.Time `db:"updated"`
}

type NodeList []*Node

func NewModel(c *Config) *Model {
	return &Model{}
}

func (m *Model) Init(config *Config) error {
	var err error
	m.db, err = sqlx.Open(config.Database, config.Dsn)
	return err
}

func (m *Model) getSiteConfig(token string) *SiteConfig {
	return &SiteConfig{
		DomainId: 0,
		Css:      "style.css",
	}
}

func (m *Model) getChildNodes(parentNodeId, itemsPerPage, page int, orderBy string) *NodeList {
	nl := &NodeList{}
	return nl
}

func (m *Model) getNode(listId int) (*Node, error) {
	var node Node
	err := m.db.Get(&node, "SELECT * FROM node WHERE id=$1", listId)
	return &node, err
}

func (m *Model) mustGetNode(listId int) *Node {
	node, err := m.getNode(listId)
	if err != nil {
		panic(err)
	}
	return node
}

func (m *Model) addNode(node *Node) (int, error) {
	_, err := m.db.NamedExec(`INSERT INTO node (
			parent_id,
			domain_id,
			title,
			vote,
			tripcode,
			body,
			rendered,
			status,
			created,
			updated
		) VALUES (
			:parent_id,
			:domain_id,
			:title,
			:vote,
			:tripcode,
			:body,
			:rendered,
			:status,
			:created,
			:updated
  		)`,
		map[string]interface{}{
			"parent_id": node.ParentId,
			"domain_id": node.DomainId,
			"title":     node.Title,
			"vote":      node.Vote,
			"tripcode":  node.Tripcode,
			"body":      node.Body,
			"rendered":  string(node.Rendered),
			"status":    1,
			"created":   time.Now(),
			"updated":   time.Now(),
		})

	return 0, err
}
