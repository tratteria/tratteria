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
	Success      bool                      `json:"success"`
	Query        string                    `json:"query"`
	Limit        int                       `json:"limit"`
	TotalResults int                       `json:"totalResults"`
	Results      []service.StockSearchItem `json:"results"`
}

func (h *Handlers) SearchStocks(w http.ResponseWriter, r *http.Request) {
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
		Success:      true,
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

	h.Logger.Info("A stock-search request processed successfully.", zap.String("query", query), zap.Int("limit", limit))
}

func (h *Handlers) GetStockDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.Logger.Error("Invalid stock id provided in a get-stock-details request.", zap.String("id", idStr))
		http.Error(w, "Invalid stock ID.", http.StatusBadRequest)

		return
	}

	h.Logger.Info("A get-stock-details request received.", zap.Int("id", id))

	stock, err := h.Service.GetStockDetails(id)
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
		http.Error(w, "Failed to encode response of a get-stock-details request.", http.StatusInternalServerError)

		return
	}

	h.Logger.Info("A get-stock-details request processed successfully.", zap.Int("id", id))
}
