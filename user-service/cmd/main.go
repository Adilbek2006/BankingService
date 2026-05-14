package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "BankingService/pb/user"
	delivery "BankingService/user-service/internal/delivery/grpc"
	"BankingService/user-service/internal/repository"
)

func main() {
	db, err := repository.NewPostgresDB("localhost", "5433", "user", "password", "banking")
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()
	fmt.Println("Successfully connected to PostgreSQL!")

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewUserRepository(db)
	userHandler := delivery.NewUserHandler(repo)

	pb.RegisterUserServiceServer(grpcServer, userHandler)

	fmt.Println("User Service is running on port 50053...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
