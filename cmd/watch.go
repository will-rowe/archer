package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	api "github.com/will-rowe/archer/pkg/api/v1"
	"github.com/will-rowe/archer/pkg/service/v1"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a running Archer service",
	Long: `Watch a running Archer service.
	
	This command will start a gRPC message stream and 
	print samples that have completed processing. It 
	will include sample name, amplicon coverage, S3
	location and processing time.`,
	Run: func(cmd *cobra.Command, args []string) {
		watcher()
	},
}

func init() {
	grpcAddr = watchCmd.Flags().String("grpcAddress", DefaultServerAddress, "address of the server hosting the Archer service")
	grpcPort = watchCmd.Flags().String("grpcPort", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	rootCmd.AddCommand(watchCmd)
}

// watcher sets up and runs a gRPC Archer client for watching the service
func watcher() {

	// connect to the gRPC server
	addr := fmt.Sprintf("%s:%s", *grpcAddr, *grpcPort)
	log.Printf("dialing %v", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to the Archer gRPC server: %v", err)
	}
	defer conn.Close()

	// establish the client
	client := api.NewArcherClient(conn)

	// prepare a graceful shutdown
	log.Println("preparing signal notifier")
	cSignalChan := make(chan os.Signal, 1)
	signal.Notify(cSignalChan, syscall.SIGINT, syscall.SIGTERM)
	sSignalChan := make(chan bool)
	errorChan := make(chan error)

	// create a watch stream request
	log.Println("opening watch stream")
	req := &api.WatchRequest{ApiVersion: DefaultAPIVersion, SendFinished: true}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.Watch(ctx, req)
	if err != nil {
		log.Fatalf("could not open watch stream: %v", err)
	}

	// wait for samples to be sent
	log.Print("completed samples:")
	go func() {
		for {
			resp, err := stream.Recv()

			// detect end of stream
			if err == io.EOF {
				sSignalChan <- true
				return
			}

			// catch errors
			if err != nil {
				errorChan <- fmt.Errorf("error receiving message: %v", err)
			}

			// log stream
			for _, sample := range resp.GetSamples() {
				covAmps, totAmps, meanCov := service.GetAmpliconCoverage(sample.GetProcessStats())
				log.Printf("\t- %v\t(%d/%d reads kept, %d/%d amplicons covered (mean coverage = %.0f))\t%v\tprocessed in %d seconds", sample.GetSampleID(), sample.GetProcessStats().GetKeptReads(), sample.GetProcessStats().GetTotalReads(), covAmps, totAmps, meanCov, sample.GetEndpoint(), (sample.GetEndTime().Seconds - sample.GetStartTime().GetSeconds()))
			}
		}
	}()

	// block until stream is finished or user input
	select {
	case <-sSignalChan:
		log.Print("server shut down signal received")
		break
	case <-cSignalChan:
		log.Print("client shut down signal received")
		break
	case <-ctx.Done():
		log.Print("context finished")
		break
	case <-errorChan:
		log.Fatal(err)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatal(err)
	}
	log.Printf("finished")
}
