package postgres

import (
	"strconv"
	"time"

	"github.com/aquilax/listboard/node"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sqlx.DB
}

func New() *Postgres {
	return &Postgres{}
}

func (m *Postgres) Open(database, DSN string) error {
	var err error
	m.db, err = sqlx.Open(database, DSN)
	if err != nil {
		return err
	}
	return m.db.Ping()
}

func (m Postgres) GetChildNodes(domainID node.DomainID, parentNodeID node.NodeID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id = $1 AND status=1 AND parent_id = $2 ORDER BY "+orderBy+" LIMIT $4 OFFSET $3", domainID, parentNodeID, offset, count)
	return &nl, err
}

func (m Postgres) GetAllNodes(domainID node.DomainID, count, offset int, orderBy string) (*node.NodeList, error) {
	var nl node.NodeList
	err := m.db.Select(&nl, "SELECT * FROM node WHERE domain_id=$1 AND status=1 ORDER BY "+orderBy+" LIMIT $3 OFFSET $2", domainID, offset, count)
	return &nl, err
}

func (m Postgres) GetTotalChildNodes(domainID node.DomainID, parentNodeID node.NodeID) (int, error) {
	var total int
	err := m.db.Get(&total, "SELECT count(*) FROM node WHERE domain_id=$1 AND parent_id=$2 AND status=1", domainID, parentNodeID)
	return total, err
}

func (m Postgres) GetNode(domainID node.DomainID, listID node.NodeID) (*node.Node, error) {
	var n node.Node
	err := m.db.Get(&n, "SELECT * FROM node WHERE id=$1 AND domain_id=$2 AND status=1", listID, domainID)
	return &n, err
}

func (m Postgres) AddNode(n *node.Node) (node.NodeID, error) {
	var id int
	var err error
	tx := m.db.MustBegin()
	if err != nil {
		return node.RootNodeID, err
	}
	err = m.db.Get(&id, "SELECT max(id) FROM node")
	if err != nil {
		return node.RootNodeID, err
	}
	newId := id + 1
	_, err = tx.NamedExec(`INSERT INTO node (
			id,
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
			:id,
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
			"id":        newId,
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
		return node.RootNodeID, err
	}
	err = tx.Commit()
	if err != nil {
		return node.RootNodeID, err
	}
	return node.NodeID(strconv.Itoa(id)), err
}

func (m Postgres) BumpVote(domainID node.DomainID, id node.NodeID, vote int, updatedAt time.Time) error {
	_, err := m.db.NamedExec(`UPDATE node set vote = vote + :vote, updated = :updated WHERE domain_id = :domain_id AND id = :id`, map[string]interface{}{
		"vote":      vote,
		"id":        id,
		"updated":   updatedAt,
		"domain_id": domainID,
	})
	return err
}

func (m Postgres) EditNode(n *node.Node) error {
	_, err := m.db.NamedExec(`UPDATE node SET
			title = :title,
			body = :body,
			rendered = :rendered,
			updated = :updated
			WHERE id = :id
			AND domain_id = :domain_id
			AND tripcode = :tripcode`,
		map[string]interface{}{
			"title":     n.Title,
			"body":      n.Body,
			"rendered":  string(n.Rendered),
			"updated":   n.Updated,
			"id":        n.ID,
			"domain_id": n.DomainID,
			"tripcode":  n.TripCode,
		})
	return err
}

func (m Postgres) Close() error {
	return m.db.Close()
}
