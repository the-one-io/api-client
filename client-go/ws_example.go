package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// Example of WebSocket client usage
func runWebSocketExample() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get API keys from environment variables
	apiKey := os.Getenv("BROKER_API_KEY")
	secretKey := os.Getenv("BROKER_SECRET_KEY")
	baseURL := os.Getenv("BROKER_BASE_URL")

	// Validate required environment variables
	if apiKey == "" || secretKey == "" || baseURL == "" {
		log.Fatal("Error: BROKER_API_KEY, BROKER_SECRET_KEY, and BROKER_BASE_URL must be set in .env file or environment")
	}

	// Convert HTTP URL to WebSocket URL
	wsURL := baseURL
	if len(wsURL) > 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:] + "/ws/v1/stream"
	} else if len(wsURL) > 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:] + "/ws/v1/stream"
	} else {
		wsURL = wsURL + "/ws/v1/stream"
	}

	// Create WebSocket client
	wsClient := NewWSClient(apiKey, secretKey, wsURL)

	// Connect to WebSocket
	fmt.Println("=== Connecting to WebSocket ===")
	if err := wsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Wait a bit for authentication to complete
	time.Sleep(2 * time.Second)

	// Subscribe to balances channel
	fmt.Println("\n=== Subscribing to balances channel ===")
	err := wsClient.Subscribe("balances", func(msg *WSMessage) {
		fmt.Printf("📊 Received balances update on channel '%s'\n", msg.Channel)
		if msg.Data != nil {
			dataBytes, _ := json.MarshalIndent(msg.Data, "", "  ")
			fmt.Printf("Data: %s\n", dataBytes)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to balances: %v", err)
	}

	// Subscribe to specific order channel
	fmt.Println("\n=== Subscribing to order updates ===")
	orderID := "ord_12345678"
	orderChannel := "orders:" + orderID
	err = wsClient.Subscribe(orderChannel, func(msg *WSMessage) {
		fmt.Printf("📦 Received order update for order '%s'\n", orderID)
		if msg.Data != nil {
			dataBytes, _ := json.MarshalIndent(msg.Data, "", "  ")
			fmt.Printf("Data: %s\n", dataBytes)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to order updates: %v", err)
	}

	// Demo REST API commands via WebSocket
	fmt.Println("\n=== Testing REST API commands via WebSocket ===")

	// Test balances request
	fmt.Println("💼 Getting account balances...")
	err = wsClient.GetBalances()
	if err != nil {
		log.Printf("Failed to get balances: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Test estimate
	fmt.Println("💰 Testing estimate swap...")
	err = wsClient.EstimateSwapSimple("10.00", "USDT", "ETH", []string{"binance", "gate"})
	if err != nil {
		log.Printf("Failed to estimate swap: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Test swap (creates new order)
	fmt.Println("🔄 Testing swap operation...")
	err = wsClient.DoSwapSimple("10.00", "USDT", "ETH", 30, []string{"binance", "gate"})
	if err != nil {
		log.Printf("Failed to do swap: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Test order status
	fmt.Println("📋 Getting order status...")
	err = wsClient.GetOrderStatus("32588")
	if err != nil {
		log.Printf("Failed to get order status: %v", err)
	}

	// Set up graceful shutdown
	fmt.Printf("\n=== WebSocket client is running ===\n")
	fmt.Println("Listening for real-time updates...")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run until interrupted
	select {
	case <-sigChan:
		fmt.Println("\n=== Shutting down ===")
	case <-wsClient.ctx.Done():
		fmt.Println("\n=== WebSocket connection lost ===")
	}

	// Unsubscribe before closing (optional)
	fmt.Println("Unsubscribing from channels...")
	wsClient.Unsubscribe("balances")
	wsClient.Unsubscribe(orderChannel)

	// Close connection
	wsClient.Close()
	fmt.Println("WebSocket client stopped")
}
