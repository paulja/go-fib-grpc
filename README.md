# Fibonacci Service Delivered Over gRPC

The purpose of the repo is to show how to create a gRPC service with TLS by using:

- Docker
- Docker Compose
- Go
- gRPC + ProtoBufs

## Building the Proto files
From the `./proto/` folder, you need to get the Go gRPC dependency.

```shell
## go get google.golang.org/grpc has been added to the go.mod file
go mod tidy
```

Then go back to the root folder for the project. To build the proto buffers source code for gRPC install `buf`.

```shell
brew install buf
```

And then run the `generate` command.

```shell
buf generate ./proto/fib/service.proto
```

Two new files will now exist in the `proto/fib` folder, one for the serialisation of the proto buffer objects and one for the gRPC comms.

> [!note]
> `buf` requires additional tooling: `protoc` for Go. More information can be found here: https://grpc.io/docs/languages/go/quickstart/

## Testing Outside of the Container Environment
You can test the service outside the container environment by launching the service. Make sure you are in the `svc` folder.

```shell
go run .cmd/server/main.go
```

Then you can use a gRPC client like `grpcurl`.

```shell
brew install grpcurl
```

Now we can send messages to the service using JSON like payloads.

```shell
grpcurl \
    -proto proto/fib/service.proto \
    -plaintext \
    -d '{"number":15}' \
    :4000 FibService/Number
```

We have not enabled reflection in our service therefore we have to tell `grpcurl` what services and operations are available. The `-plaintext` flag is required because we have not configured TLS, the `-d` specifies the data in a JSON like format, the remainder specifies the host, port and API we wish to call.

Using the other API we created is as simple.

```shell
grpcurl \
    -proto proto/fib/service.proto \
    -plaintext \
    -d '{"number":25}' \
    :4000 FibService/Sequence
```


## Enabling TLS
The TLS certificates are created by using CloudFlare PKI toolkit, `cfssl`.

```shell
brew install cfssl
```

Create the certs needed by using the `Makefile` in the TLS folder.

```shell
make gencert
```

That will use `cfssl` to create the certs for the service and move them in the right place.

We choosing to run the service with TLS via a proxy, as that is the typical approach when you run the services in production, otherwise you have to write code to create a certificate pool and add then TLS configuration to your dial options. This not hard code to write, however, it means you will have to release code when you need to change certs unless you code things very carefully. But far the easiest approach is to create a reverse proxy for you service and upload and manager your certificates there.

We will be using NGINX for the reverse proxy in this case.

## Running in Docker
We want to run the service in Docker with Docker Compose, now the certificates have been generated we are ready to run in docker.

```shell
docker compose up --build
```

We only need the `--build` tag the first time we run as that will create the docker container for the service and then run it.

After the service and reverse proxy come up, you can run then same `grpcurl` call but you can remove the `-plaintext` flag. However, because we are running a self signed certificate we have to ask `grpcurl` to not check the certs with a `-insecure` flag.

```shell
grpcurl \
    -proto proto/fib/service.proto \
    -insecure \
    -d '{"number":10}' \
    :4433 FibService/Number
```

## Quality of Life Changes
Specifying the proto file with each request is fine for production as you do not want to widen your attack surface, but during development it is convenient to enable gRPC reflection so  you can discover the API and arguments.

```shell
grpcurl -insecure :4433 list FibService
```

Returns for our service.

```shell
FibService.Number
FibService.Sequence
```

Or you can get more information than the `list` command provides by using the `describe` command.

```shell
grpcurl -insecure :4433 describe FibService
```

Which returns more information.

```shell
grpcurl -insecure :4433 describe FibService

FibService is a service:
service FibService {
  rpc Number ( .NumberRequest ) returns ( .NumberResponse );
  rpc Sequence ( .SequenceRequest ) returns ( .SequenceResponse );
}
```

## Using the Custom CLI Client

To show how to consume our service, there is client code in the project, which you can call from the `svc` folder.

```shell
go run ./cmd/client/main.go -help
Usage of /var/folders/...main:
  -n int
    	call the Number API with the value specified (default -1)
  -s int
    	call the Sequence API with the value specified (default -1)
```

Calling the client with `-n` or `-s` results in calling the service.

```shell
go run ./cmd/client/main.go -n 10

2024/11/12 09:58:45 result:55
```

## Conclusion
And there we have a Go gRPC Service using Docker Compose and TLS, we also have created a client CLI app to act a consumer of the service.

