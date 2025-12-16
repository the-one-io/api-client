package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Example of WebSocket client usage
func runWebSocketExample() {
	// API keys (obtained from server)
	apiKey := "ak_X0sw1Dr97hcXkFsDM9nXbD2gn2ZkCIptjtpqQ-MvAnc"
	secretKey := "L2pcViLs7uGdFZJd3wYKmnSgoSv-UYx8oF4c6lX95NSk3Ejm-T5eWproVlRcvQn1"
	wsURL := "ws://localhost:8080/ws/v1/stream"

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
	err = wsClient.EstimateSwap("100.00", "USDT", "ETH")
	if err != nil {
		log.Printf("Failed to estimate swap: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Test swap (creates new order)
	fmt.Println("🔄 Testing swap operation...")
	err = wsClient.DoSwap("10.00", "USDT", "ETH")
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
