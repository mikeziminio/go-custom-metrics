package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration settings for the Server.
type Config struct {
	// Address is the HTTP server host:port.
	Address string `envconfig:"ADDRESS" json:"address"`
	// PprofAddress is the pprof server host:port.
	PprofAddress string `envconfig:"PPROF_ADDRESS" json:"pprof_address"`
	// StoreInterval is the file storage interval in seconds.
	StoreInterval float64 `envconfig:"STORE_INTERVAL" json:"store_interval"`
	// FileStoragePath is the path to file for storing metrics.
	FileStoragePath string `envconfig:"FILE_STORAGE_PATH" json:"store_file"`
	// Restore indicates whether to restore metrics from file on startup.
	Restore bool `envconfig:"RESTORE" json:"restore"`
	// DatabaseDSN is the database DSN for DB storage.
	DatabaseDSN string `envconfig:"DATABASE_DSN" json:"database_dsn"`
	// HashKey is the key for request signature validation.
	HashKey string `envconfig:"KEY" json:"hash_key"`
	// Path to private key file for RSA decryption (env: CRYPTO_KEY)
	CryptoKey string `envconfig:"CRYPTO_KEY" json:"crypto_key"`
	// LogLevel is the logging level (debug, info, warn, error).
	LogLevel string
	// AuditFile is the path to audit log file.
	AuditFile string `envconfig:"AUDIT_FILE" json:"audit_file"`
	// AuditURL is the URL for audit log HTTP endpoint.
	AuditURL string `envconfig:"AUDIT_URL" json:"audit_url"`
	// ConfigFile is the path to JSON config file.
	ConfigFile string
}

var (
	DefaultAddress         = "localhost:8080"
	DefaultPprofAddress    = "localhost:7070"
	DefaultStoreInterval   = 300.0
	DefaultFileStoragePath = "./data.json"
	DefaultRestore         = false
	DefaultDatabaseDSN     = ""
	DefaultHashKey         = ""
	DefaultCryptoKey       = ""
	DefaultLogLevel        = "info"
	DefaultAuditFile       = ""
	DefaultAuditURL        = ""
)

// New creates a new Config by reading environment variables
// and command-line flags.
//
// Environment variables take precedence over flags.
// Priority order: flags → env vars (env vars override flags).
//
// Returns a *Config populated with values and an error if parsing fails.
// If successful, the returned config is always non-nil.
func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}

	c.LogLevel = DefaultLogLevel

	if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
		c.ConfigFile = envConfigFile
	}

	flag.StringVar(&c.ConfigFile, "config", "", "путь к JSON файлу конфигурации")
	flag.Parse()

	if c.ConfigFile != "" {
		fileConfig, err := LoadConfigFromFile(c.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		if err := mergo.Merge(&c, fileConfig, mergo.WithOverride); err != nil {
			return nil, fmt.Errorf("failed to merge config file: %w", err)
		}
	}

	flag.StringVar(&c.Address, "a", DefaultAddress, "хост:порт http сервера")
	flag.StringVar(&c.PprofAddress, "pprof-address", DefaultPprofAddress, "хост:порт для запуска pprof сервера")
	flag.Float64Var(&c.StoreInterval, "i", DefaultStoreInterval, "интервал сохраниения в файл в секундах")
	flag.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath,
		"путь до файла, куда сохраняются текущие значения")
	flag.BoolVar(&c.Restore, "r", DefaultRestore,
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&c.DatabaseDSN, "d", DefaultDatabaseDSN, "database DSN")
	flag.StringVar(&c.HashKey, "k", DefaultHashKey, "ключ подписи")
	flag.StringVar(&c.CryptoKey, "crypto-key", DefaultCryptoKey, "путь до файла с приватным ключом")
	flag.StringVar(&c.AuditFile, "audit-file", DefaultAuditFile, "путь к файлу, в который сохраняются логи аудита")
	flag.StringVar(&c.AuditURL, "audit-url", DefaultAuditURL, "полный URL, по которой отправляются логи аудита")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}

func LoadConfigFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &c, nil
}
