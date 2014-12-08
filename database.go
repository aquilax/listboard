package main

type Database struct{}

type SiteConfig struct {
	Css string
	Title string
}

type Node struct {}

type NodeList []*Node

func NewDatabase(c *Config) *Database {
	return &Database{}
}

func (db *Database) getSiteConfig(token string) *SiteConfig {
	return &SiteConfig{}
}

func (db *Database) getChildNodes (parentNodeId, itemsPerPage, page int) *NodeList {
	nl := &NodeList{}
	return nl
}