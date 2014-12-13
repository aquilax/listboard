package main

import (
	"encoding/json"
	"os"
)

const defaultConfigFile = "./config/listboard.json"

type Config struct {
	Server       string
	Database     string
	Dsn          string
	Translations string
	Servers      map[string]SiteConfig
}

type SiteConfig struct {
	DomainId    int
	Language    string
	Css         string
	Title       string
	Description string
	AuthorName  string
	AuthorEmail string
	PostHeader  string
	PreFooter   string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(args []string) error {
	fileName := defaultConfigFile
	if len(args) > 2 {
		fileName = args[1]
	}
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	return decoder.Decode(c)
}

func (c *Config) getSiteConfig(token string) *SiteConfig {
	var sc SiteConfig
	var ok bool
	sc, ok = c.Servers[token]
	if !ok {
		// try default
		sc, ok = c.Servers[""]
		if !ok {
			panic("No default server config found")
		}
	}
	return &sc
}
