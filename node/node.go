package node

import (
	"html/template"
	"time"
)

type NodeID = string
type DomainID = string

const RootNodeID NodeID = "0"

type Node struct {
	ID       NodeID    `db:"id"`
	ParentID NodeID    `db:"parent_id"`
	DomainID DomainID  `db:"domain_id"`
	Title    string    `db:"title"`
	Vote     int       `db:"vote"`
	TripCode string    `db:"tripcode"`
	Body     string    `db:"body"`
	Rendered string    `db:"rendered"`
	Status   int       `db:"status"`
	Level    int       `db:"level"`
	Created  time.Time `db:"created"`
	Updated  time.Time `db:"updated"`
}

type NodeList []Node

func (n *Node) GetRendered() template.HTML {
	return template.HTML(n.Rendered)
}

func (n *Node) Url() string {
	switch n.Level {
	case 0:
		return "/list/" + n.ID + hfSlug(n.Title)
	case 1:
		return "/vote/" + n.ParentID + n.ID + "/vote.html"
	default:
		return "/list/" + n.ID + hfSlug(n.Title)
	}
}

func hfSlug(t string) string {
	// TODO: FIXME
	return t
}
