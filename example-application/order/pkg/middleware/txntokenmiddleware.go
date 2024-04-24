package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/trats"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const TXN_TOKEN_CONTEXT_KEY contextKey = "txn_token"

type purpose string

const (
	tradeTransactionDetails purpose = "TradeTransactionDetails"
	stocksTrade             purpose = "StocksTrade"
)

var str2Purpose = map[string]purpose{
	"TradeTransactionDetails": tradeTransactionDetails,
	"StocksTrade":             stocksTrade,
}

type tradeAction string

const (
	Buy  tradeAction = "Buy"
	Sell tradeAction = "Sell"
)

func getFieldFromAuthorizationContext[T any](authContext map[string]any, key string) (T, error) {
	value, ok := authContext[key]
	if !ok {
		return *new(T), fmt.Errorf("missing required '%s' key in authorization context", key)
	}

	typedValue, ok := value.(T)
	if !ok {
		return *new(T), fmt.Errorf("type assertion failed for '%s' key in authorization context; expected type '%T', found type '%T'", key, *new(T), value)
	}

	return typedValue, nil
}

func verifyTradeTransactionDetailsRequestContexts(purpose purpose, authContext map[string]any, r *http.Request) error {
	if purpose != tradeTransactionDetails {
		return fmt.Errorf("invalid purpose of the stock-details request; expected '%s', found '%s'", tradeTransactionDetails, purpose)
	}

	transactionID, err := getFieldFromAuthorizationContext[string](authContext, "transactionID")
	if err != nil {
		return err
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if transactionID != id {
		return fmt.Errorf("different transaction id parameter in the authorization context and request data of the transaction-details request; authorization context: '%s', request: '%s'", transactionID, id)
	}

	return nil
}

func verifyStockTradeRequestContexts(purpose purpose, authContext map[string]any, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	var orderRequest struct {
		OrderType string `json:"orderType"`
		StockID   int    `json:"stockID"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.Unmarshal(body, &orderRequest); err != nil {
		return err
	}

	if purpose != stocksTrade {
		return fmt.Errorf("invalid purpose of the stock-details request; expected '%s', found '%s'", stocksTrade, purpose)
	}

	action, err := getFieldFromAuthorizationContext[string](authContext, "action")
	if err != nil {
		return err
	}

	if action != string(Buy) && action != string(Sell) {
		return fmt.Errorf("invalid action in the authorization context of the stock-order request; expected '%s' or '%s', found '%s'", Buy, Sell, action)
	}

	if orderRequest.OrderType != string(Buy) && orderRequest.OrderType != string(Sell) {
		return fmt.Errorf("invalid action in the request data of the stock-order request; expected '%s' or '%s', found '%s'", Buy, Sell, orderRequest.OrderType)
	}

	if action != orderRequest.OrderType {
		return fmt.Errorf("different action parameter in the authorization context and request data of the stock-order; authorization context: '%s', request: '%s'", action, orderRequest.OrderType)
	}

	stockID, err := getFieldFromAuthorizationContext[float64](authContext, "stockID")
	if err != nil {
		return err
	}

	if int(stockID) != orderRequest.StockID {
		return fmt.Errorf("different stock id parameter in the authorization context and request data; authorization context: '%d', request: '%d'", int(stockID), orderRequest.StockID)
	}

	quantity, err := getFieldFromAuthorizationContext[float64](authContext, "quantity")
	if err != nil {
		return err
	}

	if int(quantity) != orderRequest.Quantity {
		return fmt.Errorf("different quantity parameter in the authorization context and request data; authorization context: '%d', request: '%d'", int(quantity), orderRequest.Quantity)
	}

	return nil
}

func verifyRequestContexts(token *trats.TxnToken, r *http.Request) error {
	if r.Header.Get("x-user-name") != token.Subject.Email {
		return fmt.Errorf("access denied: the user in the header does not match the expected user from the txn token; expected '%s', found '%s'", token.Subject.Email, r.Header.Get("x-user-name"))
	}

	pathTemplate, err := mux.CurrentRoute(r).GetPathTemplate()
	if err != nil {
		return err
	}

	authContext, ok := token.AuthorizationContext.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid authorization context format")
	}

	purpose, ok := str2Purpose[token.Purpose]
	if !ok {
		return fmt.Errorf("invalid request purpose")
	}

	switch {
	case pathTemplate == "/api/order" && r.Method == http.MethodPost:
		err = verifyStockTradeRequestContexts(purpose, authContext, r)
	case pathTemplate == "/api/order/{id}" && r.Method == http.MethodGet:
		err = verifyTradeTransactionDetailsRequestContexts(purpose, authContext, r)
	default:
		return fmt.Errorf("unexpected route: %s", pathTemplate)
	}

	return err
}

func getTxnTokenMiddleware(traTsVerifier *trats.Verifier, jwks string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken := r.Header.Get("txn-token")
			if rawToken == "" {
				logger.Error("Txn token not provided.")
				http.Error(w, "Unauthorized: Txn token not provided", http.StatusForbidden)

				return
			}

			token, err := traTsVerifier.ParseAndVerify(rawToken, jwks)
			if err != nil {
				logger.Error("Failed to verify txn token", zap.Error(err))
				http.Error(w, "Unauthorized: Invalid txn token", http.StatusForbidden)

				return
			}

			err = verifyRequestContexts(token, r)
			if err != nil {
				logger.Error("Failed to verify request contexts.", zap.Error(err))
				http.Error(w, "Failed to verify request contexts", http.StatusForbidden)

				return
			}

			ctx := context.WithValue(r.Context(), TXN_TOKEN_CONTEXT_KEY, rawToken)

			logger.Info("Txn token verified successfully.", zap.Any("txn-token-id", token.Id))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
