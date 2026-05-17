package main

import (
	"log"
	"os"
	"strings"

	"BankingService/user-service/internal/config"
	"BankingService/user-service/internal/email"
)

func main() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("user-service/.env")

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TEST_TO")
	useTLS := strings.EqualFold(os.Getenv("SMTP_USE_TLS"), "true")
	insecure := strings.EqualFold(os.Getenv("SMTP_INSECURE_SKIP_VERIFY"), "true")

	if from == "" {
		from = username
	}

	if host == "" || port == "" || from == "" || to == "" {
		log.Fatal("Set SMTP_HOST, SMTP_PORT, SMTP_FROM, SMTP_TEST_TO")
	}

	sender := email.NewSMTPSender(host, port, username, password, from, useTLS, insecure)
	if err := sender.Send(to, "SMTP test", "SMTP test email from BankingService"); err != nil {
		log.Fatalf("Send failed: %v", err)
	}

	log.Println("Test email sent")
}
