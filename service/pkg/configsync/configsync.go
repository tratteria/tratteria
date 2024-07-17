package configsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/generationrules/v1alpha1"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"
)

const (
	MAX_REGISTRATION_ATTEMPTS       = 5
	FAILED_HEARTBEAT_RETRY_INTERVAL = 5 * time.Second
	REGISTRATION_PATH               = "register"
	HEARTBEAT_PATH                  = "heartbeat"
)

type Client struct {
	webhookPort       int
	webhookIP         string
	tconfigdUrl       url.URL
	namespace         string
	generationRules   *v1alpha1.GenerationRulesImp
	heartbeatInterval time.Duration
	httpClient        *http.Client
	logger            *zap.Logger
}

func NewClient(webhookPort int, tconfigdUrl url.URL, tconfigdSpiffeId spiffeid.ID, namespace string, generationRules *v1alpha1.GenerationRulesImp, x509Source *workloadapi.X509Source, logger *zap.Logger) (*Client, error) {
	webhookIP, err := getLocalIP()
	if err != nil {
		return nil, err
	}

	tlsConfig := tlsconfig.MTLSClientConfig(x509Source, x509Source, tlsconfig.AuthorizeID(tconfigdSpiffeId))

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &Client{
		webhookPort:     webhookPort,
		webhookIP:       webhookIP,
		tconfigdUrl:     tconfigdUrl,
		namespace:       namespace,
		generationRules: generationRules,
		httpClient:      client,
		logger:          logger,
	}, nil
}

type registrationRequest struct {
	IPAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
	Namespace string `json:"namespace"`
}

type registrationResponse struct {
	HeartBeatIntervalMinutes int                              `json:"heartBeatIntervalMinutes"`
	GenerationRules          v1alpha1.GenerationRulesTconfigd `json:"generationRules"`
}

type heartBeatRequest struct {
	IPAddress      string `json:"ipAddress"`
	Port           int    `json:"port"`
	Namespace      string `json:"namespace"`
	RulesVersionID string `json:"rulesVersionId"`
}

func (c *Client) Start() error {
	if err := c.registerWithBackoff(); err != nil {
		return fmt.Errorf("failed to register with tconfigd: %w", err)
	}

	c.logger.Info("Successfully registered to tconfigd")

	c.logger.Info("Starting heartbeats to tconfigd...")

	go c.startHeartbeat()

	return nil
}

func (c *Client) registerWithBackoff() error {
	var attempt int

	for {
		if err := c.register(); err != nil {
			c.logger.Error("Registration failed", zap.Error(err))

			attempt++

			if attempt >= MAX_REGISTRATION_ATTEMPTS {
				return fmt.Errorf("max registration attempts reached: %w", err)
			}

			backoff := time.Duration(rand.Intn(1<<attempt)) * time.Second

			c.logger.Info("Retrying registration", zap.Duration("backoff", backoff), zap.Int("attempt", attempt))

			time.Sleep(backoff)

			continue
		}

		break
	}

	return nil
}

func (c *Client) register() error {
	registrationReq := registrationRequest{
		IPAddress: c.webhookIP,
		Port:      c.webhookPort,
		Namespace: c.namespace,
	}

	jsonData, err := json.Marshal(registrationReq)
	if err != nil {
		return fmt.Errorf("failed to marshal registration data: %w", err)
	}

	registerEndpoint := c.tconfigdUrl.ResolveReference(&url.URL{Path: REGISTRATION_PATH})

	req, err := http.NewRequest(http.MethodPost, registerEndpoint.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var registrationResp registrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&registrationResp); err != nil {
		return fmt.Errorf("failed to decode registration response: %w", err)
	}

	c.heartbeatInterval = time.Duration(registrationResp.HeartBeatIntervalMinutes) * time.Minute

	c.generationRules.UpdateCompleteRules(registrationResp.GenerationRules)

	return nil
}

func (c *Client) startHeartbeat() {
	heartbeatEndpoint := c.tconfigdUrl.ResolveReference(&url.URL{Path: HEARTBEAT_PATH})

	for {
		heartBeatReq := heartBeatRequest{
			IPAddress: c.webhookIP,
			Port:      c.webhookPort,
			Namespace: c.namespace,
		}

		heartBeatRequestJson, err := json.Marshal(heartBeatReq)
		if err != nil {
			c.logger.Error("Failed to marshal heartbeat request", zap.Error(err))
			time.Sleep(FAILED_HEARTBEAT_RETRY_INTERVAL)

			continue
		}

		req, err := http.NewRequest(http.MethodPost, heartbeatEndpoint.String(), bytes.NewBuffer(heartBeatRequestJson))
		if err != nil {
			c.logger.Error("Failed to create heartbeat request", zap.Error(err))
			time.Sleep(FAILED_HEARTBEAT_RETRY_INTERVAL)

			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Failed to send heartbeat", zap.Error(err))
			time.Sleep(FAILED_HEARTBEAT_RETRY_INTERVAL)

			continue
		} else {
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				c.logger.Error("Received non-ok heartbeat response", zap.Int("status", resp.StatusCode))
				time.Sleep(FAILED_HEARTBEAT_RETRY_INTERVAL)

				continue
			} else {
				c.logger.Info("Heartbeat sent successfully")
			}
		}

		time.Sleep(c.heartbeatInterval)
	}
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("couldn't obtain a webhook IP address")
}
