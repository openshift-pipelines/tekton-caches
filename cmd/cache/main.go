package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cache",
		Short:        "cache wrapper",
		SilenceUsage: true,
	}

	// FIXME add k8s client

	cmd.AddCommand(fetchCmd())
	cmd.AddCommand(uploadCmd())

	return cmd
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	cache := rootCmd()
	if err := cache.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
