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
	DatabaseDSN     string  `envconfig:"DATABASE_DSN"`
	HashKey         string  `envconfig:"KEY"`
	LogLevel        string
	AuditFile       string `envconfig:"AUDIT_FILE"`
	AuditURL        string `envconfig:"AUDIT_URL"`
}

var (
	DefaultAddress         = "localhost:8080"
	DefaultStoreInterval   = 300.0
	DefaultFileStoragePath = "./data.json"
	DefaultRestore         = false
	DefaultDatabaseDSN     = ""
	DefaultHashKey         = ""
	DefaultLogLevel        = "info"
	DefaultAuditFile       = ""
	DefaultAuditURL        = ""
)

func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}

	c.LogLevel = DefaultLogLevel

	flag.StringVar(&c.Address, "a", DefaultAddress, "хост:порт http сервера")
	flag.Float64Var(&c.StoreInterval, "i", DefaultStoreInterval, "интервал сохраниения в файл в секундах")
	flag.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath,
		"путь до файла, куда сохраняются текущие значения")
	flag.BoolVar(&c.Restore, "r", DefaultRestore,
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&c.DatabaseDSN, "d", DefaultDatabaseDSN, "database DSN")
	flag.StringVar(&c.HashKey, "k", DefaultHashKey, "ключ подписи")
	flag.StringVar(&c.AuditFile, "audit-file", DefaultAuditFile, "путь к файлу, в который сохраняются логи аудита")
	flag.StringVar(&c.AuditURL, "audit-url", DefaultAuditURL, "полный URL, по которому отправляются логи аудита")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}
