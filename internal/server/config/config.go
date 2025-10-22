package config

import (
	"flag"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address string `envconfig:"ADDRESS"`
}

func NewFromFlags() *Config {
	c := Config{}
	flag.StringVar(&c.Address, "a", "localhost:8080", "хост:порт http сервера")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	envconfig.Process("", &c)

	return &c
}
