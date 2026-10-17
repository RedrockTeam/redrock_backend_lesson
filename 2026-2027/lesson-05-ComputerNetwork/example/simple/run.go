package simple

import (
	"github.com/spf13/cobra"
)

func ChatSimple() *cobra.Command {
	simpleCmd := &cobra.Command{
		Use:   "simple",
		Short: "C2C Chat",
	}

	return simpleCmd
}

func ChatSimpleServer() *cobra.Command {
	var addr string

	simpleServerCmd := &cobra.Command{
		Use:   "server",
		Short: "Start Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			Server(addr)
			return nil
		},
	}

	simpleServerCmd.Flags().StringVarP(
		&addr,
		"addr",
		"a",
		":9090",
		"Listener Addr",
	)

	return simpleServerCmd
}

func ChatSimpleClient() *cobra.Command {
	var addr string

	simpleClientCmd := &cobra.Command{
		Use:   "client",
		Short: "Start Client",
		RunE: func(cmd *cobra.Command, args []string) error {
			Client(addr)
			return nil
		},
	}

	simpleClientCmd.Flags().StringVarP(
		&addr,
		"addr",
		"a",
		"127.0.0.1:9090",
		"Remote Addr",
	)

	return simpleClientCmd
}
