package config

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type AgentOptions struct {
	ServerAddr     string `env:"ADDRESS,required"`
	ReportInterval int    `env:"REPORT_INTERVAL,required"`
	PollInterval   int    `env:"POLL_INTERVAL,required"`
}

type ServerOptions struct {
	ServerAddr      string `env:"ADDRESS,required"`
	StoreInterval   int    `env:"STORE_INTERVAL,required"`
	FileStoragePath string `env:"FILE_STORAGE_PATH,required"`
	RestoreOnStart  bool   `env:"RESTORE,required"`
}

func ParseAgentFlags() AgentOptions {
	var options AgentOptions

	if err := env.Parse(&options); err != nil {
		if errors.Is(err, env.ParseError{}) {
			log.Printf("can't parse agent environment variables: %s", err.Error())
		}
		if errors.Is(err, env.VarIsNotSetError{}) {
			aggErr := env.AggregateError{}
			if ok := errors.As(err, &aggErr); ok {
				for _, er := range aggErr.Errors {
					switch v := er.(type) {
					case env.VarIsNotSetError:
						setAgentFlagByName(&options, v.Key)
					default:
						fmt.Printf("Unknown error type %v", v)
					}
				}
			}
			flag.Parse()
		}
	}

	return options
}

func setAgentFlagByName(opt *AgentOptions, name string) {
	switch name {
	case "ADDRESS":
		flag.StringVar(&opt.ServerAddr, "a", "localhost:8080", "address and port server")
	case "POLL_INTERVAL":
		flag.IntVar(&opt.PollInterval, "p", 2, "interval in seconds fof getting metrics")
	case "REPORT_INTERVAL":
		flag.IntVar(&opt.ReportInterval, "r", 10, "interval in seconds for sending metrics")
	}
}

func ParseServerFlags() ServerOptions {
	var options ServerOptions
	if err := env.Parse(&options); err != nil {
		if errors.Is(err, env.ParseError{}) {
			log.Printf("can't parse agent environment variables: %s", err.Error())
		}
		if errors.Is(err, env.VarIsNotSetError{}) {
			aggErr := env.AggregateError{}
			if ok := errors.As(err, &aggErr); ok {
				for _, er := range aggErr.Errors {
					switch v := er.(type) {
					case env.VarIsNotSetError:
						setServerFlagByName(&options, v.Key)
					default:
						fmt.Printf("Unknown error type %v", v)
					}
				}
			}
			flag.Parse()
		}
	}

	return options
}

func setServerFlagByName(opt *ServerOptions, name string) {
	switch name {
	case "ADDRESS":
		flag.StringVar(&opt.ServerAddr, "a", "localhost:8080", "address and port to run server on")
	case "STORE_INTERVAL":
		flag.IntVar(&opt.StoreInterval, "i", 300, "time interval for write metrics in the storage file")
	case "FILE_STORAGE_PATH":
		flag.StringVar(&opt.FileStoragePath, "f", "C:/files/metrics.txt", "path to the storage file")
	case "RESTORE":
		flag.BoolVar(&opt.RestoreOnStart, "r", true, "need to read previous saved metrics from file?")
	}
}
