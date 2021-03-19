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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

var (
	customLogFunc = func(code codes.Code) logrus.Level {
		if code == codes.OK {
			return logrus.InfoLevel
		}
		return logrus.ErrorLevel
	}
)

// getServerOpts returns a []grpc.ServerOption with logging and metrics.
func getServerOpts(logrusEntry *logrus.Entry) []grpc.ServerOption {
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(customLogFunc),
	}
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
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
func Launch(ctx context.Context, serverAPI api.ArcherServer, cleanupAPI func() error, addr, logFile string) error {

	// setup the logger
	log := logrus.New()
	if len(logFile) != 0 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("could not start logger: %v", err)
		}
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
	grpcOpts := getServerOpts(logrus.NewEntry(log))
	server := grpc.NewServer(grpcOpts...)

	// register the Archer service
	logrus.Info("registering the Archer service on the gRPC server")
	api.RegisterArcherServer(server, serverAPI)

	// announce on the local network address
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	logrus.Infof("serving on %s", listen.Addr().String())

	// prepare a graceful shutdown
	logrus.Info("preparing signal notifier")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// start the server
	logrus.Info("starting gRPC server")
	errorChan := make(chan error)
	go func(errorChan chan<- error) {
		if err := server.Serve(listen); err != nil {
			errorChan <- err
		}
	}(errorChan)
	logrus.Info("ready")

	// wait for interrupt or context end
	select {
	case <-signalChan:
		logrus.Info("server shut down signal received")
		break
	case <-ctx.Done():
		break
	case err := <-errorChan:
		logrus.Error(err)
	}

	// stop the server and clean up the service
	logrus.Info("stopping gRPC server")
	server.GracefulStop()

	logrus.Info("cleaning up service")
	if err := cleanupAPI(); err != nil {
		return err
	}

	logrus.Info("finished")
	return nil
}
