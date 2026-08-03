package config

import (
	"errors"
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type AgentOptions struct {
	ServerAddr     string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

type ServerOptions struct {
	ServerAddr      string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func ParseAgentFlags() AgentOptions {
	options := AgentOptions{
		ServerAddr:     "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
	}

	if err := env.Parse(&options); err != nil {
		if errors.Is(err, env.ParseError{}) {
			log.Printf("can't parse agent environment variables: %s", err.Error())
		}
	}

	setAgentFlag(&options)

	return options
}

func setAgentFlag(opt *AgentOptions) {
	flag.StringVar(&opt.ServerAddr, "a", opt.ServerAddr, "address and port server")
	flag.IntVar(&opt.PollInterval, "p", opt.PollInterval, "interval in seconds fof getting metrics")
	flag.IntVar(&opt.ReportInterval, "r", opt.ReportInterval, "interval in seconds for sending metrics")

	flag.Parse()
}

func ParseServerFlags() ServerOptions {
	options := ServerOptions{
		ServerAddr:      "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "", //"C:/files/metrics.txt",
		RestoreOnStart:  true,
		DatabaseDSN:     "", //"postgres://postgres:123@localhost:5432/metrics?sslmode=disable",
	}

	if err := env.Parse(&options); err != nil {
		if errors.Is(err, env.ParseError{}) {
			log.Printf("can't parse agent environment variables: %s", err.Error())
		}
	}

	setServerFlag(&options)

	return options
}

func setServerFlag(opt *ServerOptions) {
	flag.StringVar(&opt.ServerAddr, "a", opt.ServerAddr, "address and port to run server on")
	flag.IntVar(&opt.StoreInterval, "i", opt.StoreInterval, "time interval for write metrics in the storage file")
	flag.StringVar(&opt.FileStoragePath, "f", opt.FileStoragePath, "path to the storage file")
	flag.BoolVar(&opt.RestoreOnStart, "r", opt.RestoreOnStart, "need to read previous saved metrics from file?")
	flag.StringVar(&opt.DatabaseDSN, "d", opt.DatabaseDSN, "database connection string")

	flag.Parse()
}
