// Package server is the grpc server. We import from go-plugins.
package server

import (
	"github.com/micro/go-micro/server"
	"github.com/micro/go-plugins/server/grpc"
)

var (
	DefaultServer = grpc.NewServer()
)

// NewServer returns a new grpc server
func NewServer(opts ...server.Option) server.Server {
	return grpc.NewServer(opts...)
}
