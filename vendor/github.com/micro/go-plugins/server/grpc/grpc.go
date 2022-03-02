// Package grpc provides a grpc server
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/micro/go-log"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/codec"
	"github.com/micro/go-micro/errors"
	meta "github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"
	"github.com/micro/util/go/lib/addr"
	mgrpc "github.com/micro/util/go/lib/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// DefaultMaxMsgSize define maximum message size that server can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4
)

const (
	defaultContentType = "application/grpc"
)

type grpcServer struct {
	rpc  *rServer
	srv  *grpc.Server
	exit chan chan error
	wg   sync.WaitGroup

	sync.RWMutex
	opts        server.Options
	handlers    map[string]server.Handler
	subscribers map[*subscriber][]broker.Subscriber
	// used for first registration
	registered bool
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
	encoding.RegisterCodec(bytesCodec{})

	cmd.DefaultServers["grpc"] = NewServer
}

func newGRPCServer(opts ...server.Option) server.Server {
	options := newOptions(opts...)

	// create a grpc server
	srv := &grpcServer{
		opts: options,
		rpc: &rServer{
			serviceMap: make(map[string]*service),
		},
		handlers:    make(map[string]server.Handler),
		subscribers: make(map[*subscriber][]broker.Subscriber),
		exit:        make(chan chan error),
	}

	// configure the grpc server
	srv.configure()

	return srv
}

func (g *grpcServer) configure(opts ...server.Option) {
	// Don't reprocess where there's no config
	if len(opts) == 0 && g.srv != nil {
		return
	}

	for _, o := range opts {
		o(&g.opts)
	}

	maxMsgSize := g.getMaxMsgSize()

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnknownServiceHandler(g.handler),
	}

	if creds := g.getCredentials(); creds != nil {
		gopts = append(gopts, grpc.Creds(creds))
	}

	if opts := g.getGrpcOptions(); opts != nil {
		gopts = append(gopts, opts...)
	}

	g.srv = grpc.NewServer(gopts...)
}

func (g *grpcServer) getMaxMsgSize() int {
	if g.opts.Context == nil {
		return DefaultMaxMsgSize
	}
	s, ok := g.opts.Context.Value(maxMsgSizeKey{}).(int)
	if !ok {
		return DefaultMaxMsgSize
	}
	return s
}

func (g *grpcServer) getCredentials() credentials.TransportCredentials {
	if g.opts.Context != nil {
		if v := g.opts.Context.Value(tlsAuth{}); v != nil {
			tls := v.(*tls.Config)
			return credentials.NewTLS(tls)
		}
	}
	return nil
}

func (g *grpcServer) getGrpcOptions() []grpc.ServerOption {
	if g.opts.Context == nil {
		return nil
	}

	v := g.opts.Context.Value(grpcOptions{})

	if v == nil {
		return nil
	}

	opts, ok := v.([]grpc.ServerOption)

	if !ok {
		return nil
	}

	return opts
}

func (g *grpcServer) handler(srv interface{}, stream grpc.ServerStream) error {
	g.wg.Add(1)
	defer g.wg.Done()

	fullMethod, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return grpc.Errorf(codes.Internal, "method does not exist in context")
	}

	serviceName, methodName, err := mgrpc.ServiceMethod(fullMethod)
	if err != nil {
		return status.New(codes.InvalidArgument, err.Error()).Err()
	}

	g.rpc.mu.Lock()
	service := g.rpc.serviceMap[serviceName]
	g.rpc.mu.Unlock()

	if service == nil {
		return status.New(codes.Unimplemented, fmt.Sprintf("unknown service %v", service)).Err()
	}

	mtype := service.method[methodName]
	if mtype == nil {
		return status.New(codes.Unimplemented, fmt.Sprintf("unknown service %v", service)).Err()
	}

	// get grpc metadata
	gmd, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		gmd = metadata.MD{}
	}

	// copy the metadata to go-micro.metadata
	md := meta.Metadata{}
	for k, v := range gmd {
		md[k] = strings.Join(v, ", ")
	}

	// timeout for server deadline
	to := md["timeout"]

	// get content type
	ct := defaultContentType
	if ctype, ok := md["x-content-type"]; ok {
		ct = ctype
	}

	delete(md, "x-content-type")
	delete(md, "timeout")

	// create new context
	ctx := meta.NewContext(stream.Context(), md)

	// set the timeout if we have it
	if len(to) > 0 {
		if n, err := strconv.ParseUint(to, 10, 64); err == nil {
			ctx, _ = context.WithTimeout(ctx, time.Duration(n))
		}
	}

	// process unary
	if !mtype.stream {
		return g.processRequest(stream, service, mtype, ct, ctx)
	}

	// process stream
	return g.processStream(stream, service, mtype, ct, ctx)
}

