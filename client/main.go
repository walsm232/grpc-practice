package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/walsm232/grpc-chat/chatpb"
)

const (
	maxRetries        = 5
	baseDelay         = 2
	maxBackoffTime    = 30
	connectionTimeout = 30
)

func backoffWithJitter(attempt int) time.Duration {
	expBackoff := float64(baseDelay) * math.Pow(2, float64(attempt))
	jitter := rand.Float64() * expBackoff
	backoffTime := time.Duration(math.Min(expBackoff+jitter, float64(maxBackoffTime))) * time.Second
	return backoffTime
}

func waitForServer(clientConn *grpc.ClientConn) error {
	timeout := time.After(time.Duration(connectionTimeout) * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	log.Println("[CLIENT] 🔍 Waiting for gRPC server to become available...")

	for {
		select {
		case <-timeout:
			return fmt.Errorf("❌ gRPC server did not become available within %d seconds", connectionTimeout)
		case <-ticker.C:
			clientConn.Connect()

			state := clientConn.GetState()
			log.Printf("[CLIENT] 📡 Connection state: %s", state.String())

			if state == connectivity.Ready {
				log.Println("[CLIENT] ✅ Connection is READY!")
				log.Println("[CLIENT] ✅ Connected to the gRPC server.")
				ticker.Stop()
				return nil
			}
		}
	}
}

func main() {
	serverAddress := os.Getenv("GRPC_SERVER_ADDRESS")

	var clientConn *grpc.ClientConn
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[CLIENT] Attempt %d/%d: Connecting to gRPC server at %s...", attempt, maxRetries, serverAddress)

		clientConn, err = grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {

			if err := waitForServer(clientConn); err == nil {
				break
			}
		}

		log.Printf("[CLIENT] ❌ Connection failed (attempt %d/%d): %v", attempt, maxRetries, err)

		backoffTime := backoffWithJitter(attempt)
		log.Printf("[CLIENT] ⏳ Retrying in %s...", backoffTime)
		time.Sleep(backoffTime)
	}

	if err != nil {
		log.Fatalf("[CLIENT] ❌ Unable to connect after %d attempts. Exiting.", maxRetries)
	}
	defer clientConn.Close()

	client := chatpb.NewChatServiceClient(clientConn)
	stream, err := client.ChatStream(context.Background())
	if err != nil {
		log.Fatalf("[CLIENT] ❌ Error opening gRPC stream: %v", err)
	}

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("[CLIENT] Server closed the connection.")
				return
			}
			if err != nil {
				log.Printf("[CLIENT] ❌ Error receiving message: %v", err)
				return
			}
			log.Printf("[CLIENT] 📩 Received: %s", resp.Message)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(5)+3) * time.Second)
			randomMsg := fmt.Sprintf("Client says: Random message %d", rand.Intn(100))

			msg := &chatpb.ChatMessage{
				Sender:  "Client",
				Message: randomMsg,
			}

			if err := stream.Send(msg); err != nil {
				log.Printf("[CLIENT] ❌ Error sending message: %v", err)
				return
			}
			log.Printf("[CLIENT] ✉️ Sent via gRPC: %s", randomMsg)
		}
	}()

	select {}
}
