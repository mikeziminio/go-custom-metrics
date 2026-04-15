package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/mikeziminio/go-custom-metrics/internal/config"
)

// Config holds configuration for the agent.
type Config struct {
	// хост:порт http сервера
	Address string `envconfig:"ADDRESS" json:"address" flag:"a" doc:"хост:порт http сервера"`
	// Частота отправки метрик на сервер
	ReportInterval float64 `envconfig:"REPORT_INTERVAL" json:"report_interval" flag:"r" doc:"Частота отправки метрик на сервер"`
	// Частота опроса метрик
	PollInterval float64 `envconfig:"POLL_INTERVAL" json:"poll_interval" flag:"p" doc:"Частота опроса метрик"`
	// Ключ подписи
	HashKey string `envconfig:"KEY" json:"hash_key" flag:"k" doc:"Ключ подписи"`
	// Путь до файла с публичным ключом
	CryptoKey string `envconfig:"CRYPTO_KEY" json:"crypto_key" flag:"crypto-key" doc:"Путь до файла с публичным ключом"`
	// Количество одновременных запросов к серверу
	RateLimit int `envconfig:"RATE_LIMIT" json:"rate_limit" flag:"l" doc:"Количество одновременных запросов к серверу"`
	// Уровень логирования
	LogLevel string
	// Использовать ли компрессию gzip
	UseCompress bool
	// Таймаут запросов
	Timeout float64
	// gRPC адрес сервера
	GrpcAddress string `envconfig:"GRPC_ADDRESS" json:"grpc_address" flag:"grpc-addr" doc:"gRPC адрес сервера"`
	// Использовать ли gRPC
	UseGrpc bool `envconfig:"USE_GRPC" json:"use_grpc" flag:"use-grpc" doc:"использовать gRPC"`
}

var defaultConfig = Config{
	Address:        "localhost:8080",
	ReportInterval: 10.0,
	PollInterval:   2.0,
	RateLimit:      100,
	LogLevel:       "info",
	UseCompress:    true,
	Timeout:        1.0,
	UseGrpc:        false,
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
