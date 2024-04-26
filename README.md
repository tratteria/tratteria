# Tratteria


## Deployment
### Configuration
Tratteria is configured using a YAML file and supports customization through environmental variables and JSON path expressions. Below is an example of the application's configuration:

```yaml
issuer: https://example.org/tts
audience: https://example.org/
token:
  lifeTime: "15s"
keys:
  privateKey: ${PRIVATE_KEY}
  jwks: ${JWKS}
  keyID: ${KEY_ID}
spiffe:
  endpoint_socket: unix:///run/spire/sockets/agent.sock
  serviceID: spiffe://example.org/tts
  authorizedServiceIDs:
    - spiffe://example.org/gateway
clientAuthenticationMethods:
  OIDC:
    clientId: example-client
    providerURL: http://example.org/oidcprovider
    subjectField: email
authorizationAPI:
  endpoint: https://example.authzen.com/access/v1/evaluation
  authentication:
    method: Bearer
    token:
      value: ${AUTHORIZATION_API_BEARER_TOKEN}
#requestMapping: Define how to construct the request body for the Authorization API using JSON path expressions. This mapping pulls specific values from
#the transaction-token request (subject_token, scope, request_details, and request_context) using JSON path expression and assigns them to respective
#fields below to construct the JSON request body for the Authorization API.
  requestMapping:
    subject:
      id: "$.subject_token.email"
      name: "$.subject_token.name"
    action:
      type: "$.scope"
      details:
        action: "$.request_details.action"
        quantity: "$.request_details.quantity"
    resource:
      stock: "$.request_details.stockID"
      transaction: "$.request_details.transactionID"
    context: "$.request_context"
```

### Key Features
**Environment Variables:** Use environment variables for sensitive values such as private keys and API tokens. The configuration automatically resolves these variables at runtime during the service startup.

**Authorization API Request Construction:** Specify how to construct the request body for the Authorization API. This configuration allows for constructing arbitrary JSON request bodies using JSON path expressions and YAML fields. The presence of a specific key in the JSON path determines whether it will be included in the request; if a key does not exist for a particular request, it will be omitted.

**JWKS Endpoint:** The service signing key's JWKS can be distributed through your infrastructure, or it can be accessed at the standard `GET /.well-known/jwks.json` endpoint.


### Resources
Find the configuration file of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/configs/config.yaml).

Find the Kubernetes deployment configurations of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/configs/config.yaml).


## Contributing
Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.