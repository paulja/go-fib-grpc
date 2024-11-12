package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/paulja/go-fib-grpc/proto/fib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var log slog.Logger

func main() {
	log = *slog.Default()
	slog.SetLogLoggerLevel(slog.LevelDebug)
	shutdownHandler()

	log.Info("fib-service listening", "port", 4000)
	s := new(server)
	if err := s.Run(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func shutdownHandler() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		s := <-exit
		if s == syscall.SIGINT { // if killed from CTRL+C
			fmt.Print("\b\b")
		}
		log.Info("graceful shutdown")
		os.Exit(0)
	}()
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
	return &fib.NumberResponse{
		Result: fibonacci(req.Number),
	}, nil
}

func (s *server) Sequence(
	ctx context.Context,
	req *fib.SequenceRequest,
) (
	*fib.SequenceResponse,
	error,
) {
	if req.Number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "number must be 0 or greater")
	}

	results := make([]int32, 0)
	next := sequence()
	for i := 0; i < int(req.Number); i++ {
		results = append(results, next())
	}

	return &fib.SequenceResponse{
		Result: results,
	}, nil
}

func fibonacci(n int32) int32 {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func sequence() func() int32 {
	x, y := 0, 1
	return func() int32 {
		x, y = y, x+y
		return int32(x)
	}
}
