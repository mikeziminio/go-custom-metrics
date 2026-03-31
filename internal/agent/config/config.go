package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration for the agent.
type Config struct {
	// HTTP server host:port (env: ADDRESS)
	Address string `envconfig:"ADDRESS"`
	// Metric reporting frequency to server (env: REPORT_INTERVAL)
	ReportInterval float64 `envconfig:"REPORT_INTERVAL"`
	// Metric polling frequency (env: POLL_INTERVAL)
	PollInterval float64 `envconfig:"POLL_INTERVAL"`
	// Signature key (env: KEY)
	HashKey string `envconfig:"KEY"`
	// Path to public key file for RSA encryption (env: CRYPTO_KEY)
	CryptoKey string `envconfig:"CRYPTO_KEY"`
	// Maximum concurrent requests to server (env: RATE_LIMIT)
	RateLimit int `envconfig:"RATE_LIMIT"`
	// LogLevel is the logging level (debug, info, warn, error).
	LogLevel string
	// UseCompress enables GZIP compression
	UseCompress bool
	// Request timeout in seconds
	Timeout float64
}

var (
	DefaultAddress        = "localhost:8080"
	DefaultPollInterval   = 2.0
	DefaultReportInterval = 10.0
	DefaultUseCompress    = true
	DefaultRateLimit      = 100
	DefaultLogLevel       = "info"
	DefaultTimeout        = 1.0
	DefaultHashKey        = ""
	DefaultCryptoKey      = ""
)

// NewFromEnvsAndFlags creates a new Config by reading environment variables
// and command-line flags.
//
// Environment variables take precedence over flags.
// Priority order: flags → env vars (env vars override flags).
//
// Returns a *Config populated with values and an error if parsing fails.
// If successful, the returned config is always non-nil.
func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}

	c.UseCompress = DefaultUseCompress
	c.LogLevel = DefaultLogLevel
	c.Timeout = DefaultTimeout

	flag.StringVar(&c.Address, "a", DefaultAddress, "хост:порт http сервера")
	flag.Float64Var(
		&c.ReportInterval,
		"r",
		DefaultReportInterval,
		"частота отправки метрик на сервер",
	)
	flag.Float64Var(&c.PollInterval, "p", DefaultPollInterval, "частота опроса метрик")
	flag.StringVar(&c.HashKey, "k", DefaultHashKey, "ключ подписи")
	flag.StringVar(&c.CryptoKey, "crypto-key", DefaultCryptoKey, "путь до файла с публичным ключом")
	flag.IntVar(&c.RateLimit, "l", DefaultRateLimit, "количество одновременных запросов к серверу")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}
