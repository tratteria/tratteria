# First stage: the build environment
FROM golang:1.22.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tratteria ./cmd

# Second stage: the runtime environment
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/tratteria .

ENTRYPOINT ["./tratteria"]