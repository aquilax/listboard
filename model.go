package main

import (
	"time"

	"github.com/aquilax/listboard/database"
	"github.com/aquilax/listboard/node"
)

type Model struct {
	db database.Database
}

func NewModel(db database.Database) *Model {
	return &Model{db}
}

func (m *Model) mustGetChildNodes(domainID node.DomainID, parentNodeID node.NodeID, count, offset int, orderBy string) *node.NodeList {
	nl, err := m.db.GetChildNodes(domainID, parentNodeID, count, offset, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) mustGetAllNodes(domainID node.DomainID, count, offset int, orderBy string) *node.NodeList {
	nl, err := m.db.GetAllNodes(domainID, count, offset, orderBy)
	if err != nil {
		panic(err)
	}
	return nl
}

func (m *Model) mustGetTotal(domainID node.DomainID, parentNodeID node.NodeID) int {
	total, err := m.db.GetTotalChildNodes(domainID, parentNodeID)
	if err != nil {
		panic(err)
	}
	return total
}

func (m *Model) getNode(domainID node.DomainID, nodeID node.NodeID) (*node.Node, error) {
	return m.db.GetNode(domainID, nodeID)
}

func (m *Model) addNode(node *node.Node) (node.NodeID, error) {
	node.Created = time.Now()
	node.Updated = time.Now()
	return m.db.AddNode(node)
}

func (m *Model) Vote(domainID node.DomainID, vote int, nodeID, parentID, grandParentID node.NodeID) error {
	if err := m.db.BumpVote(domainID, parentID, vote, time.Now()); err != nil {
		return err
	}
	// parent holds total number of votes
	if err := m.db.BumpVote(domainID, grandParentID, 1, time.Now()); err != nil {
		return err
	}
	return nil
}

func (m *Model) editNode(n *node.Node) error {
	n.Updated = time.Now()
	return m.db.EditNode(n)
}