package main

import (
	"html/template"
	"time"
)

type Database struct{}

type SiteConfig struct {
	Css         string
	Title       string
	Description string
	AuthorName  string
	AuthorEmail string
}

type Node struct {
	Id       int
	ParentId int
	DomainId int
	Title    string
	Vote     int
	Tripcode string
	Body     string
	Rendered template.HTML
	Status   int
	Updated  time.Time
	Created  time.Time
}

type NodeList []*Node

func NewDatabase(c *Config) *Database {
	return &Database{}
}

func (db *Database) getSiteConfig(token string) *SiteConfig {
	return &SiteConfig{
		Css: "style.css",
	}
}

func (db *Database) getChildNodes(parentNodeId, itemsPerPage, page int, orderBy string) *NodeList {
	nl := &NodeList{
		db.getNode(0),
		db.getNode(1),
	}
	return nl
}

func (db *Database) getNode(listId int) *Node {
	return &Node{
		Id:       listId,
		Title:    "Test item",
		Tripcode: "triptitrip",
		Vote:     3,
		Rendered: "<p>Rendered</p>",
		Updated:  time.Now(),
		Created:  time.Now(),
	}
}

func (db *Database) addNode(node *Node) (int, error) {
	return 0, nil
}
