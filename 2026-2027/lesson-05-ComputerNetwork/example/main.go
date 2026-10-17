package main

import (
	"example/simple"
	"os"

	"github.com/spf13/cobra"
)

func ChatCliRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "chat",
		Short: "Example Chat Server",
	}

	rootCmd.AddCommand(simple.ChatSimpleClient())
	rootCmd.AddCommand(simple.ChatSimpleServer())

	return rootCmd
}

func main() {
	if err := ChatCliRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
