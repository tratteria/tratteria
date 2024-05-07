package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"gopkg.in/yaml.v2"
)

const CONFIG_FILE_PATH = "/app/config/config.yaml"

type AppConfig struct {
	Issuer                 string               `yaml:"issuer"`
	Audience               string               `yaml:"audience"`
	Token                  Token                `yaml:"token"`
	Keys                   *Keys                `yaml:"keys"`
	Spiffe                 *Spiffe              `yaml:"spiffe,omitempty"`
	SubjectTokens          *SubjectTokens       `yaml:"subjectTokens"`
	EnableAccessEvaluation BoolFromString       `yaml:"enableAccessEvaluation"`
	AccessEvaluationAPI    *AccessEvaluationAPI `yaml:"accessEvaluationAPI,omitempty"`
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

type SubjectTokens struct {
	OIDC       *OIDC       `yaml:"OIDC,omitempty"`
	SelfSigned *SelfSigned `yaml:"selfSigned,omitempty"`
}

type OIDC struct {
	ClientID     string `yaml:"clientId"`
	ProviderURL  string `yaml:"providerURL"`
	SubjectField string `yaml:"subjectField"`
}

type SelfSigned struct {
	JwksEndpoint string         `yaml:"jwksEndpoint,omitempty"`
	Validate     BoolFromString `yaml:"validate"`
}

type Keys struct {
	PrivateKey string `yaml:"privateKey"`
	JWKS       string `yaml:"jwks"`
	KeyID      string `yaml:"keyID"`
}

type AccessEvaluationAPI struct {
	Endpoint       string                             `yaml:"endpoint,omitempty"`
	Authentication *AccessEvaluationAPIAuthentication `yaml:"authentication,omitempty"`
	RequestMapping map[string]interface{}             `yaml:"requestMapping,omitempty"`
}

type AccessEvaluationAPIAuthentication struct {
	Method string                                  `yaml:"method,omitempty"`
	Token  *AccessEvaluationAPIAuthenticationToken `yaml:"token,omitempty"`
}

type AccessEvaluationAPIAuthenticationToken struct {
	Value string `yaml:"value"`
}

func convertMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		strKey, ok := k.(string)
		if !ok {
			panic(fmt.Sprintf("map key is not a string: %v", k))
		}

		if subMap, isMap := v.(map[interface{}]interface{}); isMap {
			result[strKey] = convertMap(subMap)
		} else {
			result[strKey] = v
		}
	}

	return result
}

func (a *AccessEvaluationAPI) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		Endpoint       string                             `yaml:"endpoint"`
		Authentication *AccessEvaluationAPIAuthentication `yaml:"authentication"`
		RequestMapping interface{}                        `yaml:"requestMapping"`
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	a.Endpoint = raw.Endpoint
	a.Authentication = raw.Authentication

	switch v := raw.RequestMapping.(type) {
	case map[interface{}]interface{}:
		a.RequestMapping = convertMap(v)
	case nil:
		a.RequestMapping = nil
	default:
		return fmt.Errorf("unsupported type for requestMapping: %T", v)
	}

	return nil
}

type BoolFromString bool

func (b *BoolFromString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp interface{}
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	switch value := tmp.(type) {
	case bool:
		*b = BoolFromString(value)
	case string:
		if matched, envVarName := extractEnvVarName(value); matched {
			envValue, err := getEnvVarValue(envVarName)
			if err != nil {
				return err
			}

			boolVal, err := strconv.ParseBool(envValue)

			if err != nil {
				return fmt.Errorf("error parsing boolean from environment variable: %v", err)
			}

			*b = BoolFromString(boolVal)
		} else {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("error parsing boolean from string: %v", err)
			}

			*b = BoolFromString(boolVal)
		}
	default:
		return fmt.Errorf("invalid type for a bool variable, expected bool or string, got %T", tmp)
	}

	return nil
}

func GetAppConfig() *AppConfig {
	data, err := os.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal YAML configuration: %v", err))
	}

	resolveEnvVariables(&cfg)
	validateConfig(&cfg)

	return &cfg
}