func (g *grpcServer) processRequest(stream grpc.ServerStream, service *service, mtype *methodType, ct string, ctx context.Context) error {
	for {
		var argv, replyv reflect.Value

		// Decode the argument value.
		argIsValue := false // if true, need to indirect before calling.
		if mtype.ArgType.Kind() == reflect.Ptr {
			argv = reflect.New(mtype.ArgType.Elem())
		} else {
			argv = reflect.New(mtype.ArgType)
			argIsValue = true
		}

		// Unmarshal request
		if err := stream.RecvMsg(argv.Interface()); err != nil {
			return err
		}

		if argIsValue {
			argv = argv.Elem()
		}

		// reply value
		replyv = reflect.New(mtype.ReplyType.Elem())

		function := mtype.method.Func
		var returnValues []reflect.Value

		cc := defaultGRPCCodecs[ct]
		b, _ := cc.Marshal(argv.Interface())

		// create a client.Request
		r := &rpcRequest{
			service:     g.opts.Name,
			contentType: ct,
			method:      fmt.Sprintf("%s.%s", service.name, mtype.method.Name),
			body:        b,
			payload:     argv.Interface(),
		}

		// define the handler func
		fn := func(ctx context.Context, req server.Request, rsp interface{}) error {
			returnValues = function.Call([]reflect.Value{service.rcvr, mtype.prepareContext(ctx), reflect.ValueOf(argv.Interface()), reflect.ValueOf(rsp)})

			// The return value for the method is an error.
			if err := returnValues[0].Interface(); err != nil {
				return err.(error)
			}

			return nil
		}

		// wrap the handler func
		for i := len(g.opts.HdlrWrappers); i > 0; i-- {
			fn = g.opts.HdlrWrappers[i-1](fn)
		}

		statusCode := codes.OK
		statusDesc := ""

		// execute the handler
		if appErr := fn(ctx, r, replyv.Interface()); appErr != nil {
			if err, ok := appErr.(*rpcError); ok {
				statusCode = err.code
				statusDesc = err.desc
			} else if err, ok := appErr.(*errors.Error); ok {
				statusCode = microError(err)
				statusDesc = appErr.Error()
			} else {
				statusCode = convertCode(appErr)
				statusDesc = appErr.Error()
			}
			return status.New(statusCode, statusDesc).Err()
		}
		if err := stream.SendMsg(replyv.Interface()); err != nil {
			return err
		}
		return status.New(statusCode, statusDesc).Err()
	}
}

