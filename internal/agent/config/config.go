package config

import (
	"flag"
)

type Config struct {
	Address            string
	ReportInterval     float64
	PollInterval       float64
	ConcurrentRequests int
}

func NewFromFlags() *Config {
	c := Config{}

	// todo: next sprints
	// видимо в следующих спринтах будет расширение конфигов (через env)
	// сейчас те что не задаются через флаги - просто хардкодятся
	c.ConcurrentRequests = 10

	flag.StringVar(&c.Address, "a", "localhost:8080", "хост:порт http сервера")
	flag.Float64Var(
		&c.ReportInterval,
		"r",
		10.0,
		"частота отправки метрик на сервер",
	)
	flag.Float64Var(&c.PollInterval, "p", 2.0, "частота опроса метрик")
	flag.Parse()

	return &c
}
