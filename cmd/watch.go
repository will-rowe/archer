package cmd

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a running Archer service",
	Long:  `Watch a running Archer service.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchWatcher()
	},
}

func init() {
	grpcAddr = watchCmd.Flags().String("grpcAddress", DefaultServerAddress, "address of the server hosting the Archer service")
	grpcPort = watchCmd.Flags().String("grpcPort", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	rootCmd.AddCommand(watchCmd)
}

// launchArcher sets up and runs a gRPC Archer client for watching the service
func launchWatcher() {

	// connect to the gRPC server
	addr := fmt.Sprintf("%s:%s", *grpcAddr, *grpcPort)
	log.Printf("\tdialing %v...", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to the Archer gRPC server: %v", err)
	}
	defer conn.Close()

	// establish the client
	client := api.NewArcherClient(conn)

	// create a watch stream request
	log.Println("\topening watch stream")
	req := &api.WatchRequest{ApiVersion: DefaultAPIVersion, SendFinished: true}
	stream, err := client.Watch(context.Background(), req)
	if err != nil {
		log.Fatalf("could not open watch stream: %v", err)
	}

	// set up control
	done := make(chan bool)
	go func() {
		for {
			resp, err := stream.Recv()

			// detect end of stream
			if err == io.EOF {
				log.Println("service finished, stopping watcher")
				done <- true
				return
			}

			// catch errors
			if err != nil {
				log.Fatalf("error receiving message: %v", err)
			}

			// log stream
			for _, sample := range resp.GetSamples() {
				log.Printf("\tsample: %v", sample.GetId())
			}
		}
	}()

	// block until stream is finished
	<-done
	log.Printf("finished")
}
