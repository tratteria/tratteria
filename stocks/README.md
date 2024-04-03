# Stocks Service

The Stocks Service provides a set of APIs related to stocks and user's portfolio. It offers endpoints for searching stocks, retrieving stock details, and tracking user's portfolio.

## Endpoints

- `GET api/stocks/search?query={search-query}`: Searches for stocks based on the search query.
- `GET api/stocks/{stock-id}`: Retrieves details of a specific stock by its ID.
- `GET api/stocks/holding`: Lists all holdings in the user's portfolio.
- `POST internal/stocks`: Updates the stocks and user's holding data in the database for a trading operation.

## Running the Service

This service is containerized and intended to be run as part of the overall project using Docker Compose. Refer to the main project README for instructions on how to run the entire suite of services.
