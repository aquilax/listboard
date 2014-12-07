package main

type Database struct{}

func NewDatabase(c *Config) *Database {
	return &Database{}
}
