# Archer
[![tests](https://github.com/will-rowe/archer/actions/workflows/tests.yml/badge.svg)](https://github.com/will-rowe/archer/actions/workflows/tests.yml)
[![godoc](https://godoc.org/github.com/will-rowe/archer?status.svg)](https://godoc.org/github.com/will-rowe/archer)
[![goreportcard](https://goreportcard.com/badge/github.com/will-rowe/archer)](https://goreportcard.com/report/github.com/will-rowe/archer)

***

## About

This is a basic microservice that pre-processes data before running CLIMB workflows. It has a gRPC API that supports start/cancel/watch of sample processing tasks, and includes a command line application for running a server and client (called `archer`).

### Dependencies

As well as the external Go packages listed in [go.mod](./go.mod), the following tools and packages are required to build the microservice executable and documentation:

* Make
* Go toolchain
* protoc
* protoc-gen-go
* protoc-gen-doc

### Installing

Easy installation is handled by the [Makefile](Makefile):

```
make all
```

This command will:
* compile the proto files for Go
* compile the gRPC API docs
* run fmt, lint and vet tools on the Go code
* run the unit tests
* build the Go executable

WIP: There is also a containerised version of the service available which is built via a [Github Action](.github/workflows/docker.yml). It can be obtained from Dockerhub:

```
docker pull willrowe/archer:latest
docker run -p 9090:9090 willrowe/archer:latest
```

### Testing

Unit tests are available for the service implementation. In addition several Go tools are used (Go lint, vet, fmt) to check the codebase. All these can be run separately using:

```
make test
make lint
make vet
```

A [Github Action](.github/workflows/tests.yml) is used to run continuous integration testing using the above make commands on linux and mac OS.

To test the gRPC code without having to connect to a real server we use the [mock package](https://github.com/golang/mock); the mock class was generated using:

```
mockgen github.com/will-rowe/archer/pkg/api/v1 ArcherClient > pkg/mock/client_mock.go
```

### Running

A client and server imlementation of the Archer microservice are available in a single binary called `archer`, which will be found in the `./bin` after installation.

To run the server:

```
archer launch
```

...

### Documentation

API documentation can be found [here](api/docs/v1/archer.md). Implementation documentation can be found [here](https://godoc.org/github.com/will-rowe/archer).

### WIP

This is all a work in progress. Things to do:

* add the application service layer
* add the storage layer
* daemonise the server
* implement the client in the app
* intergrate with [herald](www.github.com/will-rowe/herald)