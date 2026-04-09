package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/mikeziminio/go-custom-metrics/internal/config"
)

// Config holds the configuration settings for the Server.
type Config struct {
	// хост:порт http сервера
	Address string `envconfig:"ADDRESS" json:"address" flag:"a" doc:"хост:порт http сервера"`
	// хост:порт для запуска pprof сервера
	PprofAddress string `envconfig:"PPROF_ADDRESS" json:"pprof_address" flag:"pprof-address" doc:"хост:порт для запуска pprof сервера"`
	// интервал сохраниения в файл в секундах
	StoreInterval float64 `envconfig:"STORE_INTERVAL" json:"store_interval" flag:"i" doc:"интервал сохраниения в файл в секундах"`
	// путь до файла, куда сохраняются текущие значения
	FileStoragePath string `envconfig:"FILE_STORAGE_PATH" json:"store_file" flag:"f" doc:"путь до файла, куда сохраняются текущие значения"`
	// следует ли загружать ранее сохранённые значения из указанного файла при старте сервера
	Restore bool `envconfig:"RESTORE" json:"restore" flag:"r" doc:"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера"`
	// database DSN
	DatabaseDSN string `envconfig:"DATABASE_DSN" json:"database_dsn" flag:"d" doc:"database DSN"`
	// ключ подписи
	HashKey string `envconfig:"KEY" json:"hash_key" flag:"k" doc:"ключ подписи"`
	// путь до файла с приватным ключом
	CryptoKey string `envconfig:"CRYPTO_KEY" json:"crypto_key" flag:"crypto-key" doc:"путь до файла с приватным ключом"`
	// путь к файлу, в который сохраняются логи аудита
	AuditFile string `envconfig:"AUDIT_FILE" json:"audit_file" flag:"audit-file" doc:"путь к файлу, в который сохраняются логи аудита"`
	// полный URL, на который отправляются логи аудита
	AuditURL string `envconfig:"AUDIT_URL" json:"audit_url" flag:"audit-url" doc:"полный URL, на который отправляются логи аудита"`
	// доверенная подсеть (CIDR)
	TrustedSubnet string `envconfig:"TRUSTED_SUBNET" json:"trusted_subnet" flag:"t" doc:"доверенная подсеть (CIDR)"`
	// gRPC адрес сервера
	GrpcAddress string `envconfig:"GRPC_ADDRESS" json:"grpc_address" flag:"grpc-addr" doc:"gRPC адрес сервера"`
	LogLevel    string
}

var defaultConfig = Config{
	Address:         "localhost:8080",
	PprofAddress:    "localhost:7070",
	StoreInterval:   300.0,
	FileStoragePath: "./data.json",
	Restore:         false,
	LogLevel:        "info",
}

func newFromDefault() *Config {
	c := defaultConfig
	return &c
}

// NewFromEnvsAndFlags creates a new Config by reading environment variables
// and command-line flags.
//
// Priority order:
// json config -> flags -> env vars (env vars override all).
//
// Returns a *Config populated with values and an error if parsing fails.
// If successful, the returned config is always non-nil.
func NewFromEnvsAndFlags() (*Config, error) {
	var c = newFromDefault()
	var configFile string

	flag.StringVar(&configFile, "config", "", "путь к JSON файлу конфигурации")

	err := config.FillFlags(c, &defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fill flags: %w", err)
	}
	flag.Parse()

	if envConfigFile, ok := os.LookupEnv("CONFIG"); ok {
		configFile = envConfigFile
	}

	if configFile != "" {
		var cf = newFromDefault()
		err := config.FillConfigFromFile(cf, &defaultConfig, configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		flags := make(map[string]struct{})
		flag.Visit(func(f *flag.Flag) {
			flags[f.Name] = struct{}{}
		})
		config.MergeOnlyFlags(cf, c, flags)
		*c = *cf
	}

	err = envconfig.Process("", c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return c, nil
}
