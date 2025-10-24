package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address         string  `envconfig:"ADDRESS"`
	StoreInterval   float64 `envconfig:"STORE_INTERVAL"`
	FileStoragePath string  `envconfig:"FILE_STORAGE_PATH"`
	Restore         bool    `envconfig:"RESTORE"`
	LogLevel        string
}

var (
	DefaultLogLevel        = "info"
	DefaultStoreInterval   = 300.0
	DefaultFileStoragePath = "./data.json"
)

func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}

	c.LogLevel = DefaultLogLevel

	flag.StringVar(&c.Address, "a", "localhost:8080", "хост:порт http сервера")
	flag.Float64Var(&c.StoreInterval, "i", DefaultStoreInterval,
		"интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath,
		"путь до файла, куда сохраняются текущие значения")
	flag.BoolVar(&c.Restore, "r", false,
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}
