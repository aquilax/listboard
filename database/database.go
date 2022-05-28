package database

import (
	"time"

	"github.com/aquilax/listboard/node"
)

type Database interface {
	Init(database, dsn string) error
	GetChildNodes(domainID, parentNodeID string, count, offset int, orderBy string) (*node.NodeList, error)
	GetAllNodes(domainID string, count, offset int, orderBy string) (*node.NodeList, error)
	GetTotalChildNodes(domainID, parentNodeID string) (int, error)
	GetNode(domainID, nodeID string) (*node.Node, error)
	AddNode(node *node.Node) (node.NodeID, error)
	BumpVote(domainID node.DomainID, id node.NodeID, vote int, updatedAt time.Time) error
	EditNode(node *node.Node) error
	Close() error
}
