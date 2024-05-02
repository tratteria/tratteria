package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/ordererrors"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/order/pkg/service"
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

type OrderRequest struct {
	OrderType service.OrderType `json:"orderType"`
	StockID   int               `json:"stockID"`
	Quantity  int               `json:"quantity"`
}

func (h *Handlers) OrderHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Order request received.")

	var orderRequest OrderRequest

	username := r.Header.Get("alpha-stock-user-name")
	if username == "" {
		h.Logger.Error("Unable to extract username from the header of the order request.")
		http.Error(w, "Unable to extract username from the header", http.StatusInternalServerError)

		return
	}

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		h.Logger.Error("Failed to decode order request body.", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	orderDetails, err := h.Service.Order(r.Context(), username, orderRequest.StockID, orderRequest.OrderType, orderRequest.Quantity)
	if err != nil {
		h.Logger.Error("Failed to process stock order.", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(orderDetails); err != nil {
		h.Logger.Error("Failed to encode response of a stock-order request.", zap.Error(err))

		return
	}

	h.Logger.Info("Order request processed successfully.")
}

func (h *Handlers) GetOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("alpha-stock-user-name")
	if username == "" {
		h.Logger.Error("Unable to extract username from the header of the order request.")
		http.Error(w, "Unable to extract username from the header", http.StatusInternalServerError)

		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	h.Logger.Info("A get-order-details request received.", zap.String("id", id))

	stock, err := h.Service.GetOrderDetails(username, id)
	if err != nil {
		if errors.Is(err, ordererrors.ErrOrderNotFound) {
			h.Logger.Error("Order not found", zap.String("id", id))
			http.Error(w, "Order not found", http.StatusNotFound)

			return
		}

		h.Logger.Error("Error encountered in a get-order-details request.", zap.String("id", id))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stock); err != nil {
		h.Logger.Error("Failed to encode response of a get-order-details request.", zap.Error(err))

		return
	}

	h.Logger.Info("Get-order-details request processed successfully.", zap.String("id", id))
}
