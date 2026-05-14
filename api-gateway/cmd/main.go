package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accountPb "BankingService/pb/account"
	userPb "BankingService/pb/user"
)

func main() {
	userConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()
	userClient := userPb.NewUserServiceClient(userConn)

	accountConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Account Service: %v", err)
	}
	defer accountConn.Close()
	accountClient := accountPb.NewAccountServiceClient(accountConn)

	router := gin.Default()

	router.POST("/users", func(c *gin.Context) {
		var reqBody struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := userClient.CreateUser(ctx, &userPb.CreateUserRequest{
			Name:  reqBody.Name,
			Email: reqBody.Email,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ERROR: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/accounts", func(c *gin.Context) {
		var reqBody struct {
			UserId   string `json:"user_id"`
			Currency string `json:"currency"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.CreateAccount(ctx, &accountPb.CreateAccountRequest{
			UserId:   reqBody.UserId,
			Currency: reqBody.Currency,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ERROR: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.GET("/accounts/:id", func(c *gin.Context) {
		accountID := c.Param("id")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{
			AccountId: accountID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ERROR: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	log.Println("API Gateway is running on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Gateway startup error: %v", err)
	}
}
