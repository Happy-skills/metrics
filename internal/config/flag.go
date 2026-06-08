package config

import "flag"

type agentOptions struct {
	ServerAddr     string
	ReportInterval int
	PollInterval   int
}

type serverOptions struct {
	ServerAddr string
}

var AgentOptions agentOptions
var ServerOptions serverOptions

func ParseAgentFlags() {
	flag.StringVar(&AgentOptions.ServerAddr, "a", "localhost:8081", "address and port server")
	flag.IntVar(&AgentOptions.ReportInterval, "r", 10, "interval in seconds for sending metrics")
	flag.IntVar(&AgentOptions.PollInterval, "p", 2, "interval in seconds fof getting metrics")
	flag.Parse()
}

func ParseServerFlags() {
	flag.StringVar(&ServerOptions.ServerAddr, "a", "localhost:8080", "address and port to run server on")
	flag.Parse()
}
