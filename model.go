package main

import (
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

func (m *Model) addNode(node *node.Node) (int, error) {
	return m.db.AddNode(node)
}

func (m *Model) Vote(domainID node.DomainID, vote, id int, itemID, listID node.NodeID) error {
	if err := m.db.BumpVote(domainID, itemID, vote); err != nil {
		return err
	}
	// parent holds total number of votes
	if err := m.db.BumpVote(domainID, itemID, 1); err != nil {
		return err
	}
	return nil
}

func (m *Model) editNode(node *node.Node) error {
	return m.db.EditNode(node)
}
