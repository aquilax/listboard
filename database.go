package main

import (
	"html/template"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Model struct {
	db *sqlx.DB
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

func (m *Model) getChildNodes(domainId, parentNodeId, count, offset int, orderBy string) (*NodeList, error) {
	var nl NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 AND parent_id=? ORDER BY "+orderBy+" LIMIT ?, ?", domainId, parentNodeId, offset, count)
	return &nl, err
}

func (m *Model) getAllNodes(domainId, count, offset int, orderBy string) (*NodeList, error) {
	var nl NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 ORDER BY "+orderBy+" LIMIT ?, ?", domainId, offset, count)
	return &nl, err
}

func (m *Model) mustGetChildNodes(domainId, parentNodeId, count, offset int, orderBy string) *NodeList {
	nl, err := m.getChildNodes(domainId, parentNodeId, count, offset, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) mustGetAllNodes(domainId, count, offset int, orderBy string) *NodeList {
	nl, err := m.getAllNodes(domainId, count, offset, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) getTotal(domainId, parentNodeId int) (int, error) {
	var total int
	err := m.db.Get(&total, "SELECT count(*) FROM node WHERE domain_id=$1 AND parent_id=$1 AND status=1", domainId, parentNodeId)
	return total, err
}

func (m *Model) mustGetTotal(domainId, parentNodeId int) int {
	total, err := m.getTotal(domainId, parentNodeId)
	if err != nil {
		panic(err)
	}
	return total
}

func (m *Model) getNode(domainId, listId int) (*Node, error) {
	var node Node
	err := m.db.Get(&node, "SELECT * FROM node WHERE id=$1 AND domain_id=$2 AND status=1", listId, domainId)
	return &node, err
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

func (m *Model) bumpVote(domainId, id, vote int) error {
	_, err := m.db.NamedExec(`UPDATE node set vote = vote + :vote, updated = :updated WHERE domain_id = :domain_id AND id = :id`, map[string]interface{}{
		"vote":      vote,
		"id":        id,
		"updated":   time.Now(),
		"domain_id": domainId,
	})
	return err
}

func (m *Model) Vote(domainId, vote, id, itemId, listId int) error {
	if err := m.bumpVote(domainId, itemId, vote); err != nil {
		return err
	}
	// parent holds total number of votes
	if err := m.bumpVote(domainId, listId, 1); err != nil {
		return err
	}
	return nil
}

func (m *Model) editNode(node *Node) error {
	_, err := m.db.NamedExec(`UPDATE node SET
			title = :title,
			body = :body,
			rendered = :rendered,
			updated = :updated
			WHERE id = :id
			AND domain_id = :domain_id
			AND tripcode = :tripcode`,
		map[string]interface{}{
			"title":     node.Title,
			"body":      node.Body,
			"rendered":  string(node.Rendered),
			"updated":   time.Now(),
			"id":        node.Id,
			"domain_id": node.DomainId,
			"tripcode":  node.Tripcode,
		})
	return err
}
