package config

import (
	"flag"
)

type Config struct {
	Address string
}

func NewFromFlags() (*Config, error) {
	c := Config{}
	flag.StringVar(&c.Address, "a", "localhost:8080", "хост:порт http сервера")
	flag.Parse()

	return &c, nil
}
