package main

type Database struct{}

type Node struct {

}

type NodeList []*Node

func NewDatabase(c *Config) *Database {
	return &Database{}
}

func (d *Database) getChildNodes (parentNodeId, itemsPerPage, page int) *NodeList {
	nl := &NodeList{}
	return nl
}