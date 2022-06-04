package cached

import (
	"fmt"
	"sync"
	"time"

	"github.com/aquilax/listboard/database"
	"github.com/aquilax/listboard/node"
)

type GetTotalChildNodesCache map[string]int
type GetNodeCache map[string]*node.Node
type GetChildNodesCache map[string]*node.NodeList

type Cached struct {
	db              database.Database
	totalsCache     map[node.DomainID]GetTotalChildNodesCache
	nodeCache       map[node.DomainID]GetNodeCache
	childNodesCache map[node.DomainID]GetChildNodesCache
	locks           map[node.DomainID]*sync.Mutex
}

func New(db database.Database) *Cached {
	return &Cached{
		db:              db,
		totalsCache:     make(map[string]GetTotalChildNodesCache),
		nodeCache:       make(map[string]GetNodeCache),
		childNodesCache: make(map[node.DomainID]GetChildNodesCache),
		locks:           make(map[node.DomainID]*sync.Mutex),
	}
}

func (m Cached) clear(domainID node.DomainID) error {
	if _, found := m.locks[domainID]; !found {
		m.locks[domainID] = &sync.Mutex{}
	} else {
		m.locks[domainID].Lock()
		if _, found := m.totalsCache[domainID]; found {
			m.totalsCache[domainID] = make(GetTotalChildNodesCache)
		}
		if _, found := m.nodeCache[domainID]; found {
			m.nodeCache[domainID] = make(GetNodeCache)
		}
		if _, found := m.childNodesCache[domainID]; found {
			m.childNodesCache[domainID] = make(GetChildNodesCache)
		}
		m.locks[domainID].Unlock()
	}
	return nil
}

func (m Cached) Open(database, dsn string) error {
	return m.db.Open(database, dsn)
}

func (m Cached) GetChildNodes(domainID, parentNodeID string, count, offset int, orderBy string) (*node.NodeList, error) {
	var result *node.NodeList
	key := fmt.Sprintf("%s|%s|%d|%d%s", domainID, parentNodeID, count, offset, orderBy)

	if _, found := m.childNodesCache[domainID]; found {
		if result, found := m.childNodesCache[domainID][key]; found {
			return result, nil
		}
	} else {
		m.childNodesCache[domainID] = make(GetChildNodesCache)
	}

	result, err := m.db.GetChildNodes(domainID, parentNodeID, count, offset, orderBy)
	if err != nil {
		m.childNodesCache[domainID][key] = result
	}
	return result, err
}

func (m Cached) GetAllNodes(domainID string, count, offset int, orderBy string) (*node.NodeList, error) {
	return m.db.GetAllNodes(domainID, count, offset, orderBy)
}

func (m Cached) GetTotalChildNodes(domainID, parentNodeID string) (int, error) {
	var result int
	key := fmt.Sprintf("%s|%s", domainID, parentNodeID)
	if _, found := m.totalsCache[domainID]; found {
		if result, found := m.totalsCache[domainID][key]; found {
			return result, nil
		}
	} else {
		m.totalsCache[domainID] = make(GetTotalChildNodesCache)
	}
	result, err := m.db.GetTotalChildNodes(domainID, parentNodeID)
	if err == nil {
		m.totalsCache[domainID][key] = result
	}
	return result, err
}

func (m Cached) GetNode(domainID, nodeID string) (*node.Node, error) {
	var result *node.Node
	key := fmt.Sprintf("%s|%s", domainID, nodeID)
	if _, found := m.nodeCache[domainID]; found {
		if result, found := m.nodeCache[domainID][key]; found {
			return result, nil
		}
	} else {
		m.nodeCache[domainID] = make(GetNodeCache)
	}
	result, err := m.db.GetNode(domainID, nodeID)
	if err == nil {
		m.nodeCache[domainID][key] = result
	}
	return result, err
}

func (m *Cached) AddNode(n *node.Node) (node.NodeID, error) {
	result, err := m.db.AddNode(n)
	if err == nil {
		m.clear(n.DomainID)
	}
	return result, err
}

func (m *Cached) BumpVote(domainID node.DomainID, nodeID node.NodeID, vote int, updatedAt time.Time) error {
	err := m.db.BumpVote(domainID, nodeID, vote, updatedAt)
	if err == nil {
		m.clear(domainID)
	}
	return err
}

func (m *Cached) EditNode(n *node.Node) error {
	err := m.db.EditNode(n)
	if err == nil {
		m.clear(n.DomainID)
	}
	return err
}

func (m Cached) Close() error {
	return m.db.Close()
}
