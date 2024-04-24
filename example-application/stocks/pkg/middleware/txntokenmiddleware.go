package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/trats"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type purpose string

const (
	stockSearch           purpose = "StockSearch"
	stockDetails          purpose = "StockDetails"
	stocksHoldingsDetails purpose = "StocksHoldingsDetails"
	stocksTrade           purpose = "StocksTrade"
)

var str2Purpose = map[string]purpose{
	"StockSearch":           stockSearch,
	"StockDetails":          stockDetails,
	"StocksHoldingsDetails": stocksHoldingsDetails,
	"StocksTrade": stocksTrade,
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

func verifyStockSearchAuthorizationContexts(purpose purpose, authContext map[string]any, r *http.Request) error {
	if purpose != stockSearch {
		return fmt.Errorf("invalid purpose of the stock-search request; expected '%s', found '%s'", stockSearch, purpose)
	}

	contextQuery, err := getFieldFromAuthorizationContext[string](authContext, "query")
	if err != nil {
		return err
	}

	query := r.URL.Query().Get("query")

	if contextQuery != query {
		return fmt.Errorf("different search query parameter in the authorization context and request data; authorization context: '%s', request: '%s'", contextQuery, query)
	}

	return nil
}

func verifyStocksHoldingsAuthorizationContexts(purpose purpose) error {
	if purpose != stocksHoldingsDetails {
		return fmt.Errorf("invalid purpose of the stock-holdings request; expected '%s', found '%s'", stocksHoldingsDetails, purpose)
	}

	return nil
}

func verifyStocksDetailsAuthorizationContexts(purpose purpose, authContext map[string]any, r *http.Request) error {
	if purpose != stockDetails {
		return fmt.Errorf("invalid purpose of the stock-details request; expected '%s', found '%s'", stockDetails, purpose)
	}

	stockID, err := getFieldFromAuthorizationContext[float64](authContext, "stockID")
	if err != nil {
		return err
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("could not verify txn token; invalid stock id in the request: %v", err)
	}

	if int(stockID) != id {
		return fmt.Errorf("different stock id parameter in the authorization context and request data; authorization context: '%d', request: '%d'", int(stockID), id)
	}

	return nil
}

func verifyUpdateUserStockAuthorizationContexts(purpose purpose, authContext map[string]any, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var updateRequest struct {
		OrderType string `json:"orderType"`
		StockID   int    `json:"stockID"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.Unmarshal(body, &updateRequest); err != nil {
		return err
	}

	if purpose != stocksTrade {
		return fmt.Errorf("invalid purpose of the user-stock-update request; expected '%s', found '%s'", stocksTrade, purpose)
	}

	action, err := getFieldFromAuthorizationContext[string](authContext, "action")
	if err != nil {
		return err
	}

	if action != string(Buy) && action != string(Sell) {
		return fmt.Errorf("invalid action in the authorization context of the user-stock-update request; expected '%s' or '%s', found '%s'", Buy, Sell, action)
	}

	if updateRequest.OrderType != string(Buy) && updateRequest.OrderType != string(Sell) {
		return fmt.Errorf("invalid action in the request data of the user-stock-update request; expected '%s' or '%s', found '%s'", Buy, Sell, updateRequest.OrderType)
	}

	if action != updateRequest.OrderType {
		return fmt.Errorf("different action parameter in the authorization context and request data; authorization context: '%s', request: '%s'", action, updateRequest.OrderType)
	}

	stockID, err := getFieldFromAuthorizationContext[float64](authContext, "stockID")
	if err != nil {
		return err
	}

	if int(stockID) != updateRequest.StockID {
		return fmt.Errorf("different stock id parameter in the authorization context and request data; authorization context: '%d', request: '%d'", int(stockID), updateRequest.StockID)
	}

	quantity, err := getFieldFromAuthorizationContext[float64](authContext, "quantity")
	if err != nil {
		return err
	}

	if int(quantity) != updateRequest.Quantity {
		return fmt.Errorf("different quantity parameter in the authorization context and request data; authorization context: '%d', request: '%d'", int(quantity), updateRequest.Quantity)
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

	purpose, ok := str2Purpose[token.Purpose]
	if !ok {
		return fmt.Errorf("invalid request purpose")
	}

	authContext, ok := token.AuthorizationContext.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid authorization context format")
	}

	switch {
	case pathTemplate == "/api/stocks/search" && r.Method == http.MethodGet:
		err = verifyStockSearchAuthorizationContexts(purpose, authContext, r)
	case pathTemplate == "/api/stocks/holdings" && r.Method == http.MethodGet:
		err = verifyStocksHoldingsAuthorizationContexts(purpose)
	case pathTemplate == "/api/stocks/{id}" && r.Method == http.MethodGet:
		err = verifyStocksDetailsAuthorizationContexts(purpose, authContext, r)
	case pathTemplate == "/internal/stocks" && r.Method == http.MethodPost:
		err = verifyUpdateUserStockAuthorizationContexts(purpose, authContext, r)
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

			logger.Info("Txn token verified successfully.", zap.Any("txn-token-id", token.Id))

			next.ServeHTTP(w, r)
		})
	}
}
