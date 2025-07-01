package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c2pc/go-pkg/interceptors"
	level2 "github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

var (
	ErrNoURLs             = errors.New("no URLs provided")
	ErrEmptyURL           = errors.New("empty url")
	ErrConnectionNotReady = errors.New("connection not ready")
)

func Connect(urls []string, serviceName string) (*grpc.ClientConn, error) {
	if len(urls) == 0 {
		return nil, ErrNoURLs
	}

	addr := make([]resolver.Address, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			return nil, ErrEmptyURL
		}
		addr = append(addr, resolver.Address{Addr: url})
	}
	var dialOptions []grpc.DialOption
	if logger.IsDebugEnabled(level2.TEST) {
		opts := []logging.Option{
			logging.WithLogOnEvents(logging.StartCall, logging.FinishCall, logging.PayloadSent, logging.PayloadReceived),
		}

		dialOptions = append(dialOptions,
			grpc.WithChainUnaryInterceptor(
				logging.UnaryClientInterceptor(interceptors.Logger(serviceName, false), opts...),
			),
			grpc.WithChainStreamInterceptor(
				logging.StreamClientInterceptor(interceptors.Logger(serviceName, false), opts...),
			),
		)
	}

	r := manual.NewBuilderWithScheme("grpc")
	r.InitialState(resolver.State{Addresses: addr})

	dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithResolvers(r))

	conn, err := grpc.NewClient(r.Scheme()+":///", dialOptions...)
	if err != nil {
		return nil, err
	}

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()

	if err := WaitForConnectionReady(ctx, conn); err != nil {
		return conn, ErrConnectionNotReady
	}

	return conn, err
}

func WaitForConnectionReady(ctx context.Context, conn *grpc.ClientConn) error {
	// A blocking dial blocks until the clientConn is ready.
	for {
		s := conn.GetState()
		if s == connectivity.Ready {
			return nil
		}
		if s == connectivity.Idle {
			conn.Connect()
		}
		if !conn.WaitForStateChange(ctx, s) {
			// ctx got timeout or canceled.
			return fmt.Errorf("waiting for connection to %s: %w", conn.Target(), ctx.Err())
		}
	}
}
