package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accountPb "BankingService/pb/account"
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

	accountConn, err := grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Account Service: %v", err)
	}
	defer accountConn.Close()
	accountClient := accountPb.NewAccountServiceClient(accountConn)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewTransactionRepository(db)
	transactionHandler := delivery.NewTransactionHandler(repo, rmq, accountClient)

	pb.RegisterTransactionServiceServer(grpcServer, transactionHandler)

	log.Println("Transaction Service running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
