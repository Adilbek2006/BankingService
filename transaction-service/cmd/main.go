package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "BankingService/pb/transaction"
	delivery "BankingService/transaction-service/internal/delivery/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	transactionHandler := delivery.NewTransactionHandler()

	pb.RegisterTransactionServiceServer(grpcServer, transactionHandler)

	log.Println("Transaction Service running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
