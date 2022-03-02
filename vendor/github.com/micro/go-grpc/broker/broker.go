// Package broker is the grpc broker. We import from go-plugins.
package broker

import (
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-plugins/broker/grpc"
)

var (
	DefaultBroker = grpc.NewBroker()
)

// NewBroker returns a new grpc broker
func NewBroker(opts ...broker.Option) broker.Broker {
	return grpc.NewBroker(opts...)
}
