package handler

import (
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/proxy"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupRoutes(cfg *config.GatewayConfig, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	stocksProxy := proxy.NewReverseProxy(cfg.StocksServiceURL, logger)

	router.PathPrefix("/stocks/").Handler(middleware.Authenticate(stocksProxy, logger))

	return router
}
