package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/paulja/go-fib-grpc/proto/fib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	var (
		num int
		seq int
	)
	flag.IntVar(&num, "n", -1, "call the Number API with the value specified")
	flag.IntVar(&seq, "s", -1, "call the Sequence API with the value specified")
	flag.Parse()

	if num >= 0 {
		c := makeClient()
		out, err := c.Number(context.Background(), &fib.NumberRequest{
			Number: int32(num),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(out)
	} else if seq >= 0 {
		c := makeClient()
		out, err := c.Sequence(context.Background(), &fib.SequenceRequest{
			Number: int32(seq),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(out)
	} else {
		log.Fatalln("no flags detected")
	}
}

func makeClient() fib.FibServiceClient {
	tlsCreds, err := makeTlsCredentials()
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}
	conn, err := grpc.NewClient(":4433", opts...)
	if err != nil {
		panic(err)
	}
	return fib.NewFibServiceClient(conn)
}

func makeTlsCredentials() (credentials.TransportCredentials, error) {
	ca, err := os.ReadFile("./etc/certs/ca.pem")
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to add CA certificate")
	}
	conf := &tls.Config{
		RootCAs: pool,
	}
	return credentials.NewTLS(conf), nil
}
