package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address        string  `envconfig:"ADDRESS"`
	ReportInterval float64 `envconfig:"REPORT_INTERVAL"`
	PollInterval   float64 `envconfig:"POLL_INTERVAL"`
	LogLevel       string
	UseCompress    bool
}

var (
	DefaultPollInterval   = 2.0
	DefaultReportInterval = 10.0
	DefaultUseCompress    = true
	DefaultLogLevel       = "info"
)

func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}

	c.UseCompress = DefaultUseCompress
	c.LogLevel = DefaultLogLevel

	flag.StringVar(&c.Address, "a", "localhost:8080", "хост:порт http сервера")
	flag.Float64Var(
		&c.ReportInterval,
		"r",
		DefaultReportInterval,
		"частота отправки метрик на сервер",
	)
	flag.Float64Var(&c.PollInterval, "p", DefaultPollInterval, "частота опроса метрик")
	flag.Parse()

	// по ТЗ переменные среды перезаписывают флаги
	// хоть это и не логично - c т.з. пользовательского опыта должно быть наоборот :)
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}
