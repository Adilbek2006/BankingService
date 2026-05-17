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
	transactionPb "BankingService/pb/transaction"
	userPb "BankingService/pb/user"
)

func main() {
	userConn, err := grpc.Dial("127.0.0.1:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()
	userClient := userPb.NewUserServiceClient(userConn)

	accountConn, err := grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Account Service: %v", err)
	}
	defer accountConn.Close()
	accountClient := accountPb.NewAccountServiceClient(accountConn)

	txConn, err := grpc.Dial("127.0.0.1:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Transaction Service: %v", err)
	}
	defer txConn.Close()
	txClient := transactionPb.NewTransactionServiceClient(txConn)

	router := gin.Default()

	router.POST("/users", func(c *gin.Context) {
		var reqBody struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := userClient.CreateUser(ctx, &userPb.CreateUserRequest{
			Name:  reqBody.Name,
			Email: reqBody.Email,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.CreateAccount(ctx, &accountPb.CreateAccountRequest{
			UserId:   reqBody.UserId,
			Currency: reqBody.Currency,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.GET("/users/:id/accounts", func(c *gin.Context) {
		userID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.ListUserAccounts(ctx, &accountPb.UserIdRequest{
			UserId: userID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.PATCH("/accounts/:id/status", func(c *gin.Context) {
		accountID := c.Param("id")
		var reqBody struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.UpdateAccountStatus(ctx, &accountPb.UpdateStatusRequest{
			AccountId: accountID,
			Status:    reqBody.Status,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/accounts/:id/close", func(c *gin.Context) {
		accountID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.CloseAccount(ctx, &accountPb.AccountIdRequest{
			AccountId: accountID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/accounts/:id/freeze", func(c *gin.Context) {
		accountID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.FreezeAccount(ctx, &accountPb.AccountIdRequest{
			AccountId: accountID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/cards", func(c *gin.Context) {
		var reqBody struct {
			AccountId string `json:"account_id"`
			CardType  string `json:"card_type"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.IssueCard(ctx, &accountPb.IssueCardRequest{
			AccountId: reqBody.AccountId,
			CardType:  reqBody.CardType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.GET("/cards/:id", func(c *gin.Context) {
		cardID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.GetCardDetails(ctx, &accountPb.CardIdRequest{
			CardId: cardID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/cards/:id/block", func(c *gin.Context) {
		cardID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.BlockCard(ctx, &accountPb.CardIdRequest{
			CardId: cardID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.PATCH("/cards/:id/limit", func(c *gin.Context) {
		cardID := c.Param("id")
		var reqBody struct {
			Limit float64 `json:"limit"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.SetCardLimit(ctx, &accountPb.SetLimitRequest{
			CardId: cardID,
			Limit:  reqBody.Limit,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.PATCH("/accounts/:id/tier", func(c *gin.Context) {
		accountID := c.Param("id")
		var reqBody struct {
			Tier string `json:"tier"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.UpdateAccountTier(ctx, &accountPb.UpdateTierRequest{
			AccountId: accountID,
			Tier:      reqBody.Tier,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.GET("/accounts/:id/limits", func(c *gin.Context) {
		accountID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := accountClient.GetAccountLimits(ctx, &accountPb.AccountIdRequest{
			AccountId: accountID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/transactions/deposit", func(c *gin.Context) {
		var reqBody struct {
			AccountId string  `json:"account_id"`
			Amount    float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := txClient.ProcessDeposit(ctx, &transactionPb.DepositRequest{
			AccountId: reqBody.AccountId,
			Amount:    reqBody.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	router.POST("/transactions/transfer", func(c *gin.Context) {
		var reqBody struct {
			FromAccount string  `json:"from_account"`
			ToAccount   string  `json:"to_account"`
			Amount      float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcResp, err := txClient.InitiateTransfer(ctx, &transactionPb.TransferRequest{
			FromAccount: reqBody.FromAccount,
			ToAccount:   reqBody.ToAccount,
			Amount:      reqBody.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, grpcResp)
	})

	log.Println("API Gateway running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Gateway run error: %v", err)
	}
}
