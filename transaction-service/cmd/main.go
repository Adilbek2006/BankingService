package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/broker"
	delivery "BankingService/transaction-service/internal/delivery/grpc"
	"BankingService/transaction-service/internal/repository"
)

func main() {
	db, err := repository.NewPostgresDB("127.0.0.1", "5433", "user", "password", "banking")
	if err != nil {
		log.Fatalf("DB Connection failed: %v", err)
	}
	defer db.Close()

	rmq, err := broker.NewRabbitMQ("amqp://guest:guest@127.0.0.1:5673/")
	if err != nil {
		log.Fatalf("RabbitMQ Connection failed: %v", err)
	}
	defer rmq.Close()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewTransactionRepository(db)
	transactionHandler := delivery.NewTransactionHandler(repo, rmq)

	pb.RegisterTransactionServiceServer(grpcServer, transactionHandler)

	log.Println("Transaction Service running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
