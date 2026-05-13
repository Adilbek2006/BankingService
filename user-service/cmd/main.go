package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "BankingService/pb/user"
	delivery "BankingService/user-service/internal/delivery/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	grpcServer := grpc.NewServer()

	userHandler := delivery.NewUserHandler()

	pb.RegisterUserServiceServer(grpcServer, userHandler)

	fmt.Println(" User Service started on port: 50053...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
