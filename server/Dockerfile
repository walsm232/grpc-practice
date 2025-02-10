FROM golang:1.22.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/grpc-server ./server/main.go



FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/grpc-server /app/grpc-server

EXPOSE 50051

CMD ["/app/grpc-server"]
