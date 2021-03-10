// Package grpc is the gRPC server implementation which runs the Archer service.
package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Launch runs a gRPC server to publish the Archer service.
func Launch(ctx context.Context, serverAPI api.ArcherServer, port string) error {

	// announce on the local network address
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	// register the Archer service
	// TODO: add logging to the gRPC server by passing options to NewServer
	server := grpc.NewServer()
	api.RegisterArcherServer(server, serverAPI)

	// prepare a graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// start the gRPC server
	go func() {
		if err := server.Serve(listen); err != grpc.ErrServerStopped {
			log.Fatalf("server error: %v", err)
		}
		log.Println("server shut down")
	}()

	// wait for signal before shutting down
	sig := <-signalChan
	log.Printf("caught signal: %v", sig)
	log.Println("shutting down server")
	server.GracefulStop()
	<-ctx.Done()
	return nil
}
