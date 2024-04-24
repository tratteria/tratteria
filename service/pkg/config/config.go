package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"gopkg.in/yaml.v2"
)

const CONFIG_FILE_PATH = "/app/config/config.yaml"

type AppConfig struct {
	Issuer                      string                       `yaml:"issuer"`
	Audience                    string                       `yaml:"audience"`
	Token                       Token                        `yaml:"token"`
	Spiffe                      *Spiffe                      `yaml:"spiffe"`
	ClientAuthenticationMethods *ClientAuthenticationMethods `yaml:"clientAuthenticationMethods"`
	Keys                        *Keys
}

type Token struct {
	LifeTime time.Duration `yaml:"lifeTime"`
}

func (t *Token) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp struct {
		LifeTime string `yaml:"lifeTime"`
	}

	if err := unmarshal(&tmp); err != nil {
		return err
	}

	duration, err := time.ParseDuration(tmp.LifeTime)
	if err != nil {
		return fmt.Errorf("error parsing token lifetime: %v", err)
	}

	t.LifeTime = duration

	return nil
}

type Spiffe struct {
	EndpointSocket       string        `yaml:"endpoint_socket"`
	ServiceID            spiffeid.ID   `yaml:"serviceID"`
	AuthorizedServiceIDs []spiffeid.ID `yaml:"authorizedServiceIDs"`
}

func (s *Spiffe) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		EndpointSocket       string   `yaml:"endpoint_socket"`
		ServiceID            string   `yaml:"serviceID"`
		AuthorizedServiceIDs []string `yaml:"authorizedServiceIDs"`
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	s.EndpointSocket = raw.EndpointSocket

	serviceID, err := spiffeid.FromString(raw.ServiceID)
	if err != nil {
		return fmt.Errorf("invalid spiffe ServiceID: %v", err)
	}

	s.ServiceID = serviceID

	for _, idStr := range raw.AuthorizedServiceIDs {
		id, err := spiffeid.FromString(idStr)
		if err != nil {
			return fmt.Errorf("invalid spiffe AuthorizedServiceID: %v", err)
		}

		s.AuthorizedServiceIDs = append(s.AuthorizedServiceIDs, id)
	}

	return nil
}

type ClientAuthenticationMethods struct {
	OIDC *OIDC `yaml:"OIDC"`
}

type OIDC struct {
	ClientID     string `yaml:"clientId"`
	ProviderURL  string `yaml:"providerURL"`
	SubjectField string `yaml:"subjectField"`
}

type Keys struct {
	PrivateKey string
	JWKS       string
	KeyID      string
}

func GetAppConfig() *AppConfig {
	data, err := os.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config file: %v", err))
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal YAML configuration: %v", err))
	}

	validateConfig(&cfg)

	cfg.Keys = loadKeys()

	return &cfg
}

func validateConfig(cfg *AppConfig) {
	if cfg.Issuer == "" {
		panic("Issuer must not be empty")
	}

	if cfg.Audience == "" {
		panic("Audience must not be empty")
	}

	if len(cfg.Spiffe.AuthorizedServiceIDs) == 0 {
		panic("Authorized services spifee ids must be specified")
	}

	validateOIDC(cfg.ClientAuthenticationMethods.OIDC)
}

func validateOIDC(oidc *OIDC) {
	if oidc == nil {
		panic("OIDC configuration must be provided")
	}

	if oidc.ClientID == "" {
		panic("OIDC Client ID must be populated")
	}

	if oidc.ProviderURL == "" {
		panic("OIDC Provider URL must be populated")
	}

	if oidc.SubjectField == "" {
		panic("OIDC Subject Field must be populated")
	}
}

func loadKeys() *Keys {
	return &Keys{
		PrivateKey: getEnv("PRIVATE_KEY"),
		JWKS:       getEnv("JWKS"),
		KeyID:      getEnv("KEY_ID"),
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}

	return value
}

func GetSpireJwtSource(endpointSocket string) (*workloadapi.JWTSource, error) {
	ctx := context.Background()

	jwtSource, err := workloadapi.NewJWTSource(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(endpointSocket)))
	if err != nil {
		return nil, err
	}

	return jwtSource, nil
}
