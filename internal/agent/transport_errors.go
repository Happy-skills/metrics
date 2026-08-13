package agent

import (
	"errors"
	"net"
	"net/url"
)

type TransportErrorClassification int

const (
	NonRetriable TransportErrorClassification = iota
	Retriable
)

func classify(err error) TransportErrorClassification {
	if _, ok := errors.AsType[*url.Error](err); ok {
		return Retriable
	}

	if netErr, ok := errors.AsType[net.Error](err); ok {
		return classifyNetError(netErr)
	}

	return NonRetriable
}

func classifyNetError(netErr net.Error) TransportErrorClassification {
	if netErr.Timeout() {
		return Retriable
	}

	return NonRetriable
}
