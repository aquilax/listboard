package main

import (
	"time"
)

type Database struct{}

type SiteConfig struct {
	Css   string
	Title string
}

type Node struct {
	Id      int
	Title   string
	Vote    int
	Updated time.Time
}

type NodeList []*Node

func NewDatabase(c *Config) *Database {
	return &Database{}
}

func (db *Database) getSiteConfig(token string) *SiteConfig {
	return &SiteConfig{}
}

func (db *Database) getChildNodes(parentNodeId, itemsPerPage, page int, orderBy string) *NodeList {
	nl := &NodeList{
		&Node{
			Id:      1,
			Title:   "Test item",
			Vote:    3,
			Updated: time.Now(),
		},
	}
	return nl
}

func (db *Database) getList(listId int) *Node {
	return &Node{
		Id:      1,
		Title:   "Test item",
		Vote:    3,
		Updated: time.Now(),
	}
}
