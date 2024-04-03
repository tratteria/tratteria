package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/stocks/pkg/service"
	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

type Handlers struct {
	Service *service.Service
	Logger  *zap.Logger
}

func NewHandlers(service *service.Service, logger *zap.Logger) *Handlers {
	return &Handlers{
		Service: service,
		Logger:  logger,
	}
}

type StocksSearchResponse struct {
	Query        string                    `json:"query"`
	Limit        int                       `json:"limit"`
	TotalResults int                       `json:"totalResults"`
	Results      []service.StockSearchItem `json:"results"`
}

type UpdateRequest struct {
	OrderType service.UpdateType `json:"orderType"`
	StockID   int                `json:"stockID"`
	Quantity  int                `json:"quantity"`
}

func (h *Handlers) SearchStocksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		h.Logger.Error("Missing search query parameter in a stock-search request.")
		http.Error(w, "Search query parameter 'query' is missing.", http.StatusBadRequest)

		return
	}

	var limit int

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = 10
	} else {
		var err error

		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			h.Logger.Error("Invalid max-search-items limit value in a stock-search request.", zap.String("limit", limitStr))
			http.Error(w, "Invalid limit value.", http.StatusBadRequest)

			return
		}
	}

	h.Logger.Info("A stock-search request received.", zap.String("query", query), zap.Int("limit", limit))

	stocks, err := h.Service.SearchStocks(query, limit)
	if err != nil {
		h.Logger.Error("Error encountered in a stock-search request.", zap.Error(err))
		http.Error(w, "Internal server error.", http.StatusInternalServerError)

		return
	}

	response := StocksSearchResponse{
		Query:        query,
		Limit:        limit,
		TotalResults: len(stocks),
		Results:      stocks,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Failed to encode JSON response of a stock-search request.", zap.Error(err))

		return
	}

	h.Logger.Info("Stock-search request processed successfully.", zap.String("query", query), zap.Int("limit", limit))
}

func (h *Handlers) GetStockDetailsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("x-user-name")
	if username == "" {
		h.Logger.Error("Unable to extract username from the header of stock-search request.")
		http.Error(w, "Unable to extract username header.", http.StatusInternalServerError)

		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.Logger.Error("Invalid stock id provided in a get-stock-details request.", zap.String("id", idStr))
		http.Error(w, "Invalid stock ID.", http.StatusBadRequest)

		return
	}

	h.Logger.Info("A get-stock-details request received.", zap.Int("id", id))

	stock, err := h.Service.GetStockDetails(username, id)
	if err != nil {
		if errors.Is(err, service.ErrStockNotFound) {
			h.Logger.Error("Stock not found", zap.Int("id", id))
			http.Error(w, "Stock not found", http.StatusNotFound)

			return
		}

		h.Logger.Error("Error encountered in a get-stock-details request.", zap.String("id", idStr))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stock); err != nil {
		h.Logger.Error("Failed to encode response of a get-stock-details request.", zap.Error(err))

		return
	}

	h.Logger.Info("Get-stock-details request processed successfully.", zap.Int("id", id))
}

func (h *Handlers) UpdateUserStockHandler(w http.ResponseWriter, r *http.Request) {
	var updateRequest UpdateRequest

	username := r.Header.Get("x-user-name")
	if username == "" {
		h.Logger.Error("Unable to extract username from the header of the update-user-stock request.")
		http.Error(w, "Unable to extract username from the header", http.StatusInternalServerError)

		return
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		h.Logger.Error("Failed to decode update-user-stock request body.", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	if updateRequest.Quantity < 1 {
		h.Logger.Error("Invalid update quantity: update quantity less than one.")
		http.Error(w, "Invalid update quantity: update quantity less than one", http.StatusBadRequest)

		return
	}

	stock, err := h.Service.GetStockDetails(username, updateRequest.StockID)
	if err != nil {
		if errors.Is(err, service.ErrStockNotFound) {
			h.Logger.Error("Stock provided in update-user-stock request not found.", zap.Int("stock-id", updateRequest.StockID))
			http.Error(w, "Stock not found", http.StatusBadRequest)

			return
		}

		h.Logger.Error("Error encountered in a update-user-stock request.", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	if updateRequest.OrderType == service.Sell && stock.Holdings < updateRequest.Quantity {
		h.Logger.Error("Stock sell quantity is more than the stock holdings.", zap.Int("total-holdings", stock.Holdings), zap.Int("sell-quantity", updateRequest.Quantity))
		http.Error(w, "Stock sell quantity is more than the stock holdings", http.StatusBadRequest)

		return
	}

	if updateRequest.OrderType == service.Buy && stock.TotalAvailableShares < updateRequest.Quantity {
		h.Logger.Error("Ordered quantity is more than available stocks.", zap.Int("total-available", stock.TotalAvailableShares), zap.Int("ordered-quantity", updateRequest.Quantity))
		http.Error(w, "Ordered quantity is more than available stocks", http.StatusBadRequest)

		return
	}

	orderDetails, err := h.Service.UpdateUserStock(username, stock, updateRequest.OrderType, updateRequest.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUpdateRequest) {
			h.Logger.Error("Invalid update request.", zap.Any("update-request", updateRequest))
			http.Error(w, "Invalid update request", http.StatusBadRequest)

			return
		}

		h.Logger.Error("Failed to process stock order.", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderDetails); err != nil {
		h.Logger.Error("Failed to encode response of a stock-order request.", zap.Error(err))

		return
	}

	h.Logger.Info("Stock-update request processed successfully.")
}

func (h *Handlers) GetUserHoldingsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("x-user-name")
	if username == "" {
		h.Logger.Error("Unable to extract username from the header for a get-user-holdings request.")
		http.Error(w, "Unable to extract username from the header", http.StatusInternalServerError)

		return
	}

	holdings, err := h.Service.GetUserHoldings(username)
	if err != nil {
		h.Logger.Error("Error encountered in a get-user-holdings request.", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(holdings); err != nil {
		h.Logger.Error("Failed to encode response of a get-user-holdings request.", zap.Error(err))

		return
	}

	h.Logger.Info("Get-user-holdings request processed successfully.")
}