func validateConfig(cfg *AppConfig) {
	if cfg.Issuer == "" {
		panic("issuer must not be empty")
	}

	if cfg.Audience == "" {
		panic("audience must not be empty")
	}

	if cfg.Keys.PrivateKey == "" {
		panic("private key must be provided")
	}

	if cfg.Keys.JWKS == "" {
		panic("key JWKS must be provided")
	}

	if cfg.Keys.KeyID == "" {
		panic("keyID must be provided")
	}

	validateSpiffe(cfg.Spiffe)

	validateSubjectTokens(cfg.SubjectTokens)

	if cfg.EnableAccessEvaluation {
		validateAccessEvaluationAPI(cfg.AccessEvaluationAPI)
	}
}

func validateSpiffe(spiffe *Spiffe) {
	if spiffe == nil {
		return
	}

	if len(spiffe.AuthorizedServiceIDs) == 0 {
		panic("authorized services Spiffe IDs must be specified")
	}

	if spiffe.EndpointSocket == "" {
		panic("endpoint socket must be provided for Spiffe")
	}
}

func validateSubjectTokens(subjectToken *SubjectTokens) {
	if subjectToken.OIDC == nil && subjectToken.SelfSigned == nil {
		panic("At least one subject token configuration must be provided")
	}

	if subjectToken.OIDC != nil {
		validateOIDC(subjectToken.OIDC)
	}

	if subjectToken.SelfSigned != nil {
		validateSelfSigned(subjectToken.SelfSigned)
	}
}

func validateOIDC(oidc *OIDC) {
	if oidc == nil {
		panic("OIDC configuration must be provided")
	}

	if oidc.ClientID == "" {
		panic("OIDC client ID must be populated")
	}

	if oidc.ProviderURL == "" {
		panic("OIDC provider URL must be populated")
	}

	if oidc.SubjectField == "" {
		panic("OIDC subject field must be populated")
	}
}

func validateSelfSigned(selfSigned *SelfSigned) {
	if selfSigned == nil {
		panic("SelfSigned configuration must be provided")
	}

	if selfSigned.Validate {
		if selfSigned.JwksEndpoint == "" {
			panic("JWKS Endpoint must be provided if Validate is true")
		}
	}
}

func validateAccessEvaluationAPI(api *AccessEvaluationAPI) {
	if api == nil {
		panic("AuthorizationAPI configuration must be provided")
	}

	if api.Endpoint == "" {
		panic("AuthorizationAPI endpoint must not be empty")
	}

	if api.Authentication == nil {
		panic("AuthorizationAPI authentication configuration must be provided")
	}

	if api.Authentication.Method == "" {
		panic("AuthorizationAPI authentication method must not be empty")
	}

	if api.Authentication.Token == nil {
		panic("AuthorizationAPI authentication token must be provided")
	}

	if api.Authentication.Token.Value == "" {
		panic("AuthorizationAPI authentication token value must not be empty")
	}

	if len(api.RequestMapping) == 0 {
		panic("AuthorizationAPI request mapping must be provided and cannot be empty")
	}
}

func extractEnvVarName(s string) (bool, string) {
	envVarRegex := regexp.MustCompile(`^\$\{([^}]+)\}$`)
	matches := envVarRegex.FindStringSubmatch(s)

	if len(matches) > 1 {
		return true, matches[1]
	}

	return false, ""
}

func getEnvVarValue(envVarName string) (string, error) {
	if envValue, exists := os.LookupEnv(envVarName); exists {
		return envValue, nil
	}

	return "", fmt.Errorf("environment variable %s not set", envVarName)
}

func resolveEnvVariablesUtil(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			fieldValue := field.String()

			if matched, envVarName := extractEnvVarName(fieldValue); matched {
				envValue, err := getEnvVarValue(envVarName)
				if err != nil {
					panic(err.Error())
				}

				field.SetString(envValue)
			}
		} else if field.Kind() == reflect.Struct {
			resolveEnvVariablesUtil(field)
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			resolveEnvVariablesUtil(field.Elem())
		}
	}
}

func resolveEnvVariables(cfg *AppConfig) {
	v := reflect.ValueOf(cfg)
	resolveEnvVariablesUtil(v)
}
