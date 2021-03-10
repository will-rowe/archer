all: proto docs fmt lint vet test build

proto:
		protoc -I.  --go_out=plugins=grpc:pkg api/proto/v1/archer.proto

docs:
		protoc -I. --doc_out=api/docs/v1 --doc_opt=markdown,archer.md api/proto/v1/archer.proto	

fmt:
		go list ./... | grep -v /api/ | go fmt

lint:
		go list ./... | grep -v /api/ | xargs -L1 golint -set_exit_status

vet:
		go vet ./...

test:
		mockgen github.com/will-rowe/archer/pkg/api/v1 ArcherClient > pkg/mock/client_mock.go
		go test -v ./...

build: proto
		go mod tidy
		CGO_ENABLED=0 go build -o ./bin/archer .

pack: build
		docker build -t willrowe/archer:latest .

push:
		docker push willrowe/archer:latest
	
serve:
		docker run -p 9090:9090 willrowe/archer

clean:
		rm -r bin