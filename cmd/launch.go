package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	server "github.com/will-rowe/archer/pkg/protocol/grpc"
	"github.com/will-rowe/archer/pkg/service/v1"
)

// command line arguments
var (
	grpcPort *string // TCP port to listen to by the gRPC server
	logFile  *string // the log file
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the Archer service",
	Long:  `Launch the Archer service.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchArcher()
	},
}

func init() {
	grpcPort = launchCmd.Flags().StringP("grpcPort", "g", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	logFile = launchCmd.Flags().StringP("logFile", "l", DefaultLogFile, "where to write the server log (use '-l -' for logging to standard out)")
	rootCmd.AddCommand(launchCmd)
}

// launchArcher sets up and runs the gRPC Archer service
func launchArcher() {

	// set up the log
	if *logFile != "-" {
		file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		log.SetOutput(file)
	}
	log.Println("starting Archer...")

	// get top level context
	ctx := context.Background()

	// get the service API
	serverAPI := service.NewArcher()

	// run the server until shutdown signal received
	if err := server.Launch(ctx, serverAPI, *grpcPort); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// clean up the service API
	if err := serverAPI.Shutdown(); err != nil {
		log.Fatalf("could not shutdown the Archer service: %v", err)
	}

	log.Println("finished")
}
