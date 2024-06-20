package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Opossum",
		Long:  `All software has versions. This is Opossum's`,
		Run:   Safely(version),
	}
	rootCmd.AddCommand(cmd)
}

func version(cmd *cobra.Command, args []string) {
	fmt.Println("Opossum v0.1")
}
