package configsync

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/keys"
	"github.com/gorilla/websocket"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

const (
	TCONFIGD_WEBSOCKET_PATH    = "ws"
	CONNECTION_INITIAL_BACKOFF = 1 * time.Second
	CONNECTION_MAX_BACKOFF     = 60 * time.Second
	CONNECTION_MAX_RETRIES     = 5
	WRITE_WAIT                 = 10 * time.Second
	PONG_WAIT                  = 60 * time.Second
	PING_PERIOD                = (PONG_WAIT * 9) / 10
	REQUEST_TIMEOUT            = 15 * time.Second
)

type Client struct {
	tconfigdHost     string
	tconfigdSpiffeId spiffeid.ID
	x509Source       *workloadapi.X509Source
	namespace        string
	generationRules  *v1alpha1.GenerationRulesImp
	logger           *zap.Logger
	conn             *websocket.Conn
	send             chan []byte
	done             chan struct{}
	closeOnce        sync.Once
}

type MessageType string

const (
	MessageTypeInitialRulesResponse                        MessageType = "INITIAL_RULES_RESPONSE"
	MessageTypeGetJWKSRequest                              MessageType = "GET_JWKS_REQUEST"
	MessageTypeGetJWKSResponse                             MessageType = "GET_JWKS_RESPONSE"
	MessageTypeTraTGenerationRuleUpsertRequest             MessageType = "TRAT_GENERATION_RULE_UPSERT_REQUEST"
	MessageTypeTraTGenerationRuleUpsertResponse            MessageType = "TRAT_GENERATION_RULE_UPSERT_RESPONSE"
	MessageTypeTratteriaConfigGenerationRuleUpsertRequest  MessageType = "TRATTERIA_CONFIG_GENERATION_RULE_UPSERT_REQUEST"
	MessageTypeTratteriaConfigGenerationRuleUpsertResponse MessageType = "TRATTERIA_CONFIG_GENERATION_RULE_UPSERT_RESPONSE"
	MessageTypeRuleReconciliationRequest                   MessageType = "RULE_RECONCILIATION_REQUEST"
	MessageTypeRuleReconciliationResponse                  MessageType = "RULE_RECONCILIATION_RESPONSE"
	MessageTypeUnknown                                     MessageType = "UNKNOWN"
)

type PingData struct {
	RuleHash string `json:"ruleHash"`
}

