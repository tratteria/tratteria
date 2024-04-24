# Tratteria


## Deployment
### Configuration
Tratteria is configured using a YAML file. Below is an example configuration:

```yaml
issuer: https://example.org/tts
audience: https://example.org/
token:
  lifeTime: "15s"
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
```

Find the configuration file of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/configs/config.yaml).




### Environment Variables
Tratteria requires the following environment variables to be set:

`PRIVATE_KEY:` The value of the private key used for signing TraTs.

`JWKS:` The JSON Web Key Set of the key.

`KEY_ID:` The identifier for the key.

The JWKS can be either distributed through your infrastructure or accessed at  `GET /.well-known/jwks.json`


<br><br>
Find the Kubernetes deployment configurations of the example application [here](https://github.com/SGNL-ai/Tratteria/tree/main/example-application/deployments/kubernetes/tratteria/).

Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.