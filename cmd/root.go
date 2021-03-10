/*
Package cmd is the command line interface for the Archer microservice.

It contains subcommands for running the server and the client.
*/
package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// default options
var (
	DefaultAPIVersion    = "1" // for now this is the only version available
	DefaultServerAddress = "localhost"
	DefaultgRPCport      = "9090"
	DefaultDbPath        = fmt.Sprintf("%v/.archer", getHome())
)

// command line arguments shared by two or more subcommands
var (
	grpcAddr *string // the address of the gRPC server
	grpcPort *string // TCP port to listen to by the gRPC server
	logFile  *string // the log file
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "archer",
	Short: "A sequence data microservice.",
	Long: `A sequence data microservice.

Currently supports:


Run help on a subcommand to find out more.`,
}

// init sets the persistent flags
func init() {
	logFile = rootCmd.PersistentFlags().StringP("logFile", "l", "", "where to write the server log (if unset, STDERR used)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getHome is used to find the user's home directory
func getHome() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return homeDir
}
