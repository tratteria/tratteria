# TODO: Use a multi-stage build process to reduce the final image size.
FROM golang:1.22.0 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd

CMD ["./service"]