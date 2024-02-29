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

// package main
//
// import (
// 	"log"
//
// 	"github.com/minio/minio-go"
// )
//
// func main() {
// 	endpoint := "play.minio.io:9000"
// 	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
// 	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
// 	useSSL := true
//
// 	// Initialize minio client object.
// 	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
//
// 	log.Printf("%#v\n", minioClient) // minioClient is now setup
// }
