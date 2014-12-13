package main

type Config struct {
	Server   string
	Database string
	Dsn      string
}

func NewConfig() *Config {
	return &Config{
		Server:   ":8080",
		Database: "sqlite3",
		Dsn:      "./db/test.sqlite",
	}
}
