# Tratteria
Tratteria is an open source Transaction Tokens (TraTs) Service. The Transaction Tokens draft is defined [here](https://datatracker.ietf.org/doc/draft-ietf-oauth-transaction-tokens/). The directory contains of a fairly elaborate sample application to demonstrate the use of Tratteria, and the Tratteria service itself. The sample application has the following architecture:

~~~
                                    ╔════════════════════════╗                                                              
                                    ║                        ║                                                              
                                    ║                        ║                                                              
                                    ║                        ║                                                              
                                    ║ Tratteria (Transaction ║                                                              
                                    ║    Tokens Service)     ║                                                              
                                    ║                        ║                                                              
                                    ║                        ║                                                              
                                    ║                        ║                                                              
                                    ║                        ║                                                              
                                    ╚════════════════════════╝                                                              
                                                 ▲                                                                          
                      ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ 
                                                 │                                         ┌────────────────────────┐      │
                      │                          │                                         │                        │       
                                                 │                                         │                        │      │
                      │                          │                                         │                        │       
                                                 │                                         │                        │      │
                      │                          │                          ┌─────────────▶│     Stocks Service     │       
                                                 │                          │              │                        │      │
                      │                          │                          │              │                        │       
                                                 │                          │              │                        │      │
┌────────────┐        │                          │                          │              │                        │       
│            │                      ┌────────────────────────┐              │              └────────────────────────┘      │
│            │        │             │                        │              │                           ▲                   
│            │                      │                        │              │                           │                  │
│            │        │             │                        │              │                           │                   
│            │                      │                        │              │                           │                  │
│    User    │────────┼────────────▶│      API Gateway       │──────────────┤                           │                   
│            │                      │                        │              │                           │                  │
│            │        │             │                        │              │                           │                   
│            │                      │                        │              │                           │                  │
│            │        │             │                        │              │                           │                   
│            │                      └────────────────────────┘              │                           │                  │
└────────────┘        │                          │                          │              ┌────────────────────────┐       
       │                                         │                          │              │                        │      │
       │              │                          │                          │              │                        │       
       │                                         │                          │              │                        │      │
       │              │                          │                          │              │                        │       
       │                                         │                          └─────────────▶│     Order Service      │      │
       │              │                          │                                         │                        │       
       │                                         │                                         │                        │      │
       │              │                          │                                         │                        │       
       │                                         │                                         │                        │      │
       │              │                          │                                         └────────────────────────┘       
       │                                         │                                                                         │
       │              │                          │                                                              Sample App  
       │               ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘
       │                                         ▼                                                                          
       │                            ┌────────────────────────┐                                                              
       │                            │                        │                                                              
       │                            │                        │                                                              
       │                            │                        │                                                              
       │                            │   Dex OpenID Connect   │                                                              
       └───────────────────────────▶│   Identity Provider    │                                                              
                                    │                        │                                                              
                                    │                        │                                                              
                                    │                        │                                                              
                                    │                        │                                                              
                                    └────────────────────────┘                                                              ~~~

As shown in the diagram above, the API Gateway in the sample app integrates with the Tratteria service to obtain TraTs that it can use to assure identity and context in its calls downstream, to the Order and Stocks services. The Order Service also calls the Stocks Service and passes the TraT it received from the API Gateway to the Stocks Service. Because TraTs can be passed between downstream services, they can assure identity and call context in arbitrarily deep call chains. The short-lived nature of TraTs makes them relatively immune to replay attacks (unless the replay happens really quickly, and the replay is exactly the same as the information in the TraT).

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
enableAccessEvaluation: true
accessEvaluationAPI:
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