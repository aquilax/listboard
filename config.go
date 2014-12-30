package main

import (
	"encoding/json"
	"log"
	"os"
)

const defaultConfigFile = "./config/listboard.json"
const defaultTemplatesBase = "./templates/default/"

type Config struct {
	Server          string                `json:"server"`
	Database        string                `json:"database"`
	Dsn             string                `json:"dsn"`
	Translations    string                `json:"translations"`
	Token           string                `json:"token"`
	PostBlockExpire string                `json:"post_block_expire"`
	Servers         map[string]SiteConfig `json:"servers"`
}

type SiteConfig struct {
	DomainId    int    `json:"domain_id"`
	Analytics   string `json:"analytics"`
	Domain      string `json:"domain"`
	Language    string `json:"language"`
	Css         string `json:"css"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	PostHeader  string `json:"post_header"`
	PreFooter   string `json:"pre_footer"`
	Templates   string `json:"templates"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(args []string) error {
	fileName := defaultConfigFile
	if len(args) > 1 {
		fileName = args[1]
	}
	log.Printf("Loading config from %s", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return err
	}
	return nil
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

func (sc *SiteConfig) templatePath(templateName string) string {
	if sc.Templates != "" {
		return sc.Templates + templateName
	}
	return defaultTemplatesBase + templateName
}
