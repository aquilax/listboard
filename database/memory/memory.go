package memory

import (
	"fmt"
	"time"

	"github.com/aquilax/listboard/node"
	"github.com/google/uuid"
)

type Memory struct {
	nl node.NodeList
}

func New() *Memory {
	return &Memory{}
}

func min(value int, values ...int) int {
	for _, v := range values {
		if v < value {
			value = v
		}
	}
	return value
}

func find(nl node.NodeList, filter func(n node.Node) bool) node.NodeList {
	var result node.NodeList
	for _, n := range nl {
		if filter(n) {
			result = append(result, n)
		}
	}
	return result
}

func (m Memory) Open(database, dsn string) error {
	return nil
}

func (m Memory) GetChildNodes(domainID, parentNodeID string, count, offset int, orderBy string) (*node.NodeList, error) {
	found := find(m.nl, func(n node.Node) bool {
		return n.DomainID == domainID && n.ParentID == parentNodeID
	})
	result := found[offset:min(len(found), offset+count)]
	return &result, nil
}

func (m Memory) GetAllNodes(domainID string, count, offset int, orderBy string) (*node.NodeList, error) {
	found := find(m.nl, func(n node.Node) bool {
		return n.DomainID == domainID
	})
	// TODO order
	result := found[offset:min(len(found), offset+count)]
	return &result, nil
}

func (m Memory) GetTotalChildNodes(domainID, parentNodeID string) (int, error) {
	found := find(m.nl, func(n node.Node) bool {
		return n.DomainID == domainID && n.ParentID == parentNodeID
	})
	return len(found), nil
}

func (m Memory) GetNode(domainID, nodeID string) (*node.Node, error) {
	found := find(m.nl, func(n node.Node) bool {
		return n.DomainID == domainID && n.ID == nodeID
	})
	if len(found) > 0 {
		return &found[0], nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *Memory) AddNode(node *node.Node) (node.NodeID, error) {
	id := uuid.New()
	node.ID = id.String()
	m.nl = append(m.nl, *node)
	return node.ID, nil
}

func (m *Memory) BumpVote(domainID node.DomainID, nodeID node.NodeID, vote int, updatedAt time.Time) error {
	for i := range m.nl {
		if m.nl[i].DomainID == domainID && m.nl[i].ID == nodeID {
			m.nl[i].Vote = m.nl[i].Vote + vote
			m.nl[i].Updated = updatedAt
			return nil
		}
	}
	return nil
}

func (m *Memory) EditNode(n *node.Node) error {
	for i := range m.nl {
		if m.nl[i].ID == n.ID && m.nl[i].TripCode == n.TripCode {
			m.nl[i] = *n
			return nil
		}
	}
	return nil
}

func (m Memory) Close() error {
	return nil
}
