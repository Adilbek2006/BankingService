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
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()
	fmt.Println("Успешно подключились к PostgreSQL!")

	redisCache, err := repository.NewRedisCache("localhost", "6379")
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	_ = redisCache
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	grpcServer := grpc.NewServer()

	repo := repository.NewAccountRepository(db)
	accountHandler := delivery.NewAccountHandler(repo)

	pb.RegisterAccountServiceServer(grpcServer, accountHandler)

	fmt.Println("Account Service запущен на порту 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
