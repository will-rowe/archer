package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Add a sample to the processing queue",
	Long: `Add a sample to the processing queue.
	
	The processing request is collecting via STDIN and should be
	in JSON. The request will be validated prior to submitting it
	to the Archer service.`,
	Run: func(cmd *cobra.Command, args []string) {
		process()
	},
}

func init() {
	grpcAddr = processCmd.Flags().String("grpcAddress", DefaultServerAddress, "address of the server hosting the Archer service")
	grpcPort = processCmd.Flags().String("grpcPort", DefaultgRPCport, "TCP port to listen to by the gRPC server")
	rootCmd.AddCommand(processCmd)
}

// process sets up and runs a gRPC Archer client for sending a request to process a sample
func process() {
	log.Print("starting Archer client")

	// check STDIN
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		log.Fatal("no data received from STDIN")
	}

	// collect request
	var processRequest *api.ProcessRequest
	if err := json.NewDecoder(os.Stdin).Decode(&processRequest); err != nil {
		log.Fatal(err)
	}

	// connect to the gRPC server
	addr := fmt.Sprintf("%s:%s", *grpcAddr, *grpcPort)
	log.Printf("dialing %v...", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to the Archer gRPC server: %v", err)
	}
	defer conn.Close()

	// establish the client
	client := api.NewArcherClient(conn)

	// send the request
	resp, err := client.Process(context.Background(), processRequest)
	if err != nil {
		errStatus, _ := status.FromError(err)
		log.Fatal(errStatus.Message())
	}

	// check response
	log.Print("response received from Archer service:")
	log.Print(resp)
}
