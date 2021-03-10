/*
Package cmd is the command line interface for the Archer microservice.

It contains subcommands for running the server and the client.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// default options
const (
	DefaultAPIVersion    = "1"
	DefaultgRPCport      = "9090"
	DefaultServerAddress = "localhost"
	DefaultLogFile       = "./archer.log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "archer",
	Short: "A sequence data microservice.",
	Long: `A sequence data microservice.

Currently supports:


Run help on a subcommand to find out more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
