package txntokenmiddleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

type orderRouteType int

const (
	orderRouteStocksTrade orderRouteType = iota
	orderRouteTradeTransactionDetails
	orderRouteUnknown
)

const (
	tradeTransactionDetails scope = "TradeTransactionDetails"
	stocksTrade             scope = "StocksTrade"
)

type tradeAction string

const (
	Buy  tradeAction = "Buy"
	Sell tradeAction = "Sell"
)

func getOrderRoute(r *http.Request) orderRouteType {
	regex := regexp.MustCompile(`^/api/order/([a-zA-Z0-9-_\.]+)$`)

	switch {
	case r.URL.Path == "/api/order" && r.Method == http.MethodPost:
		return orderRouteStocksTrade
	case regex.MatchString(r.URL.Path) && r.Method == http.MethodGet:
		return orderRouteTradeTransactionDetails
	default:
		return orderRouteUnknown
	}
}

func getOrderApiScope(r *http.Request) (scope, error) {
	routeType := getOrderRoute(r)

	switch routeType {
	case orderRouteStocksTrade:
		return stocksTrade, nil
	case orderRouteTradeTransactionDetails:
		return tradeTransactionDetails, nil
	default:
		return "", errors.New("unexpected route")
	}
}

func getOrderApiRequestDetails(r *http.Request) (*requestDetails, error) {
	routeType := getOrderRoute(r)
	regex := regexp.MustCompile(`^/api/order/([a-zA-Z0-9-_\.]+)$`)

	switch routeType {
	case orderRouteStocksTrade:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var requestBody struct {
			Action   tradeAction `json:"orderType"`
			StockID  int         `json:"stockID"`
			Quantity int         `json:"quantity"`
		}

		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			return &requestDetails{}, err
		}

		return &requestDetails{
			Action:   requestBody.Action,
			StockID:  requestBody.StockID,
			Quantity: requestBody.Quantity,
		}, nil

	case orderRouteTradeTransactionDetails:
		return &requestDetails{
			TransactionID: regex.FindStringSubmatch(r.URL.Path)[1],
		}, nil

	default:
		return &requestDetails{}, errors.New("unexpected route")
	}
}
