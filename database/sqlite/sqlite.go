package sqlite

import (
	"time"

	"github.com/aquilax/listboard/node"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sqlx.DB
}

func New() *SQLite {
	return &SQLite{}
}

func (m *SQLite) Init(database, DSN string) error {
	var err error
	m.db, err = sqlx.Open(database, DSN)
	return err
}

func (m *SQLite) GetChildNodes(domainID node.DomainID, parentNodeID node.NodeID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 AND parent_id=? ORDER BY "+orderBy+" LIMIT ?, ?", domainID, parentNodeID, offset, count)
	return &nl, err
}

func (m *SQLite) GetAllNodes(domainID node.DomainID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 ORDER BY "+orderBy+" LIMIT ?, ?", domainID, offset, count)
	return &nl, err
}

func (m *SQLite) GetTotalChildNodes(domainID node.DomainID, parentNodeID node.NodeID) (int, error) {
	var total int
	err := m.db.Get(&total, "SELECT count(*) FROM node WHERE domain_id=$1 AND parent_id=$1 AND status=1", domainID, parentNodeID)
	return total, err
}

func (m *SQLite) GetNode(domainID node.DomainID, listID node.NodeID) (*node.Node, error) {
	var node node.Node
	err := m.db.Get(&node, "SELECT * FROM node WHERE id=$1 AND domain_id=$2 AND status=1", listID, domainID)
	return &node, err
}

func (m *SQLite) AddNode(node *node.Node) (int, error) {
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
			"parent_id": node.ParentID,
			"domain_id": node.DomainID,
			"title":     node.Title,
			"vote":      node.Vote,
			"tripcode":  node.TripCode,
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

func (m *SQLite) BumpVote(domainID node.DomainID, id node.NodeID, vote int) error {
	_, err := m.db.NamedExec(`UPDATE node set vote = vote + :vote, updated = :updated WHERE domain_id = :domain_id AND id = :id`, map[string]interface{}{
		"vote":      vote,
		"id":        id,
		"updated":   time.Now(),
		"domain_id": domainID,
	})
	return err
}

func (m *SQLite) EditNode(node *node.Node) error {
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
			"id":        node.ID,
			"domain_id": node.DomainID,
			"tripcode":  node.TripCode,
		})
	return err
}
