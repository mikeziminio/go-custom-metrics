package config

import (
	"flag"
	"log"

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
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatalf("failed to process envs: %v", err)
	}

	return &c
}
