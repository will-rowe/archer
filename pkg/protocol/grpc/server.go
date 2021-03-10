// Package grpc is the gRPC server implementation which runs the Archer service.
package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// getServerOpts returns a []grpc.ServerOption with logging and metrics.
func getServerOpts() []grpc.ServerOption {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, logrusOpts...),
			//grpc_prometheus.UnaryServerInterceptor,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logrusOpts...),
			//grpc_prometheus.StreamServerInterceptor,
		),
	}
}

// Launch runs a gRPC server to publish the Archer service.
func Launch(ctx context.Context, serverAPI api.ArcherServer, addr, logFile string) error {

	// set up the logger
	var log = logrus.New()
	if len(logFile) != 0 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("could not start logger: %v", err)
		}
		defer file.Close()
		log.Out = file
	} else {
		log.Out = os.Stderr
	}
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	log.SetLevel(logrus.TraceLevel)

	// set up the gRPC server with logging
	grpcOpts := getServerOpts()
	server := grpc.NewServer(grpcOpts...)

	// register the Archer service
	log.Trace("registering the Archer service on the gRPC server")
	api.RegisterArcherServer(server, serverAPI)

	// announce on the local network address
	log.Tracef("announcing on %v", addr)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// prepare a graceful shutdown
	log.Trace("preparing signal notifier")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {

		// wait for incoming shutdown signal
		for range signalChan {
			log.Trace("shut down signal received")
			server.GracefulStop()
			<-ctx.Done()
			log.Trace("server stopped")
		}
	}()

	// start the gRPC server
	log.Trace("starting gRPC server")
	return server.Serve(listen)
}
