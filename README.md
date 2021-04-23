# Archer
## Artic Resource for Classifying, Honing & Exporting Reads

***

[![tests](https://github.com/will-rowe/archer/actions/workflows/tests.yml/badge.svg)](https://github.com/will-rowe/archer/actions/workflows/tests.yml)
[![godoc](https://godoc.org/github.com/will-rowe/archer?status.svg)](https://godoc.org/github.com/will-rowe/archer)
[![goreportcard](https://goreportcard.com/badge/github.com/will-rowe/archer)](https://goreportcard.com/report/github.com/will-rowe/archer)

***

## About

This is a basic microservice that is used to pre-process data before running CLIMB workflows. It has a gRPC API that supports start/cancel/watch of sample processing tasks, and includes a command line application for running a server and client implementation (called `archer`).

If a user wants CLIMB to run the [artic pipeline](https://github.com/artic-network/fieldbioinformatics) on their data, the data needs to be checked locally and uploaded to an S3 bucket before CLIMB will process it. This is where `archer` comes in. It will:

* validate a sample as provided in a `ProcessRequest` for minimal metadata
* filter reads linked to that sample against the amplicon primer scheme
* compress all on-target reads and upload to S3
* report back

### Dependencies

As well as the external Go packages listed in [go.mod](./go.mod), the following tools and packages are required to build the microservice executable and documentation:

* Make
* Go toolchain
* protoc
* protoc-gen-go
* protoc-gen-doc

### Installing

Grab the release or use conda:

```
conda install -c bioconda archer
```

Easy installation from source is handled by the [Makefile](Makefile):

```
make all
```

This command will:
* compile the proto files for Go
* compile the gRPC API docs
* run fmt, lint and vet tools on the Go code
* run the unit tests
* build the Go executable

WIP: There is also a containerised version of the service available which is built via a [Github Action](.github/workflows/docker.yml). It can be obtained from Dockerhub but is not tested/supported currently:

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

To run the watch client:

```
archer watch
```

To run the process client:

```
cat sample.json | archer process
```

### Documentation

API documentation can be found [here](api/docs/v1/archer.md). Implementation documentation can be found [here](https://godoc.org/github.com/will-rowe/archer).

### Limitations/TODOs

* the S3 bucket upload is limited at the moment, it will need improving before this is in production
* might be worth adding an option to daemonise the server
* there is no index for amplicon bottom-k sketches, so each read loops over the ~90 amplicon sketches and picks the best one - not great
* read length filtering and jacard filtering are hard coded atm
* there is no containment search etc., just a basic similarity compairison between an amplicon sketch and a read sketch
* if something goes wrong during the sample processing, it is just marked as errored and there is no attempt to try again/fix it
    * it might be worth cycling through the db on service start and trying to re-process anything marked as errored - definitely need to report back to user
* intergrate with [herald](www.github.com/will-rowe/herald)
