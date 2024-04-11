package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/middleware"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/proxy"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/sessionmanager"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type DexCodeExchangeRequest struct {
	Code string `json:"code"`
}

type IDTokenClaims struct {
	Email string `json:"email"`
}

func SetupRoutes(cfg *config.GatewayConfig, oauth2Config oauth2.Config, oidcProvider *oidc.Provider, spireJwtSource *workloadapi.JWTSource, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	stocksProxy := proxy.NewReverseProxy(cfg.StocksServiceURL, logger)
	orderProxy := proxy.NewReverseProxy(cfg.OrderServiceURL, logger)

	router.PathPrefix("/api/stocks").Handler(middleware.GatewayMiddleware(stocksProxy, cfg.SpiffeIDs.Stocks, spireJwtSource, logger))
	router.PathPrefix("/api/order").Handler(middleware.GatewayMiddleware(orderProxy, cfg.SpiffeIDs.Order, spireJwtSource, logger))

	router.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		handleLogout(w, r, logger)
	}).Methods("POST")

	router.HandleFunc("/api/exchange-code", func(w http.ResponseWriter, r *http.Request) {
		handleOidcCodeExchange(w, r, logger, oauth2Config, oidcProvider)
	}).Methods("POST")

	return router
}

func handleLogout(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionmanager.DeleteSession(cookie.Value)
		
		logger.Info("User logged out", zap.String("email", cookie.Value))
	}

	expiration := time.Now().Add(-24 * time.Hour)
	invalidated_cookie := http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &invalidated_cookie)

	w.WriteHeader(http.StatusOK)
}

func handleOidcCodeExchange(w http.ResponseWriter, r *http.Request, logger *zap.Logger, oauth2Config oauth2.Config, oidcProvider *oidc.Provider) {
	var dexCodeExchangeRequest DexCodeExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&dexCodeExchangeRequest); err != nil {
		logger.Error("Invalid to OIDC code exchange request.", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	oauth2Token, err := oauth2Config.Exchange(ctx, dexCodeExchangeRequest.Code)
	if err != nil {
		logger.Error("Failed to exchange the authorization code for a token.", zap.Error(err))
		http.Error(w, "Failed to exchange the authorization code for a token", http.StatusInternalServerError)

		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		logger.Error("ID Token not found in the OAuth2Token.")
		http.Error(w, "ID Token not found", http.StatusInternalServerError)

		return
	}

	oidcConfig := &oidc.Config{
		ClientID: oauth2Config.ClientID,
	}
	verifier := oidcProvider.Verifier(oidcConfig)

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.Error("Failed to verify OIDC ID token.", zap.Error(err))
		http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)

		return
	}

	var claims IDTokenClaims

	if err := idToken.Claims(&claims); err != nil {
		logger.Error("Failed to parse OIDC ID token claims.", zap.Error(err))
		http.Error(w, "Failed to parse ID token claims", http.StatusInternalServerError)

		return
	}

	logger.Info("OIDC ID Token verified successfully.", zap.String("email", claims.Email))

	expiration := time.Now().Add(24 * time.Hour)
	userSession := sessionmanager.UserSession{
		Email:   claims.Email,
		Expires: expiration,
	}

	sessionID := sessionmanager.SaveSession(userSession)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
	})

	logger.Info("User session created", zap.String("email", claims.Email))
	w.WriteHeader(http.StatusOK)
}
