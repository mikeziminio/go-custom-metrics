package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration settings for the Agent.
//
// Values can be set via environment variables (prefixed with AGENT_) or
// command-line flags. Environment variables take precedence over flags.
type Config struct {
	Address        string  `envconfig:"ADDRESS"`
	ReportInterval float64 `envconfig:"REPORT_INTERVAL"`
	PollInterval   float64 `envconfig:"POLL_INTERVAL"`
	HashKey        string  `envconfig:"KEY"`
	RateLimit      int     `envconfig:"RATE_LIMIT"`
	LogLevel       string
	UseCompress    bool
	Timeout        float64
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
)

// New creates a new Config by reading environment variables
// and command-line flags.
//
// Environment variables take precedence over flags. Valid environment variable
// prefixes are: ADDRESS, REPORT_INTERVAL, POLL_INTERVAL, KEY, RATE_LIMIT.
//
// Returns a *Config and an error if parsing fails.
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
