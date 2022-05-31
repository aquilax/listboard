package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aquilax/listboard/node"
)

const ENV_CONFIG_FILE = "LB_CONFIG_FILE"
const ENV_DB_DSN = "LB_DB_DSN"
const ENV_ENVIRONMENT = "GO_ENV"

const defaultConfigFile = "./config/listboard.json"
const defaultTemplatesBase = "./templates/default/"

type Config struct {
	Server          string                 `json:"server"`
	Database        string                 `json:"database"`
	Dsn             string                 `json:"dsn"`
	Translations    string                 `json:"translations"`
	Token           string                 `json:"token"`
	PostBlockExpire string                 `json:"post_block_expire"`
	Servers         map[string]*SiteConfig `json:"servers"`
}

type SiteConfig struct {
	DomainID    node.DomainID `json:"domain_id"`
	Analytics   string        `json:"analytics"`
	Domain      string        `json:"domain"`
	Language    string        `json:"language"`
	Css         string        `json:"css"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	AuthorName  string        `json:"author_name"`
	AuthorEmail string        `json:"author_email"`
	PostHeader  string        `json:"post_header"`
	PreFooter   string        `json:"pre_footer"`
	Templates   string        `json:"templates"`
	BaseUrl     *url.URL
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(args []string) error {
	fileName := defaultConfigFile

	if os.Getenv(ENV_CONFIG_FILE) != "" {
		fileName = os.Getenv(ENV_CONFIG_FILE)
	} else if len(args) > 1 {
		fileName = args[1]
	}

	log.Printf("loading config from %s", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return err
	}
	for name := range c.Servers {
		u, err := url.Parse(c.Servers[name].Domain)
		if err != nil {
			return err
		}
		c.Servers[name].BaseUrl = u
	}
	return nil
}

func (c *Config) getSiteConfig(token string) *SiteConfig {
	var sc *SiteConfig
	var ok bool
	sc, ok = c.Servers[token]
	if !ok {
		// try default
		sc, ok = c.Servers[""]
		if !ok {
			log.Fatal("no default server config found")
		}
	}
	return sc
}

func (sc SiteConfig) templatePath(templateName string) string {
	if sc.Templates != "" {
		return sc.Templates + templateName
	}
	return defaultTemplatesBase + templateName
}

func (sc SiteConfig) getTemplateCacheKey(name string) string {
	var sb strings.Builder
	sb.WriteString(sc.Templates)
	sb.WriteString(sc.Language)
	sb.WriteString(name)
	return sb.String()
}

func (sc Config) Environment() string {
	return os.Getenv(ENV_ENVIRONMENT)
}

func (sc Config) Port() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = sc.Server
	}
	return port
}

func (sc Config) DSN() string {
	dsn := os.Getenv(ENV_DB_DSN)
	if dsn == "" {
		dsn = sc.Dsn
	}
	return dsn
}
