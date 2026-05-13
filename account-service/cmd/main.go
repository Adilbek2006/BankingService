package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	delivery "BankingService/account-service/internal/delivery/grpc"
	pb "BankingService/pb/account"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	grpcServer := grpc.NewServer()
	accountHandler := delivery.NewAccountHandler()
	pb.RegisterAccountServiceServer(grpcServer, accountHandler)

	fmt.Println("Account Service started on port: 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
