# Gateway Service

The Gateway Service acts as the entry point to the TraTs-Demo-Svcs platform, performing the role of a reverse proxy that routes incoming requests to the appropriate backend services, such as the Orders and Stocks services.

## Features

- **Request Routing**: Directs incoming requests to the correct service based on the request path and method.

- **Security and User Authentication**: Integrates with the Auth service to perform tasks such as user authentication and token generation.

This service is containerized and intended to be run as part of the overall project using Docker Compose. Refer to the main project README for instructions on how to run the entire suite of services.