# First stage: build environment
FROM --platform=$BUILDPLATFORM golang:1.22.0 AS builder
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o tratteria ./cmd

# Second stage: runtime environment
FROM --platform=$TARGETPLATFORM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/tratteria .

LABEL org.opencontainers.image.source=https://github.com/tratteria/tratteria
LABEL org.opencontainers.image.description="Tratteria - Transaction Tokens Service"
LABEL org.opencontainers.image.licenses=Apache

ENTRYPOINT ["./tratteria"]