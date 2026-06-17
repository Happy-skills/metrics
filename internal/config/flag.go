package config

import "flag"

type AgentOptions struct {
	ServerAddr     string
	ReportInterval int
	PollInterval   int
}

type ServerOptions struct {
	ServerAddr string
}

func ParseAgentFlags() AgentOptions {
	var options AgentOptions
	flag.StringVar(&options.ServerAddr, "a", "localhost:8080", "address and port server")
	flag.IntVar(&options.ReportInterval, "r", 10, "interval in seconds for sending metrics")
	flag.IntVar(&options.PollInterval, "p", 2, "interval in seconds fof getting metrics")
	flag.Parse()
	return options
}

func ParseServerFlags() ServerOptions {
	var options ServerOptions
	flag.StringVar(&options.ServerAddr, "a", "localhost:8080", "address and port to run server on")
	flag.Parse()
	return options
}
