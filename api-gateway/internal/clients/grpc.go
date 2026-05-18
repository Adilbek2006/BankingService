package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accountPb "BankingService/pb/account"
	transactionPb "BankingService/pb/transaction"
	userPb "BankingService/pb/user"
)

type GRPCClients struct {
	User        userPb.UserServiceClient
	Account     accountPb.AccountServiceClient
	Transaction transactionPb.TransactionServiceClient

	userConn    *grpc.ClientConn
	accountConn *grpc.ClientConn
	txConn      *grpc.ClientConn
}

func NewGRPCClients(userAddr, accountAddr, txAddr string) (*GRPCClients, error) {
	userConn, err := grpc.Dial(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect user service: %w", err)
	}

	accountConn, err := grpc.Dial(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		return nil, fmt.Errorf("connect account service: %w", err)
	}

	txConn, err := grpc.Dial(txAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		_ = accountConn.Close()
		return nil, fmt.Errorf("connect transaction service: %w", err)
	}

	return &GRPCClients{
		User:        userPb.NewUserServiceClient(userConn),
		Account:     accountPb.NewAccountServiceClient(accountConn),
		Transaction: transactionPb.NewTransactionServiceClient(txConn),
		userConn:    userConn,
		accountConn: accountConn,
		txConn:      txConn,
	}, nil
}

func (c *GRPCClients) Close() error {
	var firstErr error
	if c == nil {
		return nil
	}
	if c.userConn != nil {
		if err := c.userConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.accountConn != nil {
		if err := c.accountConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.txConn != nil {
		if err := c.txConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
