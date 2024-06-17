package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/common"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/config"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/service"
	"github.com/SGNL-ai/TraTs-Demo-Svcs/txn-token-service/pkg/txntokenerrors"

	"go.uber.org/zap"
)

type Handlers struct {
	Service *service.Service
	Config  *config.AppConfig
	Logger  *zap.Logger
}

func NewHandlers(service *service.Service, config *config.AppConfig, logger *zap.Logger) *Handlers {
	return &Handlers{
		Service: service,
		Config:  config,
		Logger:  logger,
	}
}

const (
	GRANT_TYPE = "urn:ietf:params:oauth:grant-type:token-exchange"
)

func (h *Handlers) GetJwksHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Get-Jwks request received.")

	jwks := h.Service.GetJwks()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		h.Logger.Error("Failed to encode response of a get-jwks request.", zap.Error(err))

		return
	}

	h.Logger.Info("Get-Jwks request processed successfully.")
}

func (h *Handlers) TokenEndpointHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Txn-Token request received.")

	if err := r.ParseForm(); err != nil {
		h.Logger.Info("Failed to parse the txn-token request.", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if r.FormValue("grant_type") != GRANT_TYPE {
		h.Logger.Error("Invalid grant type.", zap.String("grant-type", r.FormValue("grant_type")))
		http.Error(w, "Invalid grant type.", http.StatusUnprocessableEntity)

		return
	}

	subjectTokenType := common.Str2TokenType[r.FormValue("subject_token_type")]
	if subjectTokenType != common.OIDC_ID_TOKEN_TYPE && subjectTokenType != common.SELF_SIGNED_TOKEN_TYPE {
		h.Logger.Error("Invalid or unsupported subject token type.", zap.String("subject-token-type", string(subjectTokenType)))
		http.Error(w, "Invalid or unsupported subject token type. Only OIDC ID and self-signed tokens are supported.", http.StatusUnprocessableEntity)

		return
	}

	subjectToken := r.FormValue("subject_token")
	if subjectToken == "" {
		h.Logger.Error("Subject token not provided.")
		http.Error(w, "Subject token not provided.", http.StatusBadRequest)

		return
	}

	if r.FormValue("audience") != h.Config.Audience {
		h.Logger.Error("The requested audience is not supported by this txn-token service.", zap.String("audience", r.FormValue("audience")))
		http.Error(w, "The requested audience is not supported by this txn-token service.", http.StatusForbidden)

		return
	}

	requestedTokenType := common.Str2TokenType[r.FormValue("requested_token_type")]
	if requestedTokenType != common.TXN_TOKEN_TYPE {
		h.Logger.Error("Invalid requested token type.", zap.String("requested-token-type", string(requestedTokenType)))
		http.Error(w, "Invalid requested token type.", http.StatusUnprocessableEntity)

		return
	}

	requestDetailsEncoded := r.FormValue("request_details")

	requestDetailsJSON, err := base64.RawURLEncoding.DecodeString(requestDetailsEncoded)
	if err != nil {
		h.Logger.Error("Failed to base64url decode the request details", zap.Error(err))
		http.Error(w, "Invalid request details encoding", http.StatusBadRequest)

		return
	}

	var requestDetails common.RequestDetails

	if err := json.Unmarshal(requestDetailsJSON, &requestDetails); err != nil {
		h.Logger.Error("Failed to unmarshal request details from the request", zap.Error(err))
		http.Error(w, "Invalid request details format", http.StatusBadRequest)

		return
	}

	if err := requestDetails.Validate(); err != nil {
		h.Logger.Error("Invalid request details:", zap.Error(err))
		http.Error(w, "Invalid request details: "+err.Error(), http.StatusBadRequest)

		return
	}

	requestContextEncoded := r.FormValue("request_context")

	requestContextJSON, err := base64.RawURLEncoding.DecodeString(requestContextEncoded)
	if err != nil {
		h.Logger.Error("Failed to base64url decode the request context", zap.Error(err))
		http.Error(w, "Invalid request context encoding", http.StatusBadRequest)

		return
	}

	requestContext := make(map[string]any)

	if err := json.Unmarshal(requestContextJSON, &requestContext); err != nil {
		h.Logger.Error("Failed to unmarshal request context from the request", zap.Error(err))
		http.Error(w, "Invalid request context format", http.StatusBadRequest)

		return
	}

	txnTokenRequest := common.TokenRequest{
		RequestedTokenType: requestedTokenType,
		SubjectToken:       subjectToken,
		SubjectTokenType:   subjectTokenType,
		RequestDetails:     requestDetails,
		RequestContext:     requestContext,
	}

	txnTokenResponse, err := h.Service.GenerateTxnToken(r.Context(), &txnTokenRequest)
	if err != nil {
		h.Logger.Error("Error generating txn token.", zap.Error(err))

		switch err {
		case txntokenerrors.ErrParsingSubjectToken, txntokenerrors.ErrInvalidSubjectTokenClaims, txntokenerrors.ErrUnsupportedTokenType, txntokenerrors.ErrSubjectFieldNotFound:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case txntokenerrors.ErrAccessDenied:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(txnTokenResponse); err != nil {
		h.Logger.Error("Failed to encode the token response.", zap.Error(err))

		return
	}

	h.Logger.Info("Txn-Token request processed successfully.")
}
