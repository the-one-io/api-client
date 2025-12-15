package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Op        string      `json:"op,omitempty"`
	Channel   string      `json:"ch,omitempty"`
	Key       string      `json:"key,omitempty"`
	Timestamp int64       `json:"ts,omitempty"`
	Nonce     string      `json:"nonce,omitempty"`
	Signature string      `json:"sig,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// WSClient represents WebSocket client for Broker Trading API
type WSClient struct {
	apiKey    string
	secretKey string
	wsURL     string
	conn      *websocket.Conn
	mu        sync.RWMutex
	handlers  map[string][]MessageHandler
	connected bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// MessageHandler is a function to handle incoming messages
type MessageHandler func(msg *WSMessage)

// NewWSClient creates a new WebSocket client
func NewWSClient(apiKey, secretKey, wsURL string) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WSClient{
		apiKey:    apiKey,
		secretKey: secretKey,
		wsURL:     wsURL,
		handlers:  make(map[string][]MessageHandler),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Connect establishes WebSocket connection and authenticates
func (ws *WSClient) Connect() error {
	u, err := url.Parse(ws.wsURL)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	log.Printf("Connecting to WebSocket: %s", ws.wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.connected = true
	ws.mu.Unlock()

	log.Println("WebSocket connected successfully")

	// Start message reading in background
	go ws.readMessages()

	// Authenticate
	if err := ws.authenticate(); err != nil {
		ws.Close()
		return fmt.Errorf("authentication failed: %w", err)
	}

	log.Println("Authentication successful")
	return nil
}

// authenticate sends authentication message
func (ws *WSClient) authenticate() error {
	timestamp := time.Now().UnixMilli()
	nonce := ws.generateNonce()

	// For WebSocket auth, use the same signature format as REST API:
	// Method: "WS", Path: "/ws/v1/stream", Body: empty (SHA256 of empty bytes)
	method := "WS"
	pathWithQuery := "/ws/v1/stream"
	bodySHA256 := ws.hashBody([]byte{}) // Empty body for WS auth

	canonicalString := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
		method, pathWithQuery, timestamp, nonce, bodySHA256)

	hash := sha256.Sum256([]byte(ws.secretKey))
	secretKeyBase64 := base64.URLEncoding.EncodeToString(hash[:])
	hmacKey := []byte(secretKeyBase64)

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write([]byte(canonicalString))
	signature := hex.EncodeToString(mac.Sum(nil))

	authMsg := WSMessage{
		Op:        "auth",
		Key:       ws.apiKey,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
	}

	return ws.sendMessage(&authMsg)
}

// hashBody creates SHA256 hash of request body
func (ws *WSClient) hashBody(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// Subscribe subscribes to a channel
func (ws *WSClient) Subscribe(channel string, handler MessageHandler) error {
	ws.mu.Lock()
	ws.handlers[channel] = append(ws.handlers[channel], handler)
	ws.mu.Unlock()

	subscribeMsg := WSMessage{
		Op:      "subscribe",
		Channel: channel,
	}

	return ws.sendMessage(&subscribeMsg)
}

// Unsubscribe unsubscribes from a channel
func (ws *WSClient) Unsubscribe(channel string) error {
	ws.mu.Lock()
	delete(ws.handlers, channel)
	ws.mu.Unlock()

	unsubscribeMsg := WSMessage{
		Op:      "unsubscribe",
		Channel: channel,
	}

	return ws.sendMessage(&unsubscribeMsg)
}

// sendMessage sends a message to WebSocket
func (ws *WSClient) sendMessage(msg *WSMessage) error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if !ws.connected || ws.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	return ws.conn.WriteJSON(msg)
}

// readMessages reads messages from WebSocket in a loop
func (ws *WSClient) readMessages() {
	defer ws.Close()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
		}

		var msg WSMessage
		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			return
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		ws.handleMessage(&msg)
	}
}

// handleMessage handles incoming WebSocket messages
func (ws *WSClient) handleMessage(msg *WSMessage) {
	// Handle error messages
	if msg.Error != "" {
		log.Printf("WebSocket error: %s", msg.Error)
		return
	}

	// Handle response messages by operation
	if msg.Op != "" {
		log.Printf("Received response for operation: %s", msg.Op)
		if msg.Data != nil {
			log.Printf("Response data: %+v", msg.Data)
		}
		return
	}

	// Handle data messages by channel
	if msg.Channel != "" {
		ws.mu.RLock()
		handlers := ws.handlers[msg.Channel]
		ws.mu.RUnlock()

		for _, handler := range handlers {
			go handler(msg)
		}
	}
}

// generateNonce creates unique nonce
func (ws *WSClient) generateNonce() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
}

// Close closes WebSocket connection
func (ws *WSClient) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.cancel != nil {
		ws.cancel()
	}

	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}

	ws.connected = false
	log.Println("WebSocket connection closed")
}

// IsConnected returns connection status
func (ws *WSClient) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}

// Wait waits for context to be done
func (ws *WSClient) Wait() {
	<-ws.ctx.Done()
}

// EstimateSwap estimates a swap operation via WebSocket
func (ws *WSClient) EstimateSwap(amountIn, assetIn, assetOut string) error {
	estimateData := map[string]interface{}{
		"amountIn": amountIn,
		"assetIn":  assetIn,
		"assetOut": assetOut,
	}

	estimateMsg := ws.createSignedMessage("estimate", estimateData)
	return ws.sendMessage(estimateMsg)
}

// DoSwap executes a swap operation via WebSocket
func (ws *WSClient) DoSwap(amountIn, assetIn, assetOut string) error {
	swapData := map[string]interface{}{
		"amountIn": amountIn,
		"assetIn":  assetIn,
		"assetOut": assetOut,
	}

	swapMsg := WSMessage{
		Op:   "swap",
		Data: swapData,
	}

	return ws.sendMessage(&swapMsg)
}

// GetOrderStatus gets order status via WebSocket
func (ws *WSClient) GetOrderStatus(orderID string) error {
	orderData := map[string]interface{}{
		"id": orderID,
	}

	orderMsg := WSMessage{
		Op:   "order_status",
		Data: orderData,
	}

	return ws.sendMessage(&orderMsg)
}

// GetBalances gets account balances via WebSocket
func (ws *WSClient) GetBalances() error {
	balancesMsg := ws.createSignedMessage("balances", nil)
	return ws.sendMessage(balancesMsg)
}

// createSignedMessage creates a signed message for operations that require authentication
func (ws *WSClient) createSignedMessage(operation string, data interface{}) *WSMessage {
	timestamp := time.Now().UnixMilli()
	nonce := ws.generateNonce()

	// For WebSocket operations, use the operation as the "method" and a path
	method := "WS"
	pathWithQuery := "/ws/v1/" + operation

	// Convert data to JSON for body hash calculation
	var bodyBytes []byte
	if data != nil {
		var err error
		bodyBytes, err = json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling data: %v", err)
			return &WSMessage{
				Op:    operation,
				Data:  data,
				Error: "Failed to marshal data",
			}
		}
	}
	bodySHA256 := ws.hashBody(bodyBytes)

	canonicalString := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
		method, pathWithQuery, timestamp, nonce, bodySHA256)

	hash := sha256.Sum256([]byte(ws.secretKey))
	secretKeyBase64 := base64.URLEncoding.EncodeToString(hash[:])
	hmacKey := []byte(secretKeyBase64)

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write([]byte(canonicalString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return &WSMessage{
		Op:        operation,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
		Data:      data,
	}
}
