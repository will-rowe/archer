package cmd

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/will-rowe/archer/pkg/bucket"
	"github.com/will-rowe/archer/pkg/protocol/grpc"
	"github.com/will-rowe/archer/pkg/service/v1"
)

// command line options
var (
	grpcAddr      *string // the address of the gRPC server
	grpcPort      *string // TCP port to listen to by the gRPC server
	dbPath        *string // dbPath sets the location and filename for the Archer database
	manifestURL   *string // manifestURL tells archer where to collect the ARTIC primer scheme manifest
	numWorkers    *int    // number of concurrent request handlers to use
	numProcessors *int    // number of processors to use
	awsBucketName *string // the AWS S3 bucket name for uploading data to
	awsRegion     *string // the AWS region to use
	logFile       *string // the log file
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the Archer service",
	Long: `Launch the Archer service.
	
	This will start a gRPC server running that will
	accept incoming Process and Watch requests. It
	will offer the Archer API for filtering, compressing
	and uploading ARTIC reads to an S3 endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchArcher()
	},
}

func init() {
	grpcAddr = launchCmd.Flags().String("grpcAddress", DefaultServerAddress, "address to announce on")
	grpcPort = launchCmd.Flags().String("grpcPort", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	dbPath = launchCmd.Flags().String("dbPath", DefaultDbPath, "location to store the Archer database")
	manifestURL = launchCmd.Flags().String("manifestURL", DefaultManifestURL, "the ARTIC primer scheme manifest url")
	numWorkers = launchCmd.Flags().Int("numWorkers", 2, "number of concurrent request handlers to use")
	numProcessors = launchCmd.Flags().IntP("numProcessors", "p", -1, "number of processors to use (-1 == all)")
	awsBucketName = launchCmd.Flags().String("awsBucketName", DefaultBucketName, "the AWS S3 bucket name for data upload")
	awsRegion = launchCmd.Flags().String("awsRegion", bucket.DefaultRegion, "the AWS region to use")
	logFile = launchCmd.Flags().StringP("logFile", "l", "", "where to write the server log (if unset, STDERR used)")
	rootCmd.AddCommand(launchCmd)
}

// launchArcher sets up and runs the gRPC Archer service
func launchArcher() {
	runtime.GOMAXPROCS(*numProcessors)

	// get top level context
	ctx := context.Background()

	// get the service API
	serverAPI, cleanupAPI, err := service.NewArcher(service.SetNumWorkers(*numWorkers), service.SetDb(*dbPath), service.SetManifest(*manifestURL), service.SetBucket(*awsBucketName, *awsRegion))
	if err != nil {
		log.Fatalf("could not create Archer service: %v", err)
	}

	// run the server until shutdown signal received
	addr := fmt.Sprintf("%s:%s", *grpcAddr, *grpcPort)
	if err := grpc.Launch(ctx, serverAPI, cleanupAPI, addr, *logFile); err != nil {
		log.Fatal(err)
	}
}
