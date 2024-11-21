package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/paulja/go-fib-grpc/proto/fib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	log     slog.Logger
	timeout time.Duration = 2 * time.Second
	port                  = 4000
)

func main() {
	log = *slog.Default()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	log.Info("fib-service listening", "port", port)
	s := new(server)
	if err := s.Run(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

type server struct {
	fib.UnimplementedFibServiceServer
}

func logInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	hander grpc.UnaryHandler,
) (
	interface{},
	error,
) {
	log.DebugContext(ctx, "->", "method", info.FullMethod, "req", req)
	res, err := hander(ctx, req)
	log.Debug("<-", "method", info.FullMethod, "res", res, "error", err)
	return res, err
}

func (s *server) Run() error {
	listen, err := net.Listen("tcp", ":4000")
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logInterceptor),
	)
	reflection.Register(grpcServer) // do not run in production
	fib.RegisterFibServiceServer(grpcServer, s)
	return grpcServer.Serve(listen)
}

func (s *server) Number(
	ctx context.Context,
	req *fib.NumberRequest,
) (
	*fib.NumberResponse,
	error,
) {
	if req.Number <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "number must be greater than 0")
	}

	done := make(chan interface{})
	defer close(done)

	select {
	case num := <-fibonacci(done, req.Number):
		return &fib.NumberResponse{
			Result: num,
		}, nil
	case <-time.Tick(timeout):
		done <- true
		return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
	}
}

func (s *server) Sequence(
	ctx context.Context,
	req *fib.SequenceRequest,
) (
	*fib.SequenceResponse,
	error,
) {
	num := req.Number

	if num < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "number must be 0 or greater")
	}

	res := make([]int32, num)
	x, y := 0, 1
	for i := int32(0); i < num; i++ {
		x, y = y, x+y
		res = append(res, int32(x))
	}

	return &fib.SequenceResponse{
		Result: res,
	}, nil
}

func fibonacci(done <-chan interface{}, n int32) <-chan int32 {
	res := make(chan int32)
	go func() {
		defer close(res)

		select {
		case <-done:
			return
		default:
		}

		if n <= 2 {
			res <- 1
			return
		}
		res <- <-fibonacci(done, n-1) + <-fibonacci(done, n-2)
	}()
	return res
}
