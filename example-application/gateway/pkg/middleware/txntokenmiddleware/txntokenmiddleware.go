package txntokenmiddleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"net"
	"strings"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/gateway/pkg/common"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

type gatewayRouteType int

const (
	stocksRoute gatewayRouteType = iota
	orderRoute
	unknownRoute
)

const (
	OIDC_ID_TOKEN_TYPE = "urn:ietf:params:oauth:token-type:id_token"
	TXN_TOKEN_TYPE     = "urn:ietf:params:oauth:token-type:txn_token"
	GRANT_TYPE         = "urn:ietf:params:oauth:grant-type:token-exchange"
)

const (
	AUDIENCE = "https://alphastocks.com/"
)

type scope string

type requestDetails struct {
	Action        tradeAction `json:"action,omitempty"`
	Query         string      `json:"query,omitempty"`
	StockID       int         `json:"stockID,omitempty"`
	TransactionID string      `json:"transactionID,omitempty"`
	Quantity      int         `json:"quantity,omitempty"`
}

type txnToken struct {
	IssuedTokenType string `json:"issued_token_type"`
	AccessToken     string `json:"access_token"`
}

func getRequesterIP(r *http.Request) (string, error) {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if validIP := net.ParseIP(ip); validIP != nil {
				return ip, nil
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if validIP := net.ParseIP(r.RemoteAddr); validIP != nil {
			return r.RemoteAddr, nil
		}

		return "", fmt.Errorf("failed to parse RemoteAddr: %v", err)
	}

	if validIP := net.ParseIP(ip); validIP != nil {
		return ip, nil
	}

	return "", fmt.Errorf("no valid IP address found")
}

func getRequestContext(r *http.Request) (map[string]interface{}, error) {
	ip, err := getRequesterIP(r)
	if err != nil {
		return nil, err
	}

	request_context := make(map[string]interface{})

	request_context["req_ip"] = ip

	return request_context, nil
}

func getGatewayRoute(r *http.Request) (gatewayRouteType, error) {
	pathTemplate, err := mux.CurrentRoute(r).GetPathTemplate()
	if err != nil {
		return unknownRoute, err
	}

	switch pathTemplate {
	case "/api/stocks":
		return stocksRoute, nil
	case "/api/order":
		return orderRoute, nil
	default:
		return unknownRoute, nil
	}
}

func getRequestDetails(r *http.Request) (*requestDetails, error) {
	routeType, err := getGatewayRoute(r)
	if err != nil {
		return &requestDetails{}, nil
	}

	switch routeType {
	case stocksRoute:
		return getStocksApiRequestDetails(r)
	case orderRoute:
		return getOrderApiRequestDetails(r)
	default:
		return &requestDetails{}, fmt.Errorf("unexpected route: %v", routeType)
	}

}

func getScope(r *http.Request) (scope, error) {
	routeType, err := getGatewayRoute(r)
	if err != nil {
		return "", nil
	}

	switch routeType {
	case stocksRoute:
		return getStocksApiScope(r)
	case orderRoute:
		return getOrderApiScope(r)
	default:
		return "", fmt.Errorf("unexpected route: %v", routeType)
	}

}

func GetTxnTokenMiddleware(txnTokenServiceURL string, httpClient *http.Client, spireJwtSource *workloadapi.JWTSource, txnTokenServiceSpiffeID spiffeid.ID, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestDetails, err := getRequestDetails(r)
			if err != nil {
				logger.Error("Error creating request details.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			scope, err := getScope(r)
			if err != nil {
				logger.Error("Error getting request score.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			oidcIdToken, ok := r.Context().Value(common.OIDC_ID_TOKEN_CONTEXT_KEY).(string)
			if !ok {
				logger.Error("Failed to retrieve OIDC id_token.")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			requestDetailsJSON, err := json.Marshal(requestDetails)
			if err != nil {
				logger.Error("Failed to marshal request details.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			encodedRequestDetails := base64.RawURLEncoding.EncodeToString(requestDetailsJSON)

			requestContext, err := getRequestContext(r)
			if err != nil {
				logger.Error("Error generating request context.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			requestContextJSON, err := json.Marshal(requestContext)
			if err != nil {
				logger.Error("Failed to marshal request context.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			encodedRequestContext := base64.RawURLEncoding.EncodeToString(requestContextJSON)

			requestData := url.Values{}
			requestData.Set("grant_type", GRANT_TYPE)
			requestData.Set("requested_token_type", TXN_TOKEN_TYPE)
			requestData.Set("audience", AUDIENCE)
			requestData.Set("scope", string(scope))
			requestData.Set("subject_token", oidcIdToken)
			requestData.Set("subject_token_type", OIDC_ID_TOKEN_TYPE)
			requestData.Set("request_details", encodedRequestDetails)
			requestData.Set("request_context", encodedRequestContext)

			req, _ := http.NewRequest("POST", txnTokenServiceURL+"/token_endpoint", bytes.NewBufferString(requestData.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
			defer cancel()

			svid, err := spireJwtSource.FetchJWTSVID(ctx, jwtsvid.Params{
				Audience: txnTokenServiceSpiffeID.String(),
			})
			if err != nil {
				logger.Error("Failed to fetch JWT-SVID.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			req.Header.Set("Authorization", "Bearer "+svid.Marshal())

			resp, err := httpClient.Do(req)
			if err != nil {
				logger.Error("Failed to request txn token from token service.", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Error("Non-OK HTTP status received from token service.", zap.Int("status", resp.StatusCode))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Failed to read the response from token service", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			var token txnToken
			if err := json.Unmarshal(body, &token); err != nil {
				logger.Error("Failed to parse transaction token", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			if token.IssuedTokenType != TXN_TOKEN_TYPE {
				logger.Error("Issued invalid token type in txn-token response.", zap.String("token-type", string(token.IssuedTokenType)))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			if token.AccessToken == "" {
				logger.Error("Received empty access token from token service")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			r.Header.Set("Txn-Token", token.AccessToken)

			next.ServeHTTP(w, r)
		})
	}
}
