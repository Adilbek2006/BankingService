package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"BankingService/internal/config"
	accountPb "BankingService/pb/account"
	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/broker"
	delivery "BankingService/transaction-service/internal/delivery/grpc"
	"BankingService/transaction-service/internal/repository"
	"BankingService/transaction-service/internal/usecase"
)

func main() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("transaction-service/.env")

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "user")
	dbPass := getEnv("DB_PASS", "password")
	dbName := getEnv("DB_NAME", "banking")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5673/")
	accountAddr := getEnv("ACCOUNT_GRPC_ADDR", "127.0.0.1:50051")
	grpcAddr := getEnv("GRPC_ADDR", ":50052")

	db, err := repository.NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("DB Connection failed: %v", err)
	}
	defer db.Close()

	rmq, err := broker.NewRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ Connection failed: %v", err)
	}
	defer rmq.Close()

	accountConn, err := grpc.Dial(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Account Service: %v", err)
	}
	defer accountConn.Close()
	accountClient := accountPb.NewAccountServiceClient(accountConn)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewTransactionRepository(db)
	uc := usecase.NewTransactionUsecase(repo, rmq, accountClient)
	h := delivery.NewTransactionHandler(uc)

	pb.RegisterTransactionServiceServer(grpcServer, h)

	log.Println("Transaction Service running on", grpcAddr)

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
