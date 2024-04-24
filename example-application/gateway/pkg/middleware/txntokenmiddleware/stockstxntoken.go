package txntokenmiddleware

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

type stockRouteType int

const (
	stocksRouteStockSearch stockRouteType = iota
	stocksRouteHoldingsDetails
	stocksRouteStockDetails
	stocksRouteUnknown
)

const (
	stockSearch           scope = "StockSearch"
	stockDetails          scope = "StockDetails"
	stocksHoldingsDetails scope = "StocksHoldingsDetails"
)

func getStocksRoute(r *http.Request) stockRouteType {
	regex := regexp.MustCompile(`^/api/stocks/(\d+)$`)

	switch {
	case r.URL.Path == "/api/stocks/search" && r.Method == http.MethodGet:
		return stocksRouteStockSearch
	case r.URL.Path == "/api/stocks/holdings" && r.Method == http.MethodGet:
		return stocksRouteHoldingsDetails
	case regex.MatchString(r.URL.Path) && r.Method == http.MethodGet:
		return stocksRouteStockDetails
	default:
		return stocksRouteUnknown
	}
}

func getStocksApiScope(r *http.Request) (scope, error) {
	routeType := getStocksRoute(r)

	switch routeType {
	case stocksRouteStockSearch:
		return stockSearch, nil
	case stocksRouteHoldingsDetails:
		return stocksHoldingsDetails, nil
	case stocksRouteStockDetails:
		return stockDetails, nil
	default:
		return "", errors.New("unexpected route")
	}
}

func getStocksApiRequestDetails(r *http.Request) (*requestDetails, error) {
	routeType := getStocksRoute(r)
	regex := regexp.MustCompile(`^/api/stocks/(\d+)$`)

	switch routeType {
	case stocksRouteStockSearch:
		return &requestDetails{
			Query: r.URL.Query().Get("query"),
		}, nil

	case stocksRouteHoldingsDetails:
		return &requestDetails{}, nil

	case stocksRouteStockDetails:
		stockID, err := strconv.Atoi(regex.FindStringSubmatch(r.URL.Path)[1])
		if err != nil {
			return &requestDetails{}, err
		}
		return &requestDetails{
			StockID: stockID,
		}, nil

	default:
		return &requestDetails{}, errors.New("unexpected route")
	}
}
