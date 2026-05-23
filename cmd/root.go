package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "lan-notify",
	Version: "1.0.0",
	Short:   "LAN Notify is a local network notification system",
	Long:    `A lightweight background service and CLI client to send and receive desktop notifications over the local network using mDNS and HTTPS.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
