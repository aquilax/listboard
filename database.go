package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"strconv"
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
	Id       int       `db:"id"`
	ParentId int       `db:"parent_id"`
	DomainId int       `db:"domain_id"`
	Title    string    `db:"title"`
	Vote     int       `db:"vote"`
	Tripcode string    `db:"tripcode"`
	Body     string    `db:"body"`
	Rendered string    `db:"rendered"`
	Status   int       `db:"status"`
	Level    int       `db:"level"`
	Created  time.Time `db:"created"`
	Updated  time.Time `db:"updated"`
}

type NodeList []Node

func NewModel(c *Config) *Model {
	return &Model{}
}

func (n *Node) GetRendered() template.HTML {
	return template.HTML(n.Rendered)
}

func (n *Node) Url() string {
	switch n.Level {
	case 0:
		return "/list/" + strconv.Itoa(n.Id) + hfSlug(n.Title)
	case 1:
		return "/vote/" + strconv.Itoa(n.ParentId) + strconv.Itoa(n.Id) + "/vote.html"
	default:
		return "/list/" + strconv.Itoa(n.Id) + hfSlug(n.Title)
	}
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

func (m *Model) getChildNodes(parentNodeId, itemsPerPage, page int, orderBy string) (*NodeList, error) {
	var nl NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE status=1 AND parent_id=? ORDER BY "+orderBy+" LIMIT ?, ?", parentNodeId, page, itemsPerPage)
	return &nl, err
}

func (m *Model) getAllNodes(itemsPerPage, page int, orderBy string) (*NodeList, error) {
	var nl NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE status=1 ORDER BY "+orderBy+" LIMIT ?, ?", page, itemsPerPage)
	return &nl, err
}

func (m *Model) mustGetChildNodes(parentNodeId, itemsPerPage, page int, orderBy string) *NodeList {
	nl, err := m.getChildNodes(parentNodeId, itemsPerPage, page, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) mustGetAllNodes(itemsPerPage, page int, orderBy string) *NodeList {
	nl, err := m.getAllNodes(itemsPerPage, page, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) getNode(listId int) (*Node, error) {
	var node Node
	err := m.db.Get(&node, "SELECT * FROM node WHERE id=$1 AND status=1", listId)
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
	res, err := m.db.NamedExec(`INSERT INTO node (
			parent_id,
			domain_id,
			title,
			vote,
			tripcode,
			body,
			rendered,
			status,
			level,
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
			:level,
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
			"status":    node.Status,
			"level":     node.Level,
			"created":   time.Now(),
			"updated":   time.Now(),
		})
	if err != nil {
		return 0, err
	}
	var id int64
	id, err = res.LastInsertId()
	return int(id), err
}

func (m *Model) bumpVote(id, vote int) error {
	_, err := m.db.NamedExec(`UPDATE node set vote = vote + :vote, updated = :updated WHERE id = :id`, map[string]interface{}{
		"vote":    vote,
		"id":      id,
		"updated": time.Now(),
	})
	return err
}

func (m *Model) Vote(vote, id, itemId, listId int) error {
	if err := m.bumpVote(itemId, vote); err != nil {
		return err
	}
	// parent holds total number of votes
	if err := m.bumpVote(listId, 1); err != nil {
		return err
	}
	return nil
}
