package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/proxy"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Username string `json:"username"`
}

func SetupRoutes(cfg *config.GatewayConfig, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	stocksProxy := proxy.NewReverseProxy(cfg.StocksServiceURL, logger)

	router.PathPrefix("/stocks/").Handler(middleware.Authenticate(stocksProxy, logger))

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, logger)
	}).Methods("POST")

	return router
}

func handleLogin(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	var loginRequest LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)

		logger.Error("Unable to parse a login request.", zap.Error(err))
		return
	}

	username := loginRequest.Username
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)

		logger.Error("Username is missing in a login request.")
		return
	}

	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:    "session_token",
		Value:   username,
		Expires: expiration,
		Path:    "/",
	}
	http.SetCookie(w, &cookie)

	logger.Info("User logged in", zap.String("username", username))
	w.WriteHeader(http.StatusOK)
}
