# Tratteria


## Deployment
### Configuration
Tratteria is configured using a YAML file named `config.yaml`, which should be located in the `/app/config/` directory. This configuration file supports customization through environmental variables and JSON path expressions. Below is an example of the application's configuration:

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
  requestMapping:
    subject:
      id: "$.subject_token.email"
    action:
      name: "$.scope"
    resource:
      stockID: "$.request_details.stockID"
      transactionID: "$.request_details.transactionID"
    context: "$.request_context"
```

**Environment Variables:** Use environment variables for sensitive values such as private keys and API tokens. The configuration automatically resolves these variables at runtime during the service startup.

**Authorization API Request:** Specify how to construct the request body for the access evaluation API using JSON path expressions and YAML fields. The configuration allows for the construction of arbitrary JSON using transaction-token request components: subject_token, scope, request_details, and request_context. If a JSON path does not exist for a request, the field is omitted. 

**JWKS Endpoint:** The service signing key's JWKS can be distributed through your infrastructure, or it can be accessed at the standard `GET /.well-known/jwks.json` endpoint.


### Resources
Find the configuration file of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/configs/config.yaml).

Find the Kubernetes deployment configurations of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/).


## Contributing
Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.