package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/will-rowe/archer/pkg/protocol/grpc"
	"github.com/will-rowe/archer/pkg/service/v1"
)

// command line options
var (
	dbPath *string // dbPath sets the location and filename for the Archer database
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
	grpcAddr = launchCmd.Flags().String("grpcAddress", DefaultServerAddress, "address to announce on")
	grpcPort = launchCmd.Flags().String("grpcPort", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	dbPath = launchCmd.Flags().String("dbPath", DefaultDbPath, "location to store the Archer database")
	rootCmd.AddCommand(launchCmd)
}

// launchArcher sets up and runs the gRPC Archer service
func launchArcher() {

	// get top level context
	ctx := context.Background()

	// get the service API
	serverAPI, cleanupAPI, err := service.NewArcher(*dbPath)
	if err != nil {
		log.Fatalf("could not create Archer service: %v", err)
	}

	// run the server until shutdown signal received
	addr := fmt.Sprintf("%s:%s", *grpcAddr, *grpcPort)
	if err := grpc.Launch(ctx, serverAPI, cleanupAPI, addr, *logFile); err != nil {
		log.Fatal(err)
	}
}
