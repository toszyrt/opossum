package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "opossum",
	Short: "Opossum is a collection of tools",
	Long:  "Opossum is a collection of tools",
	Run: Safely(func(cmd *cobra.Command, args []string) {
		fmt.Println("run opossum...")
	}),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Safely(fn func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var crash = true
		defer func() {
			if crash {
				log.Println("Crashed", "panic", recover())
				debug.PrintStack()
			}
		}()

		fn(cmd, args)
		crash = false
	}
}
