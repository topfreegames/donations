package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/topfreegames/donations/metadata"
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "returns donations version",
	Long:  `returns donations version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("donations v%s\n", metadata.GetVersion())
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
