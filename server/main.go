package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/walsm232/grpc-chat/chatpb"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
}

func (s *ChatServer) ChatStream(stream chatpb.ChatService_ChatStreamServer) error {
	p, ok := peer.FromContext(stream.Context())
	if ok {
		log.Printf("🔗 [SERVER] New client connected from: %s", p.Addr.String())
	} else {
		log.Println("🆕 [SERVER] New client connected, but peer info is unavailable.")
	}

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Println("👋 [SERVER] Client disconnected.")
				return
			}
			if err != nil {
				log.Printf("❌ [SERVER] Error receiving message: %v", err)
				return
			}
			log.Printf("📩 [SERVER] Received from client: %s", msg.Message)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(5)+3) * time.Second)
			randomMsg := fmt.Sprintf("Server says: Random message %d", rand.Intn(100))

			response := &chatpb.ChatMessage{
				Sender:  "Server",
				Message: randomMsg,
			}

			if err := stream.Send(response); err != nil {
				log.Printf("🚨 [SERVER] Error sending message: %v", err)
				return
			}
			log.Printf("📤 [SERVER] Sent via gRPC: %s", randomMsg)
		}
	}()

	select {}
}

func main() {
	port := os.Getenv("GRPC_PORT")
	log.Printf("🚀 [SERVER] Starting gRPC server on port %s...", port)

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("❌ [SERVER] Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, &ChatServer{})

	log.Println("✅ [SERVER] gRPC server started. Waiting for clients...")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ [SERVER] Failed to serve: %v", err)
	}
}
