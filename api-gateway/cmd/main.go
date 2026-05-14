package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userPb "BankingService/pb/user"
)

func main() {
	userConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()

	userClient := userPb.NewUserServiceClient(userConn)

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User successfully created",
			"user_id": grpcResp.UserId,
			"name":    grpcResp.Name,
		})
	})

	log.Println("API Gateway is running on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Gateway startup error: %v", err)
	}
}
