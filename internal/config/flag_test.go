package config

import (
	"reflect"
	"testing"
)

func TestParseAgentFlags(t *testing.T) {
	tests := []struct {
		name string
		want AgentOptions
	}{
		{
			name: "default values",
			want: AgentOptions{
				ServerAddr:     "localhost:8080",
				ReportInterval: 10,
				PollInterval:   2,
			},
		},
	}
	for _, tt := range tests {
		got := ParseAgentFlags()
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAgentFlags() = %v, want %v", got, tt.want)
			}
		})

	}
}

func TestParseServerFlags(t *testing.T) {
	tests := []struct {
		name string
		want ServerOptions
	}{
		{
			name: "default values server options",
			want: ServerOptions{
				ServerAddr:      "localhost:8080",
				StoreInterval:   300,
				RestoreOnStart:  true,
				FileStoragePath: "C:/files/metrics.txt",
				DatabaseDSN:     "postgres://postgres:123@localhost:5432/metrics",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServerFlags()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseServerFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setAgentFlag(t *testing.T) {
	type args struct {
		opt  *AgentOptions
		name string
	}
	tests := []struct {
		name string
		args args
		want AgentOptions
	}{
		{
			name: "set ServerAddr",
			args: args{
				opt:  &AgentOptions{},
				name: "ADDRESS",
			},
			want: AgentOptions{
				ServerAddr: "localhost:8080",
			},
		},
		{
			name: "set ReportInterval",
			args: args{
				opt:  &AgentOptions{},
				name: "REPORT_INTERVAL",
			},
			want: AgentOptions{
				ReportInterval: 10,
			},
		},
		{
			name: "set PollInterval",
			args: args{
				opt:  &AgentOptions{},
				name: "POLL_INTERVAL",
			},
			want: AgentOptions{
				PollInterval: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAgentFlag(tt.args.opt)
			if tt.args.opt.ServerAddr != tt.want.ServerAddr {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.ServerAddr, tt.want.ServerAddr)
			}
			if tt.args.opt.ReportInterval != tt.want.ReportInterval {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.ReportInterval, tt.want.ReportInterval)
			}
			if tt.args.opt.PollInterval != tt.want.PollInterval {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.PollInterval, tt.want.PollInterval)
			}
		})
	}
}

func Test_setServerFlag(t *testing.T) {
	type args struct {
		opt  *ServerOptions
		name string
	}
	tests := []struct {
		name string
		args args
		want ServerOptions
	}{
		{
			name: "set ServerAddr",
			args: args{
				opt:  &ServerOptions{},
				name: "ADDRESS",
			},
			want: ServerOptions{
				ServerAddr: "localhost:8080",
			},
		},
		{
			name: "set StoreInterval",
			args: args{
				opt:  &ServerOptions{},
				name: "STORE_INTERVAL",
			},
			want: ServerOptions{
				StoreInterval: 300,
			},
		},
		{
			name: "set FileStoragePath",
			args: args{
				opt:  &ServerOptions{},
				name: "FILE_STORAGE_PATH",
			},
			want: ServerOptions{
				FileStoragePath: "C:/files/metrics.txt",
			},
		},
		{
			name: "set RestoreOnStart",
			args: args{
				opt:  &ServerOptions{},
				name: "RESTORE",
			},
			want: ServerOptions{
				RestoreOnStart: true,
			},
		},
		{
			name: "set DatabaseDSN",
			args: args{
				opt:  &ServerOptions{},
				name: "DATABASE_DSN",
			},
			want: ServerOptions{
				DatabaseDSN: "postgres://postgres:123@localhost:5432/metrics",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setServerFlag(tt.args.opt)
			if tt.args.opt.ServerAddr != tt.want.ServerAddr {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.ServerAddr, tt.want.ServerAddr)
			}
			if tt.args.opt.StoreInterval != tt.want.StoreInterval {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.StoreInterval, tt.want.StoreInterval)
			}
			if tt.args.opt.FileStoragePath != tt.want.FileStoragePath {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.FileStoragePath, tt.want.FileStoragePath)
			}
			if tt.args.opt.RestoreOnStart != tt.want.RestoreOnStart {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.RestoreOnStart, tt.want.RestoreOnStart)
			}
			if tt.args.opt.DatabaseDSN != tt.want.DatabaseDSN {
				t.Errorf("setAgentFlag() %s got = %v, want %v", tt.name, tt.args.opt.DatabaseDSN, tt.want.DatabaseDSN)
			}
		})
	}
}
