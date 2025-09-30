package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vlad1028/order-manager/internal/cache"
	"github.com/vlad1028/order-manager/internal/db"
	grpc2 "github.com/vlad1028/order-manager/internal/grpc"
	"github.com/vlad1028/order-manager/internal/kafka"
	"github.com/vlad1028/order-manager/internal/metrics"
	"github.com/vlad1028/order-manager/internal/order/service"
	desc "github.com/vlad1028/order-manager/pkg/order-service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	grpcHost    = "localhost:7001"
	httpHost    = "localhost:7000"
	metricsHost = "localhost:2112"
	kafkaHost   = "localhost:9092"
)

const (
	kafkaTopic = "pvz.events.log"
	cacheTTL   = 45 * time.Second
	day        = 24 * time.Hour
	week       = 7 * day
)

func main() {
	env := os.Getenv("ENV")
	ctx := context.Background()
	pool, err := db.ConnectToDB(env)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	orderRepo := db.SetupOrderRepository(pool)
	kafkaProducer, err := kafka.NewSyncProducer([]string{kafkaHost}, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	redis := cache.MustNew(ctx, cacheTTL)

	orderService := service.NewOrderService(0, week, 2*day, orderRepo, kafkaProducer, redis)
	grpcAdaptor := grpc2.NewOrderGrpcAdaptor(orderService)

	lis, err := net.Listen("tcp", grpcHost)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	desc.RegisterOrderServiceServer(grpcServer, grpcAdaptor)

	mux := runtime.NewServeMux()
	err = desc.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, grpcHost, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	if err != nil {
		log.Fatalf("failed to register order service handler: %v", err)
	}

	go func() {
		if err = http.ListenAndServe(httpHost, mux); err != nil {
			log.Fatalf("failed to listen and serve order service handler: %v", err)
		}
	}()
	metrics.StartMetricsServer(metricsHost)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
