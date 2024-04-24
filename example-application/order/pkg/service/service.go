package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/middleware"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

type Service struct {
	DB             *sql.DB
	HTTPClient     *http.Client
	Config         *config.OrderConfig
	SpireJwtSource *workloadapi.JWTSource
	Logger         *zap.Logger
}

func NewService(db *sql.DB, httpClient *http.Client, config *config.OrderConfig, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) *Service {
	return &Service{
		DB:             db,
		HTTPClient:     httpClient,
		Config:         config,
		SpireJwtSource: spireJwtSource,
		Logger:         logger,
	}
}

type OrderType string

const (
	Buy  OrderType = "Buy"
	Sell OrderType = "Sell"
)

type OrderDetails struct {
	TransactionID string    `json:"transactionID"`
	Operation     OrderType `json:"operation"`
	StockName     string    `json:"stockName"`
	StockSymbol   string    `json:"stockSymbol"`
	StockID       int       `json:"stockID"`
	StockExchange string    `json:"stockExchange"`
	StockPrice    float64   `json:"stockPrice"`
	Quantity      int       `json:"quantity"`
	TotalValue    float64   `json:"totalValue"`
}

type UpdateRequest struct {
	OrderType OrderType `json:"orderType"`
	StockID   int       `json:"stockID"`
	Quantity  int       `json:"quantity"`
}

func (s *Service) Order(ctx context.Context, username string, stockID int, orderType OrderType, quantity int) (OrderDetails, error) {
	updateRequest := UpdateRequest{
		OrderType: orderType,
		StockID:   stockID,
		Quantity:  quantity,
	}

	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		s.Logger.Error("Error marshaling update request to the stocks server for an order request.", zap.Error(err))

		return OrderDetails{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/internal/stocks", s.Config.StocksServiceURL), bytes.NewBuffer(requestBody))
	if err != nil {
		s.Logger.Error("Error creating request to stocks server.", zap.Error(err))

		return OrderDetails{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-name", username)

	svidCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svid, err := s.SpireJwtSource.FetchJWTSVID(svidCtx, jwtsvid.Params{
		Audience: s.Config.SpiffeIDs.Stocks.String(),
	})
	if err != nil {
		s.Logger.Error("Failed to fetch JWT-SVID.", zap.Error(err))

		return OrderDetails{}, err
	}

	req.Header.Set("Authorization", "Bearer "+svid.Marshal())

	if txnToken, ok := ctx.Value(middleware.TXN_TOKEN_CONTEXT_KEY).(string); ok {
		req.Header.Set("txn-token", txnToken)
	}

	stockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req = req.WithContext(stockCtx)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.Logger.Error("Error calling stocks server for user stock update request.", zap.Error(err))

		return OrderDetails{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Failed to read response body from stocks server for user stock update request.", zap.Error(err))

		return OrderDetails{}, err
	}

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Received non-ok status from stocks server for user stock update request.",
			zap.Int("http-status-code", resp.StatusCode),
			zap.String("http-response", string(bodyBytes)))

		return OrderDetails{}, errors.New("unexpected response from stock server")
	}

	var orderDetails OrderDetails
	if err := json.Unmarshal(bodyBytes, &orderDetails); err != nil {
		s.Logger.Error("Error decoding order details from the response of the stocks server.", zap.Error(err))

		return OrderDetails{}, err
	}

	transactionID, err := gonanoid.New(10)
	if err != nil {
		s.Logger.Error("Transaction id generation failed: %v", zap.Error(err))

		return OrderDetails{}, err
	}

	orderDetails.TransactionID = transactionID
	orderDetails.TotalValue = float64(quantity) * orderDetails.StockPrice

	_, err = s.DB.Exec(`INSERT INTO order_table (order_id, username, stock_symbol, stock_name, stock_id, stock_exchange, stock_price, order_type, quantity, total_value) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		transactionID, username, orderDetails.StockSymbol, orderDetails.StockName, stockID, orderDetails.StockExchange, orderDetails.StockPrice, orderType, quantity, orderDetails.TotalValue)
	if err != nil {
		s.Logger.Error("Error registering an order transaction on the order database.", zap.Error(err))
	}

	return orderDetails, nil
}

func (s *Service) GetOrderDetails(username string, id string) (OrderDetails, error) {
	var orderDetails OrderDetails

	query := `SELECT order_id, stock_symbol, stock_name, stock_id, stock_exchange, stock_price, order_type, quantity, total_value FROM order_table WHERE order_id = ? and username = ?`
	row := s.DB.QueryRow(query, id, username)

	var operation string

	err := row.Scan(&orderDetails.TransactionID, &orderDetails.StockSymbol, &orderDetails.StockName, &orderDetails.StockID, &orderDetails.StockExchange, &orderDetails.StockPrice, &operation, &orderDetails.Quantity, &orderDetails.TotalValue)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Error("No order found with the given id for the user.", zap.String("order-id", id), zap.String("user-name", username), zap.Error(err))

			return OrderDetails{}, ErrOrderNotFound
		}

		s.Logger.Error("Error querying order details from the database.", zap.Error(err))

		return OrderDetails{}, err
	}

	orderDetails.Operation = OrderType(operation)

	return orderDetails, nil
}
