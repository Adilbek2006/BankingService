package main

import (
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"BankingService/account-service/internal/broker"
	delivery "BankingService/account-service/internal/delivery/grpc"
	"BankingService/account-service/internal/repository"
	"BankingService/account-service/internal/usecase"
	"BankingService/internal/config"
	pb "BankingService/pb/account"
)

func main() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("account-service/.env")

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "user")
	dbPass := getEnv("DB_PASS", "password")
	dbName := getEnv("DB_NAME", "banking")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5673/")
	grpcAddr := getEnv("GRPC_ADDR", ":50051")

	db, err := repository.NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	redisHost, redisPort := splitHostPort(redisAddr, "6379")
	redisCache, err := repository.NewRedisCache(redisHost, redisPort)
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}

	repo := repository.NewAccountRepository(db)
	uc := usecase.NewAccountUsecase(repo, redisCache)

	rmq, err := broker.NewRabbitMQ(rabbitURL, repo, redisCache)
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer rmq.Close()

	rmq.ConsumeDeposits()
	rmq.ConsumeTransfers()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Listener error: %v", err)
	}

	grpcServer := grpc.NewServer()
	accountHandler := delivery.NewAccountHandler(uc)
	pb.RegisterAccountServiceServer(grpcServer, accountHandler)

	log.Println("Account Service running on", grpcAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func splitHostPort(addr, defaultPort string) (string, string) {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return addr, defaultPort
}
