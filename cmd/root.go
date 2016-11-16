package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Verbose determines how verbose donations will run under
var Verbose int

// RootCmd is the root command for donations CLI application
var RootCmd = &cobra.Command{
	Use:   "donations",
	Short: "Donation requests for clans.",
	Long:  "Donation requests for clans.",
}

// Execute runs RootCmd to initialize donations CLI application
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	RootCmd.PersistentFlags().IntVarP(
		&Verbose, "verbose", "v", 0,
		"Verbosity level => v0: Error, v1=Warning, v2=Info, v3=Debug",
	)
}
