package grpc

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/micro/go-micro/codec"
	"github.com/micro/go-micro/codec/bytes"
	"github.com/micro/go-micro/codec/jsonrpc"
	"github.com/micro/go-micro/codec/protorpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

type jsonCodec struct{}
type protoCodec struct{}
type bytesCodec struct{}
type wrapCodec struct{ encoding.Codec }

var (
	defaultGRPCCodecs = map[string]encoding.Codec{
		"application/json":         jsonCodec{},
		"application/proto":        protoCodec{},
		"application/protobuf":     protoCodec{},
		"application/octet-stream": protoCodec{},
		"application/grpc+json":    jsonCodec{},
		"application/grpc+proto":   protoCodec{},
		"application/grpc+bytes":   bytesCodec{},
	}

	defaultRPCCodecs = map[string]codec.NewCodec{
		"application/json":         jsonrpc.NewCodec,
		"application/json-rpc":     jsonrpc.NewCodec,
		"application/protobuf":     protorpc.NewCodec,
		"application/proto-rpc":    protorpc.NewCodec,
		"application/octet-stream": protorpc.NewCodec,
	}

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// UseNumber fix unmarshal Number(8234567890123456789) to interface(8.234567890123457e+18)
func UseNumber() {
	json = jsoniter.Config{
		UseNumber:              true,
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
}

func (w wrapCodec) String() string {
	return w.Codec.Name()
}

func (w wrapCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*bytes.Frame)
	if ok {
		return b.Data, nil
	}
	return w.Codec.Marshal(v)
}

func (w wrapCodec) Unmarshal(data []byte, v interface{}) error {
	b, ok := v.(*bytes.Frame)
	if ok {
		b.Data = data
		return nil
	}
	return w.Codec.Unmarshal(data, v)
}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*bytes.Frame)
	if ok {
		return b.Data, nil
	}
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) Name() string {
	return "proto"
}

func (bytesCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*[]byte)
	if !ok {
		return nil, fmt.Errorf("failed to marshal: %v is not type of *[]byte", v)
	}
	return *b, nil
}

func (bytesCodec) Unmarshal(data []byte, v interface{}) error {
	b, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal: %v is not type of *[]byte", v)
	}
	*b = data
	return nil
}

func (bytesCodec) Name() string {
	return "bytes"
}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

type grpcCodec struct {
	// headers
	id       string
	target   string
	method   string
	endpoint string

	s grpc.ClientStream
	c encoding.Codec
}

func (g *grpcCodec) ReadHeader(m *codec.Message, mt codec.MessageType) error {
	md, err := g.s.Header()
	if err != nil {
		return err
	}
	if m == nil {
		m = new(codec.Message)
	}
	if m.Header == nil {
		m.Header = make(map[string]string)
	}
	for k, v := range md {
		m.Header[k] = strings.Join(v, ",")
	}
	m.Id = g.id
	m.Target = g.target
	m.Method = g.method
	m.Endpoint = g.endpoint
	return nil
}

func (g *grpcCodec) ReadBody(v interface{}) error {
	if f, ok := v.(*bytes.Frame); ok {
		return g.s.RecvMsg(f)
	}
	return g.s.RecvMsg(v)
}

func (g *grpcCodec) Write(m *codec.Message, v interface{}) error {
	// if we don't have a body
	if v != nil {
		return g.s.SendMsg(v)
	}
	// write the body using the framing codec
	return g.s.SendMsg(&bytes.Frame{m.Body})
}

func (g *grpcCodec) Close() error {
	return g.s.CloseSend()
}

func (g *grpcCodec) String() string {
	return g.c.Name()
}
