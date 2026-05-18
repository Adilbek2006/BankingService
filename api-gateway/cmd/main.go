package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"BankingService/api-gateway/internal/clients"
	httpDelivery "BankingService/api-gateway/internal/delivery/http"
	"BankingService/api-gateway/internal/usecase"
	"BankingService/internal/config"
)

func main() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("api-gateway/.env")

	userAddr := getEnv("USER_GRPC_ADDR", "127.0.0.1:50053")
	accountAddr := getEnv("ACCOUNT_GRPC_ADDR", "127.0.0.1:50051")
	txAddr := getEnv("TX_GRPC_ADDR", "127.0.0.1:50052")
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	requestTimeout := getEnvDuration("REQUEST_TIMEOUT", 5*time.Second)

	grpcClients, err := clients.NewGRPCClients(userAddr, accountAddr, txAddr)
	if err != nil {
		log.Fatalf("Failed to init gRPC clients: %v", err)
	}
	defer func() {
		if err := grpcClients.Close(); err != nil {
			log.Printf("Failed to close gRPC clients: %v", err)
		}
	}()

	userUC := usecase.NewUserUsecase(grpcClients.User)
	accountUC := usecase.NewAccountUsecase(grpcClients.Account)
	txUC := usecase.NewTransactionUsecase(grpcClients.Transaction)

	h := httpDelivery.NewHandler(userUC, accountUC, txUC, requestTimeout)

	router := gin.Default()
	h.RegisterRoutes(router)

	log.Println("API Gateway running on", httpAddr)
	if err := router.Run(httpAddr); err != nil {
		log.Fatalf("Gateway run error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}