func (g *grpcServer) processStream(stream grpc.ServerStream, service *service, mtype *methodType, ct string, ctx context.Context) error {
	opts := g.opts

	r := &rpcRequest{
		service:     opts.Name,
		contentType: ct,
		method:      fmt.Sprintf("%s.%s", service.name, mtype.method.Name),
		stream:      true,
	}

	ss := &rpcStream{
		request: r,
		s:       stream,
	}

	function := mtype.method.Func
	var returnValues []reflect.Value

	// Invoke the method, providing a new value for the reply.
	fn := func(ctx context.Context, req server.Request, stream interface{}) error {
		returnValues = function.Call([]reflect.Value{service.rcvr, mtype.prepareContext(ctx), reflect.ValueOf(stream)})
		if err := returnValues[0].Interface(); err != nil {
			return err.(error)
		}

		return nil
	}

	for i := len(opts.HdlrWrappers); i > 0; i-- {
		fn = opts.HdlrWrappers[i-1](fn)
	}

	statusCode := codes.OK
	statusDesc := ""

	appErr := fn(ctx, r, ss)
	if appErr != nil {
		if err, ok := appErr.(*rpcError); ok {
			statusCode = err.code
			statusDesc = err.desc
		} else if err, ok := appErr.(*errors.Error); ok {
			statusCode = microError(err)
			statusDesc = appErr.Error()
		} else {
			statusCode = convertCode(appErr)
			statusDesc = appErr.Error()
		}
	}

	return status.New(statusCode, statusDesc).Err()
}

