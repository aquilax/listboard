package database

import "github.com/aquilax/listboard/node"

type Database interface {
	Init(database, dsn string) error
	GetChildNodes(domainID, parentNodeID string, count, offset int, orderBy string) (*node.NodeList, error)
	GetAllNodes(domainID string, count, offset int, orderBy string) (*node.NodeList, error)
	GetTotalChildNodes(domainID, parentNodeID string) (int, error)
	GetNode(domainID, nodeID string) (*node.Node, error)
	AddNode(node *node.Node) (int, error)
	BumpVote(domainID node.DomainID, id node.NodeID, vote int) error
	EditNode(node *node.Node) error
}
