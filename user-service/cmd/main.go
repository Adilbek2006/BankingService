package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	pb "BankingService/pb/user"
	"BankingService/user-service/internal/config"
	delivery "BankingService/user-service/internal/delivery/grpc"
	"BankingService/user-service/internal/email"
	"BankingService/user-service/internal/repository"
)

func main() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("user-service/.env")

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
	sender := buildSenderFromEnv()
	userHandler := delivery.NewUserHandler(repo, sender)

	pb.RegisterUserServiceServer(grpcServer, userHandler)

	fmt.Println("User Service is running on port 50053...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}

func buildSenderFromEnv() email.Sender {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	useTLS := strings.EqualFold(os.Getenv("SMTP_USE_TLS"), "true")
	insecure := strings.EqualFold(os.Getenv("SMTP_INSECURE_SKIP_VERIFY"), "true")

	if from == "" {
		from = username
	}

	if host == "" || port == "" || from == "" {
		log.Println("SMTP sender disabled: set SMTP_HOST, SMTP_PORT, SMTP_FROM")
		return email.NewNoopSender()
	}

	log.Println("SMTP sender enabled")
	return email.NewSMTPSender(host, port, username, password, from, useTLS, insecure)
}