func (g *grpcServer) newCodec(contentType string) (codec.NewCodec, error) {
	if cf, ok := g.opts.Codecs[contentType]; ok {
		return cf, nil
	}
	if cf, ok := defaultRPCCodecs[contentType]; ok {
		return cf, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type: %s", contentType)
}

func (g *grpcServer) Options() server.Options {
	opts := g.opts
	return opts
}

func (g *grpcServer) Init(opts ...server.Option) error {
	g.configure(opts...)
	return nil
}

func (g *grpcServer) NewHandler(h interface{}, opts ...server.HandlerOption) server.Handler {
	return newRpcHandler(h, opts...)
}

func (g *grpcServer) Handle(h server.Handler) error {
	if err := g.rpc.register(h.Handler()); err != nil {
		return err
	}

	g.handlers[h.Name()] = h
	return nil
}

func (g *grpcServer) NewSubscriber(topic string, sb interface{}, opts ...server.SubscriberOption) server.Subscriber {
	return newSubscriber(topic, sb, opts...)
}

func (g *grpcServer) Subscribe(sb server.Subscriber) error {
	sub, ok := sb.(*subscriber)
	if !ok {
		return fmt.Errorf("invalid subscriber: expected *subscriber")
	}
	if len(sub.handlers) == 0 {
		return fmt.Errorf("invalid subscriber: no handler functions")
	}

	if err := validateSubscriber(sb); err != nil {
		return err
	}

	g.Lock()

	_, ok = g.subscribers[sub]
	if ok {
		return fmt.Errorf("subscriber %v already exists", sub)
	}
	g.subscribers[sub] = nil
	g.Unlock()
	return nil
}

func (g *grpcServer) Register() error {
	// parse address for host, port
	config := g.opts
	var advt, host string
	var port int

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// register service
	node := &registry.Node{
		Id:       config.Name + "-" + config.Id,
		Address:  addr,
		Port:     port,
		Metadata: config.Metadata,
	}

	node.Metadata["broker"] = config.Broker.String()
	node.Metadata["registry"] = config.Registry.String()
	node.Metadata["server"] = g.String()
	node.Metadata["transport"] = g.String()
	// node.Metadata["transport"] = config.Transport.String()

	g.RLock()
	// Maps are ordered randomly, sort the keys for consistency
	var handlerList []string
	for n, e := range g.handlers {
		// Only advertise non internal handlers
		if !e.Options().Internal {
			handlerList = append(handlerList, n)
		}
	}
	sort.Strings(handlerList)

	var subscriberList []*subscriber
	for e := range g.subscribers {
		// Only advertise non internal subscribers
		if !e.Options().Internal {
			subscriberList = append(subscriberList, e)
		}
	}
	sort.Slice(subscriberList, func(i, j int) bool {
		return subscriberList[i].topic > subscriberList[j].topic
	})

	var endpoints []*registry.Endpoint
	for _, n := range handlerList {
		endpoints = append(endpoints, g.handlers[n].Endpoints()...)
	}
	for _, e := range subscriberList {
		endpoints = append(endpoints, e.Endpoints()...)
	}
	g.RUnlock()

	service := &registry.Service{
		Name:      config.Name,
		Version:   config.Version,
		Nodes:     []*registry.Node{node},
		Endpoints: endpoints,
	}

	g.Lock()
	registered := g.registered
	g.Unlock()

	if !registered {
		log.Logf("Registering node: %s", node.Id)
	}

	// create registry options
	rOpts := []registry.RegisterOption{registry.RegisterTTL(config.RegisterTTL)}

	if err := config.Registry.Register(service, rOpts...); err != nil {
		return err
	}

	// already registered? don't need to register subscribers
	if registered {
		return nil
	}

	g.Lock()
	defer g.Unlock()

	g.registered = true

	for sb, _ := range g.subscribers {
		handler := g.createSubHandler(sb, g.opts)
		var opts []broker.SubscribeOption
		if queue := sb.Options().Queue; len(queue) > 0 {
			opts = append(opts, broker.Queue(queue))
		}
		sub, err := config.Broker.Subscribe(sb.Topic(), handler, opts...)
		if err != nil {
			return err
		}
		g.subscribers[sb] = []broker.Subscriber{sub}
	}

	return nil
}

func (g *grpcServer) Deregister() error {
	config := g.opts
	var advt, host string
	var port int

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	node := &registry.Node{
		Id:      config.Name + "-" + config.Id,
		Address: addr,
		Port:    port,
	}

	service := &registry.Service{
		Name:    config.Name,
		Version: config.Version,
		Nodes:   []*registry.Node{node},
	}

	log.Logf("Deregistering node: %s", node.Id)
	if err := config.Registry.Deregister(service); err != nil {
		return err
	}

	g.Lock()

	if !g.registered {
		g.Unlock()
		return nil
	}

	g.registered = false

	for sb, subs := range g.subscribers {
		for _, sub := range subs {
			log.Logf("Unsubscribing from topic: %s", sub.Topic())
			sub.Unsubscribe()
		}
		g.subscribers[sb] = nil
	}

	g.Unlock()
	return nil
}

func (g *grpcServer) Start() error {
	registerDebugHandler(g)
	config := g.opts

	// micro: config.Transport.Listen(config.Address)
	ts, err := net.Listen("tcp", config.Address)
	if err != nil {
		return err
	}

	log.Logf("Listening on %s", ts.Addr().String())
	g.Lock()
	g.opts.Address = ts.Addr().String()
	g.Unlock()

	// connect to the broker
	if err := config.Broker.Connect(); err != nil {
		return err
	}

	log.Logf("Broker [%s] Listening on %s", config.Broker.String(), config.Broker.Address())

	// announce self to the world
	if err := g.Register(); err != nil {
		log.Log("Server register error: ", err)
	}

	// micro: go ts.Accept(s.accept)
	go func() {
		if err := g.srv.Serve(ts); err != nil {
			log.Log("gRPC Server start error: ", err)
		}
	}()

	go func() {
		t := new(time.Ticker)

		// only process if it exists
		if g.opts.RegisterInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(g.opts.RegisterInterval)
		}

		// return error chan
		var ch chan error

	Loop:
		for {
			select {
			// register self on interval
			case <-t.C:
				if err := g.Register(); err != nil {
					log.Log("Server register error: ", err)
				}
			// wait for exit
			case ch = <-g.exit:
				break Loop
			}
		}

		// deregister self
		if err := g.Deregister(); err != nil {
			log.Log("Server deregister error: ", err)
		}

		// wait for waitgroup
		if wait(g.opts.Context) {
			g.wg.Wait()
		}

		// stop the grpc server
		g.srv.GracefulStop()

		// close transport
		ch <- nil

		// disconnect broker
		config.Broker.Disconnect()
	}()

	return nil
}

func (g *grpcServer) Stop() error {
	ch := make(chan error)
	g.exit <- ch
	return <-ch
}

func (g *grpcServer) String() string {
	return "grpc"
}

func NewServer(opts ...server.Option) server.Server {
	return newGRPCServer(opts...)
}
