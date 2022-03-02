// Package client is the grpc client. We import from go-plugins.
package client

import (
	"github.com/micro/go-micro/client"
	"github.com/micro/go-plugins/client/grpc"
)

var (
	DefaultClient = grpc.NewClient()
)

// NewClient returns a new grpc client
func NewClient(opts ...client.Option) client.Client {
	return grpc.NewClient(opts...)
}