type Request struct {
	ID      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	ID      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Status  int             `json:"status"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type RegistrationRequest struct {
	Namespace string `json:"namespace"`
}

type AllActiveGenerationRules struct {
	GenerationRules *v1alpha1.GenerationRules `json:"generationRules"`
}

func NewClient(tconfigdHost string, tconfigdSpiffeId spiffeid.ID, namespace string, generationRules *v1alpha1.GenerationRulesImp, x509Source *workloadapi.X509Source, logger *zap.Logger) *Client {
	return &Client{
		tconfigdHost:     tconfigdHost,
		tconfigdSpiffeId: tconfigdSpiffeId,
		x509Source:       x509Source,
		namespace:        namespace,
		generationRules:  generationRules,
		logger:           logger,
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.conn.Close()
		close(c.send)
		close(c.done)
		c.logger.Info("Connection closed and resources released")
	})
}

func (c *Client) Start(ctx context.Context) error {
	backoff := CONNECTION_INITIAL_BACKOFF

	for retries := 0; retries < CONNECTION_MAX_RETRIES; retries++ {
		if retries > 0 {
			time.Sleep(backoff)

			backoff *= 2

			if backoff > CONNECTION_MAX_BACKOFF {
				backoff = CONNECTION_MAX_BACKOFF
			}

			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			backoff = backoff/2 + jitter
		}

		if err := c.connect(ctx); err != nil {
			c.logger.Error("Failed to connect to tconfigd. Retrying...", zap.Error(err), zap.Int("retry", retries+1))

			continue
		}

		c.logger.Info("Successfully connected to tconfigd.")

		c.done = make(chan struct{})
		c.send = make(chan []byte, 256)

		go c.readPump()
		go c.writePump()

		backoff = CONNECTION_INITIAL_BACKOFF
		retries = 0

		select {
		case <-c.done:
			c.logger.Info("Connection closed. Attempting to reconnect...")
		case <-ctx.Done():
			c.logger.Info("Context cancelled, shutting down config sync client...")

			c.close()

			return ctx.Err()
		}
	}

	c.logger.Info("Max retries reached. Shutting down...")

	return fmt.Errorf("max retries reached; shutting down")
}

func (c *Client) connect(ctx context.Context) error {
	wsURL := url.URL{
		Scheme:   "wss",
		Host:     c.tconfigdHost,
		Path:     TCONFIGD_WEBSOCKET_PATH,
		RawQuery: url.Values{"namespace": {c.namespace}}.Encode(),
	}

	tlsConfig := tlsconfig.MTLSClientConfig(c.x509Source, c.x509Source, tlsconfig.AuthorizeID(c.tconfigdSpiffeId))

	dialer := websocket.Dialer{
		TLSClientConfig: tlsConfig,
	}

	c.logger.Info("Connecting to tconfigd's WebSocket server.", zap.String("url", wsURL.String()))

	conn, _, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		c.logger.Error("Failed to connect to tconfigd's websocket server.", zap.Error(err))

		return fmt.Errorf("failed to connect to tconfigd's websocket server: %w", err)
	}

	c.conn = conn

	c.logger.Info("Successfully connected to tconfigd's websocket server.", zap.String("url", wsURL.String()))

	_, message, err := conn.ReadMessage()
	if err != nil {
		c.logger.Error("Failed to read initial configuration message", zap.Error(err))

		conn.Close()

		return fmt.Errorf("failed to read initial configuration: %w", err)
	}

	var initialRuleResponse Response

	err = json.Unmarshal(message, &initialRuleResponse)
	if err != nil {
		c.logger.Error("Failed to unmarshal initial rules response", zap.Error(err))

		conn.Close()

		return fmt.Errorf("failed to unmarshal initial rules response: %w", err)
	}

	if initialRuleResponse.Type != MessageTypeInitialRulesResponse {
		c.logger.Error("Unexpected message type for initial rules response", zap.String("type", string(initialRuleResponse.Type)))

		conn.Close()

		return fmt.Errorf("unexpected message type for initial rules response: %s", initialRuleResponse.Type)
	}

	if initialRuleResponse.Status != http.StatusCreated {
		c.logger.Error("Received unexpected status code for initial rules response.", zap.Int("status", initialRuleResponse.Status), zap.ByteString("response", initialRuleResponse.Payload))

		conn.Close()

		return fmt.Errorf("received unexpected status code for initial rules response: %v", initialRuleResponse.Status)
	}

	var initialGenerationRulesResponsePayload AllActiveGenerationRules

	err = json.Unmarshal(initialRuleResponse.Payload, &initialGenerationRulesResponsePayload)
	if err != nil {
		c.logger.Error("Failed to unmarshal initial generation rules response payload", zap.Error(err))

		conn.Close()

		return fmt.Errorf("failed to unmarshal initial generation rules response payload: %w", err)
	}

	if initialGenerationRulesResponsePayload.GenerationRules == nil {
		c.logger.Error("Received empty initial generation rules")

		conn.Close()

		return fmt.Errorf("received empty initial generation rules")
	}

	c.generationRules.UpdateCompleteRules(initialGenerationRulesResponsePayload.GenerationRules)

	c.logger.Info("Received and applied initial generation rules")

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))

		return nil
	})

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("WebSocket connection closed unexpectedly.", zap.Error(err))
				}

				return
			}

			c.handleMessage(message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(PING_PERIOD)

	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.send:
			if !ok {
				return
			}

			if err := c.writeMessage(websocket.TextMessage, message); err != nil {
				c.logger.Error("Failed to write message.", zap.Error(err))

				return
			}
		case <-ticker.C:
			generationHash, err := c.generationRules.GetGenerationRulesHash()
			if err != nil {
				c.logger.Error("Error getting generation rule hash.", zap.Error(err))

				return
			}

			pingData := PingData{
				RuleHash: generationHash,
			}

			pingPayload, err := json.Marshal(pingData)
			if err != nil {
				c.logger.Error("Failed to marshal ping data", zap.Error(err))

				return
			}

			if err := c.writeMessage(websocket.PingMessage, pingPayload); err != nil {
				c.logger.Error("Failed to write ping message.", zap.Error(err))

				return
			}
		}
	}
}

func (c *Client) writeMessage(messageType int, data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

	return c.conn.WriteMessage(messageType, data)
}

func (c *Client) handleMessage(message []byte) {
	var temp struct {
		Type MessageType `json:"type"`
	}

	if err := json.Unmarshal(message, &temp); err != nil {
		c.logger.Error("Failed to unmarshal message type.", zap.Error(err))

		return
	}

	switch temp.Type {
	case MessageTypeTraTGenerationRuleUpsertRequest,
		MessageTypeTratteriaConfigGenerationRuleUpsertRequest,
		MessageTypeGetJWKSRequest,
		MessageTypeRuleReconciliationRequest:
		c.handleRequest(message)
	default:
		c.logger.Error("Received unknown or unexpected message type.", zap.String("type", string(temp.Type)))
	}
}

func (c *Client) handleRequest(message []byte) {
	var request Request
	if err := json.Unmarshal(message, &request); err != nil {
		c.logger.Error("Failed to unmarshal request", zap.Error(err))

		return
	}

	c.logger.Debug("Received request", zap.String("id", request.ID), zap.String("type", string(request.Type)))

	switch request.Type {
	case MessageTypeTraTGenerationRuleUpsertRequest, MessageTypeTratteriaConfigGenerationRuleUpsertRequest:
		c.handleRuleUpsertRequest(request)
	case MessageTypeGetJWKSRequest:
		c.handleGetJWKSRequest(request)
	case MessageTypeRuleReconciliationRequest:
		c.handleRuleReconciliationRequest(request)
	default:
		c.logger.Error("Received unknown or unexpected request type", zap.String("type", string(request.Type)))

		c.sendErrorResponse(
			request.ID,
			MessageTypeUnknown,
			http.StatusBadRequest,
			"received unknown or unexpected request type",
		)
	}
}

func (c *Client) handleRuleUpsertRequest(request Request) {
	switch request.Type {
	case MessageTypeTraTGenerationRuleUpsertRequest:
		var traTGenerationRule v1alpha1.TraTGenerationRule

		if err := json.Unmarshal(request.Payload, &traTGenerationRule); err != nil {
			c.logger.Error("Failed to unmarshal trat generation rule", zap.Error(err))
			c.sendErrorResponse(
				request.ID,
				MessageTypeTraTGenerationRuleUpsertResponse,
				http.StatusBadRequest,
				"error parsing trat generation rule",
			)

			return
		}

		c.logger.Info("Received trat generation rule upsert request",
			zap.String("endpoint", traTGenerationRule.Endpoint),
			zap.Any("method", traTGenerationRule.Method))

		err := c.generationRules.UpsertTraTRule(traTGenerationRule)
		if err != nil {
			c.logger.Error("Failed to upsert trat generation rule", zap.Error(err))
			c.sendErrorResponse(
				request.ID,
				MessageTypeTraTGenerationRuleUpsertResponse,
				http.StatusInternalServerError,
				"error upserting trat generation rule",
			)

			return
		}

		err = c.sendResponse(request.ID, MessageTypeTraTGenerationRuleUpsertResponse, http.StatusOK, nil)
		if err != nil {
			c.logger.Error("Error sending trat generation upsert request response", zap.Error(err))
		}

	case MessageTypeTratteriaConfigGenerationRuleUpsertRequest:
		var tratteriaConfigGenerationRule v1alpha1.TratteriaConfigGenerationRule

		if err := json.Unmarshal(request.Payload, &tratteriaConfigGenerationRule); err != nil {
			c.logger.Error("Failed to unmarshal tratteria config generation rule", zap.Error(err))
			c.sendErrorResponse(
				request.ID,
				MessageTypeTratteriaConfigGenerationRuleUpsertResponse,
				http.StatusBadRequest,
				"error parsing tratteria config generation rule",
			)

			return
		}

		c.logger.Info("Received tratteria config generation rule upsert request")

		c.generationRules.UpdateTratteriaConfigRule(tratteriaConfigGenerationRule)

		err := c.sendResponse(request.ID, MessageTypeTratteriaConfigGenerationRuleUpsertResponse, http.StatusOK, nil)
		if err != nil {
			c.logger.Error("Error sending trat generation upsert request response", zap.Error(err))
		}
	default:
		c.logger.Error("Received unknown or unexpected rule upsert request", zap.String("type", string(request.Type)))

		c.sendErrorResponse(
			request.ID,
			MessageTypeUnknown,
			http.StatusBadRequest,
			"received unknown or unexpected rule upsert request",
		)
	}
}

func (c *Client) handleGetJWKSRequest(request Request) {
	jwks := keys.GetJWKS()

	err := c.sendResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusOK, jwks)
	if err != nil {
		c.logger.Error("Error sending JWKS", zap.Error(err))
	}
}

func (c *Client) handleRuleReconciliationRequest(request Request) {
	c.logger.Info("Received generation rules reconciliation request")

	var allActiveGenerationRules AllActiveGenerationRules

	if err := json.Unmarshal(request.Payload, &allActiveGenerationRules); err != nil {
		c.logger.Error("Error parsing generation rule reconciliation request", zap.Error(err))
		c.sendErrorResponse(
			request.ID,
			MessageTypeRuleReconciliationResponse,
			http.StatusBadRequest,
			"error parsing generation rule reconciliation request",
		)

		return
	}

	c.generationRules.UpdateCompleteRules(allActiveGenerationRules.GenerationRules)

	err := c.sendResponse(request.ID, MessageTypeRuleReconciliationResponse, http.StatusOK, nil)
	if err != nil {
		c.logger.Error("Error sending generation rule reconciliation request response", zap.Error(err))
	}
}

func (c *Client) sendResponse(id string, respType MessageType, status int, payload interface{}) error {
	var payloadJSON json.RawMessage

	if payload != nil {
		var err error

		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal response payload: %w", err)
		}
	}

	response := Response{
		ID:      id,
		Type:    respType,
		Status:  status,
		Payload: payloadJSON,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	select {
	case c.send <- responseJSON:
		return nil
	default:
		return fmt.Errorf("send channel is full")
	}
}

func (c *Client) sendErrorResponse(requestID string, messageType MessageType, statusCode int, errorMessage string) {
	err := c.sendResponse(requestID, messageType, statusCode, map[string]string{"error": errorMessage})
	if err != nil {
		c.logger.Error("Failed to send error response",
			zap.String("request-id", requestID),
			zap.Error(err))
	}
}
