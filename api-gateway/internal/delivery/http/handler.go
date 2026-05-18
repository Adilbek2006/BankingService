package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"BankingService/api-gateway/internal/usecase"
)

type Handler struct {
	userUC     *usecase.UserUsecase
	accountUC  *usecase.AccountUsecase
	txUC       *usecase.TransactionUsecase
	reqTimeout time.Duration
}

func NewHandler(userUC *usecase.UserUsecase, accountUC *usecase.AccountUsecase, txUC *usecase.TransactionUsecase, timeout time.Duration) *Handler {
	return &Handler{userUC: userUC, accountUC: accountUC, txUC: txUC, reqTimeout: timeout}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/users", h.CreateUser)
	router.GET("/users/:id", h.GetUser)
	router.PATCH("/users/:id", h.UpdateUser)
	router.DELETE("/users/:id", h.DeleteUser)
	router.PATCH("/users/:id/password", h.ChangePassword)
	router.POST("/users/:id/verify-email", h.VerifyEmail)
	router.POST("/users/password-reset", h.SendPasswordReset)
	router.POST("/users/reset-password", h.ResetPassword)
	router.PATCH("/users/:id/kyc", h.UpdateKYC)
	router.GET("/users/:id/kyc", h.GetKYC)
	router.GET("/users", h.ListUsers)
	router.POST("/users/:id/suspend", h.SuspendUser)

	router.POST("/accounts", h.CreateAccount)
	router.GET("/accounts/:id", h.GetAccount)
	router.GET("/users/:id/accounts", h.ListUserAccounts)
	router.PATCH("/accounts/:id/status", h.UpdateAccountStatus)
	router.POST("/accounts/:id/close", h.CloseAccount)
	router.POST("/accounts/:id/freeze", h.FreezeAccount)
	router.POST("/cards", h.IssueCard)
	router.GET("/cards/:id", h.GetCard)
	router.POST("/cards/:id/block", h.BlockCard)
	router.PATCH("/cards/:id/limit", h.SetCardLimit)
	router.PATCH("/accounts/:id/tier", h.UpdateAccountTier)
	router.GET("/accounts/:id/limits", h.GetAccountLimits)

	router.POST("/transactions/deposit", h.ProcessDeposit)
	router.POST("/transactions/transfer", h.InitiateTransfer)
	router.POST("/transactions/withdrawal", h.ProcessWithdrawal)
	router.GET("/transactions/:id/status", h.GetTransactionStatus)
	router.GET("/accounts/:id/transactions", h.GetTransactionHistory)
	router.GET("/transactions/pending", h.ListPendingTransactions)
	router.POST("/transactions/:id/approve", h.ApproveTransaction)
	router.POST("/transactions/:id/reject", h.RejectTransaction)
	router.POST("/transactions/:id/reverse", h.ReverseTransaction)
	router.POST("/transactions/statement", h.GenerateStatement)
	router.POST("/transactions/fee", h.CalculateTransferFee)
	router.GET("/transactions/volume/daily", h.GetDailyVolume)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.CreateUser(ctx, reqBody.Name, reqBody.Email, reqBody.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var reqBody struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.UpdateUserProfile(ctx, userID, reqBody.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.DeleteUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	var reqBody struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.ChangePassword(ctx, userID, reqBody.OldPassword, reqBody.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	userID := c.Param("id")
	var reqBody struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.VerifyEmail(ctx, userID, reqBody.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SendPasswordReset(c *gin.Context) {
	var reqBody struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.SendPasswordReset(ctx, reqBody.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var reqBody struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.ResetPassword(ctx, reqBody.Token, reqBody.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateKYC(c *gin.Context) {
	userID := c.Param("id")
	var reqBody struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.UpdateKYCStatus(ctx, userID, reqBody.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetKYC(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.GetKYCStatus(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.ListUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SuspendUser(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.userUC.SuspendUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var reqBody struct {
		UserId   string `json:"user_id"`
		Currency string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.CreateAccount(ctx, reqBody.UserId, reqBody.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.GetAccountDetails(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListUserAccounts(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.ListUserAccounts(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateAccountStatus(c *gin.Context) {
	accountID := c.Param("id")
	var reqBody struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.UpdateAccountStatus(ctx, accountID, reqBody.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CloseAccount(c *gin.Context) {
	accountID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.CloseAccount(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) FreezeAccount(c *gin.Context) {
	accountID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.FreezeAccount(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) IssueCard(c *gin.Context) {
	var reqBody struct {
		AccountId string `json:"account_id"`
		CardType  string `json:"card_type"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.IssueCard(ctx, reqBody.AccountId, reqBody.CardType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCard(c *gin.Context) {
	cardID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.GetCardDetails(ctx, cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) BlockCard(c *gin.Context) {
	cardID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.BlockCard(ctx, cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SetCardLimit(c *gin.Context) {
	cardID := c.Param("id")
	var reqBody struct {
		Limit float64 `json:"limit"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.SetCardLimit(ctx, cardID, reqBody.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateAccountTier(c *gin.Context) {
	accountID := c.Param("id")
	var reqBody struct {
		Tier string `json:"tier"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.UpdateAccountTier(ctx, accountID, reqBody.Tier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetAccountLimits(c *gin.Context) {
	accountID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.accountUC.GetAccountLimits(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ProcessDeposit(c *gin.Context) {
	var reqBody struct {
		AccountId string  `json:"account_id"`
		Amount    float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.ProcessDeposit(ctx, reqBody.AccountId, reqBody.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) InitiateTransfer(c *gin.Context) {
	var reqBody struct {
		FromAccount string  `json:"from_account"`
		ToAccount   string  `json:"to_account"`
		Amount      float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.InitiateTransfer(ctx, reqBody.FromAccount, reqBody.ToAccount, reqBody.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ProcessWithdrawal(c *gin.Context) {
	var reqBody struct {
		AccountId string  `json:"account_id"`
		Amount    float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.ProcessWithdrawal(ctx, reqBody.AccountId, reqBody.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.GetTransactionStatus(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTransactionHistory(c *gin.Context) {
	accountID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.GetTransactionHistory(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListPendingTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.ListPendingTransactions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ApproveTransaction(c *gin.Context) {
	transactionID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.ApproveTransaction(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) RejectTransaction(c *gin.Context) {
	transactionID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.RejectTransaction(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ReverseTransaction(c *gin.Context) {
	transactionID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.ReverseTransaction(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GenerateStatement(c *gin.Context) {
	var reqBody struct {
		AccountId string `json:"account_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.GenerateStatement(ctx, reqBody.AccountId, reqBody.StartDate, reqBody.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CalculateTransferFee(c *gin.Context) {
	var reqBody struct {
		Amount       float64 `json:"amount"`
		TransferType string  `json:"transfer_type"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.CalculateTransferFee(ctx, reqBody.Amount, reqBody.TransferType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetDailyVolume(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.reqTimeout)
	defer cancel()

	resp, err := h.txUC.GetDailyTransactionVolume(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
