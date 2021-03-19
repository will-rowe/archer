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

	"github.com/will-rowe/archer/pkg/version"
)

// default options
var (
	DefaultAPIVersion    = "1" // for now this is the only version available
	DefaultServerAddress = "localhost"
	DefaultgRPCport      = "9090"
	DefaultDbPath        = fmt.Sprintf("%v/.archer", getHome())
	DefaultManifestURL   = "https://raw.githubusercontent.com/artic-network/primer-schemes/master/schemes_manifest.json"
	DefaultBucketName    = "artic-archer-uploads-test"
)

// command line arguments shared by two or more subcommands
var (
	grpcAddr *string // the address of the gRPC server
	grpcPort *string // TCP port to listen to by the gRPC server
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "archer",
	Version: version.GetVersion(),
	Short:   "Artic Resource for Classifying, Honing & Exporting Reads",
	Long: `Artic Resource for Classifying, Honing & Exporting Reads.

	Archer is a command line application that implements the ARCHER API.
	It provides a service to pre-process data before running CLIMB workflows.
	It will check reads against a primer scheme, upload on target reads to
	an S3 bucket and then report back.`,
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
