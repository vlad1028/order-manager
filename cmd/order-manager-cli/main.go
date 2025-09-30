package main

import (
	"github.com/vlad1028/order-manager/internal/cli"
	desc "github.com/vlad1028/order-manager/pkg/order-service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	grpcServerHost = "localhost:7001"
)

func main() {
	conn, err := grpc.NewClient(grpcServerHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create grpc client: %v", err)
	}
	defer conn.Close()

	orderServiceClient := desc.NewOrderServiceClient(conn)

	cliAdaptor := cli.NewOrderGrpcAdaptor(orderServiceClient)
	orderManagerCLI := cli.NewOrderManagerCLI(cliAdaptor, os.Stdin, os.Stdout)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s. Shutting down gracefully...", sig)
		orderManagerCLI.Shutdown()
		log.Printf("Shut down successfully")
		os.Exit(0)
	}()

	orderManagerCLI.RunInteractive()
}
