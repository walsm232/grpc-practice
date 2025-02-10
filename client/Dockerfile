FROM golang:1.22.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/grpc-client ./client/main.go



FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/grpc-client /app/grpc-client

CMD ["/app/grpc-client"]
