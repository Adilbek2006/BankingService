package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	delivery "BankingService/account-service/internal/delivery/grpc"
	"BankingService/account-service/internal/repository"
	pb "BankingService/pb/account"
)

func main() {
	db, err := repository.NewPostgresDB("localhost", "5433", "user", "password", "banking")
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()
	fmt.Println("Successfully connected to PostgreSQL!")

	redisCache, err := repository.NewRedisCache("localhost", "6379")
	if err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewAccountRepository(db)

	accountHandler := delivery.NewAccountHandler(repo, redisCache)

	pb.RegisterAccountServiceServer(grpcServer, accountHandler)

	fmt.Println("Account Service is running on port 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
