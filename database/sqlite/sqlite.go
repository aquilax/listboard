package sqlite

import (
	"strconv"
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

func (m *SQLite) Open(database, DSN string) error {
	var err error
	m.db, err = sqlx.Open(database, DSN)
	if err != nil {
		return err
	}
	return m.db.Ping()
}

func (m SQLite) GetChildNodes(domainID node.DomainID, parentNodeID node.NodeID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 AND parent_id=? ORDER BY "+orderBy+" LIMIT ?, ?", domainID, parentNodeID, offset, count)
	return &nl, err
}

func (m SQLite) GetAllNodes(domainID node.DomainID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = ? AND status=1 ORDER BY "+orderBy+" LIMIT ?, ?", domainID, offset, count)
	return &nl, err
}

func (m SQLite) GetTotalChildNodes(domainID node.DomainID, parentNodeID node.NodeID) (int, error) {
	var total int
	err := m.db.Get(&total, "SELECT count(*) FROM node WHERE domain_id=$1 AND parent_id=$1 AND status=1", domainID, parentNodeID)
	return total, err
}

func (m SQLite) GetNode(domainID node.DomainID, listID node.NodeID) (*node.Node, error) {
	var node node.Node
	err := m.db.Get(&node, "SELECT * FROM node WHERE id=$1 AND domain_id=$2 AND status=1", listID, domainID)
	return &node, err
}

func (m SQLite) AddNode(n *node.Node) (node.NodeID, error) {
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
			"parent_id": n.ParentID,
			"domain_id": n.DomainID,
			"title":     n.Title,
			"vote":      n.Vote,
			"tripcode":  n.TripCode,
			"body":      n.Body,
			"rendered":  string(n.Rendered),
			"status":    n.Status,
			"level":     n.Level,
			"created":   n.Created,
			"updated":   n.Updated,
		})
	if err != nil {
		return "", err
	}
	var id int64
	id, err = res.LastInsertId()
	return node.NodeID(strconv.Itoa(int(id))), err
}

func (m SQLite) BumpVote(domainID node.DomainID, id node.NodeID, vote int, updatedAt time.Time) error {
	_, err := m.db.NamedExec(`UPDATE node set vote = vote + :vote, updated = :updated WHERE domain_id = :domain_id AND id = :id`, map[string]interface{}{
		"vote":      vote,
		"id":        id,
		"updated":   updatedAt,
		"domain_id": domainID,
	})
	return err
}

func (m SQLite) EditNode(node *node.Node) error {
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
			"updated":   node.Updated,
			"id":        node.ID,
			"domain_id": node.DomainID,
			"tripcode":  node.TripCode,
		})
	return err
}

func (m SQLite) Close() error {
	return m.db.Close()
}
