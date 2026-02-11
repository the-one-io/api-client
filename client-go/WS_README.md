# WebSocket Client for Go

WebSocket client for TheOne Trading API in Go with authentication and real-time support.

> 📖 **Related Documentation:**
> - [Go REST API Client](./README.md) - REST API documentation
> - [Main README](../README.md) - Overview of all clients
> - [JavaScript WebSocket](../client-js/WS_README.md) - JavaScript WebSocket documentation
> - [Python WebSocket](../client-py/WS_README.md) - Python WebSocket documentation

## Features

- ✅ Full WebSocket API implementation
- ✅ HMAC-SHA256 authentication
- ✅ Auto-reconnection with exponential backoff
- ✅ Channel subscriptions (balances, orders)
- ✅ Real-time updates
- ✅ Thread-safe operations
- ✅ Graceful shutdown
- ✅ Environment variables support (.env file)

## Installation

```bash
cd client-go
go mod download
go mod tidy
```

## Dependencies

- `github.com/gorilla/websocket` - WebSocket library for Go
- `github.com/joho/godotenv` - Environment variables loader

## Quick Start

### Using Makefile (Recommended)

```bash
# Run WebSocket client example
make run-ws

# See all available commands
make help
```

### Manual Run

```bash
go run ws_main.go ws_client.go ws_example.go
```

## Configuration

Create a `.env` file in the `client-go` directory:

```bash
BROKER_API_KEY=your_api_key_here
BROKER_SECRET_KEY=your_secret_key_here
BROKER_BASE_URL=https://partner-api-dev.the-one.io
```

The WebSocket client will automatically:
1. Load environment variables from `.env` file
2. Convert HTTP URL to WebSocket URL (https:// → wss://)
3. Append `/ws/v1/stream` to the URL

## Usage

### Simple Connection

```go
package main

import (
    "log"
    "os"
    "time"
    
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }
    
    // Get API keys
    apiKey := os.Getenv("BROKER_API_KEY")
    secretKey := os.Getenv("BROKER_SECRET_KEY")
    baseURL := os.Getenv("BROKER_BASE_URL")
    
    // Convert to WebSocket URL
    wsURL := strings.Replace(baseURL, "https://", "wss://", 1) + "/ws/v1/stream"
    
    // Create client
    client := NewWSClient(apiKey, secretKey, wsURL)
    
    // Connect
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Subscribe to balances
    err := client.Subscribe("balances", func(msg *WSMessage) {
        log.Printf("Balances update: %+v", msg.Data)
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for activity
    time.Sleep(60 * time.Second)
}
```

### Channel Subscriptions

#### User Balances

```go
err := client.Subscribe("balances", func(msg *WSMessage) {
    fmt.Printf("📊 Balance Update:\n")
    if msg.Data != nil {
        dataBytes, _ := json.MarshalIndent(msg.Data, "", "  ")
        fmt.Printf("%s\n", dataBytes)
    }
})
if err != nil {
    log.Printf("Failed to subscribe: %v", err)
}
```

#### Order Status

```go
orderID := "ord_12345678"
channel := "orders:" + orderID

err := client.Subscribe(channel, func(msg *WSMessage) {
    fmt.Printf("📦 Order Update for %s:\n", orderID)
    if msg.Data != nil {
        dataBytes, _ := json.MarshalIndent(msg.Data, "", "  ")
        fmt.Printf("%s\n", dataBytes)
    }
})
if err != nil {
    log.Printf("Failed to subscribe: %v", err)
}
```

### Error Handling

```go
client.Subscribe("balances", func(msg *WSMessage) {
    if msg.Error != "" {
        log.Printf("❌ Error: %s", msg.Error)
        return
    }
    
    // Process data
    log.Printf("✅ Data: %+v", msg.Data)
})
```

### Graceful Shutdown

```go
import (
    "os"
    "os/signal"
    "syscall"
)

// Set up signal handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Wait for interrupt
select {
case <-sigChan:
    fmt.Println("\n=== Shutting down ===")
case <-wsClient.ctx.Done():
    fmt.Println("\n=== Connection lost ===")
}

// Unsubscribe and close
wsClient.Unsubscribe("balances")
wsClient.Unsubscribe(orderChannel)
wsClient.Close()

fmt.Println("WebSocket client stopped")
```

## API Reference

### WSClient

```go
type WSClient struct {
    // Private fields
}
```

#### Connection Methods

##### NewWSClient

```go
func NewWSClient(apiKey, secretKey, wsURL string) *WSClient
```

Creates a new WebSocket client.

**Parameters:**
- `apiKey` - API key
- `secretKey` - Secret key
- `wsURL` - WebSocket server URL

**Returns:** `*WSClient`

##### Connect

```go
func (ws *WSClient) Connect() error
```

Establishes connection and authenticates.

**Returns:** `error` - nil on success

##### Subscribe

```go
func (ws *WSClient) Subscribe(channel string, handler MessageHandler) error
```

Subscribes to a channel.

**Parameters:**
- `channel` - Channel name
- `handler` - Message handler function

**Returns:** `error` - nil on success

**Handler signature:**
```go
type MessageHandler func(*WSMessage)
```

##### Unsubscribe

```go
func (ws *WSClient) Unsubscribe(channel string) error
```

Unsubscribes from a channel.

**Parameters:**
- `channel` - Channel name

**Returns:** `error` - nil on success

##### Close

```go
func (ws *WSClient) Close()
```

Closes WebSocket connection and cleans up resources.

##### IsConnected

```go
func (ws *WSClient) IsConnected() bool
```

Checks connection status.

**Returns:** `bool` - true if connected

#### REST API Commands via WebSocket

##### GetBalances

```go
func (ws *WSClient) GetBalances() error
```

Requests current account balances (signed message).

**Returns:** `error` - nil on success

##### EstimateSwapSimple

```go
func (ws *WSClient) EstimateSwapSimple(amountIn, assetIn, assetOut string, filter []string) error
```

Estimates swap cost (signed message).

**Parameters:**
- `amountIn` - Amount of input asset
- `assetIn` - Input asset symbol
- `assetOut` - Output asset symbol
- `filter` - List of exchanges to use (optional)

**Returns:** `error` - nil on success

##### DoSwapSimple

```go
func (ws *WSClient) DoSwapSimple(amountIn, assetIn, assetOut string, slippageBps int, filter []string) error
```

Executes swap operation (signed message).

**Parameters:**
- `amountIn` - Amount of input asset
- `assetIn` - Input asset symbol
- `assetOut` - Output asset symbol
- `slippageBps` - Slippage tolerance in basis points (e.g., 30 = 0.3%)
- `filter` - List of exchanges to use (optional)

**Returns:** `error` - nil on success

##### GetOrderStatus

```go
func (ws *WSClient) GetOrderStatus(orderID string) error
```

Gets order status (signed message).

**Parameters:**
- `orderID` - Order ID to query

**Returns:** `error` - nil on success

### WSMessage

```go
type WSMessage struct {
    Op        string      `json:"op,omitempty"`        // Operation type
    Channel   string      `json:"ch,omitempty"`        // Channel name
    Key       string      `json:"key,omitempty"`       // API key (auth only)
    Timestamp int64       `json:"ts,omitempty"`        // Timestamp in milliseconds
    Nonce     string      `json:"nonce,omitempty"`     // Unique nonce
    Signature string      `json:"sig,omitempty"`       // HMAC signature
    Data      interface{} `json:"data,omitempty"`      // Message data
    Error     string      `json:"error,omitempty"`     // Error description
}
```

## Available Channels

### balances

Receives user balance updates.

**Channel:** `balances`

**Example data:**
```go
{
  "ch": "balances",
  "data": [
    {
      "asset": "USDT",
      "total": "1000.00",
      "locked": "0"
    },
    {
      "asset": "BTC",
      "total": "0.5",
      "locked": "0.1"
    }
  ]
}
```

### orders:{orderId}

Receives status updates for a specific order.

**Channel:** `orders:{orderId}` (e.g., `orders:ord_12345678`)

**Example data:**
```go
{
  "ch": "orders:ord_12345678",
  "data": {
    "orderId": "ord_12345678",
    "status": "FILLED",
    "txHash": "0xabc123...",
    "filledOut": "100.00",
    "updatedAt": 1640995200000
  }
}
```

## Running the Example

```bash
# Using Makefile (recommended)
make run-ws

# Using go run
go run ws_main.go ws_client.go ws_example.go

# Build and run
go build -o ws_client_example ws_main.go ws_client.go ws_example.go
./ws_client_example
```

## Features Implementation

### Auto-Reconnection

The client automatically reconnects on connection loss with exponential backoff:

- **Initial delay:** 1 second
- **Maximum delay:** 32 seconds
- **Exponential backoff:** delay doubles after each failed attempt
- **Maximum attempts:** Unlimited (continues retrying)

All subscriptions are automatically restored after reconnection.

### Thread Safety

The WebSocket client is thread-safe:

- Safe concurrent subscription/unsubscription
- Safe concurrent message sending
- Protected internal state with mutexes

### Context Support

The client uses Go's context for cancellation and timeout:

```go
// Client has internal context
wsClient.ctx     // Context for lifecycle management
wsClient.cancel  // Cancel function
```

### Message Queue

Messages are queued internally:

- Non-blocking send operations
- Automatic retry on send failure
- Graceful handling of connection issues

## Security

### Authentication

All messages are authenticated using HMAC-SHA256:

1. **Auth message** on connection:
   ```go
   {
     "op": "auth",
     "key": "your_api_key",
     "ts": 1640995200000,
     "nonce": "unique_nonce",
     "sig": "hmac_signature"
   }
   ```

2. **Signed operations** (swap, estimate, order status, balances):
   - Timestamp in milliseconds
   - Unique nonce (format: `{nanoseconds}_{random}`)
   - HMAC-SHA256 signature

### Best Practices

1. **Never commit API keys** to version control
   - Use `.env` files (add to `.gitignore`)
   - Use environment variables in production

2. **Use secure connections**
   - Always use `wss://` URLs, not `ws://`

3. **Handle disconnections**
   - Auto-reconnection is built-in
   - Implement proper error handling
   - Monitor connection status

4. **Resource cleanup**
   - Always call `Close()` when done
   - Use `defer client.Close()`
   - Handle graceful shutdown signals

## Integration Examples

### With HTTP Server

```go
package main

import (
    "net/http"
    "sync"
)

var (
    wsClient *WSClient
    mu       sync.RWMutex
)

func main() {
    // Initialize WebSocket client
    wsClient = NewWSClient(apiKey, secretKey, wsURL)
    if err := wsClient.Connect(); err != nil {
        log.Fatal(err)
    }
    defer wsClient.Close()
    
    // HTTP endpoints
    http.HandleFunc("/status", statusHandler)
    http.HandleFunc("/swap", swapHandler)
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    mu.RLock()
    connected := wsClient.IsConnected()
    mu.RUnlock()
    
    json.NewEncoder(w).Encode(map[string]bool{
        "websocket_connected": connected,
    })
}

func swapHandler(w http.ResponseWriter, r *http.Request) {
    // Execute swap via WebSocket
    err := wsClient.DoSwapSimple("100", "ETH", "USDT", 30, nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "swap initiated",
    })
}
```

### With Goroutines

```go
// Multiple concurrent subscriptions
go func() {
    wsClient.Subscribe("balances", handleBalances)
}()

go func() {
    for orderID := range orderIDsChan {
        channel := "orders:" + orderID
        wsClient.Subscribe(channel, handleOrder)
    }
}()

// Monitoring goroutine
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if !wsClient.IsConnected() {
                log.Println("⚠️  WebSocket disconnected")
            }
        case <-ctx.Done():
            return
        }
    }
}()
```

## Troubleshooting

### Common Issues

1. **Connection refused**
   - Check WebSocket URL format (wss:// or ws://)
   - Verify network connectivity
   - Check firewall settings

2. **Authentication failed**
   - Verify API keys in `.env` file
   - Check signature generation
   - Ensure system clock is synchronized

3. **Subscriptions not working**
   - Ensure connection is established before subscribing
   - Check channel name format
   - Verify handler function is not nil

4. **Auto-reconnection not working**
   - Check if `Close()` was called
   - Verify context is not cancelled
   - Check logs for reconnection attempts

### Debug Mode

Enable debug logging:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)

// In handler
func handleMessage(msg *WSMessage) {
    log.Printf("DEBUG: Received message: %+v\n", msg)
}
```

## Performance Considerations

- **Concurrent handlers:** Handler functions run concurrently
- **Message queueing:** Internal queue prevents blocking
- **Memory management:** Old subscriptions are cleaned up
- **Connection pooling:** Single connection per client

## Limitations

- Maximum 100 active subscriptions per connection
- Authentication timeout: 10 seconds
- Maximum message size: 64KB
- Heartbeat interval: 30 seconds
- Reconnection delay range: 1-32 seconds

## Related Documentation

- **[REST API Client](./README.md)** - Complete REST API documentation with .env support
- **[Main README](../README.md)** - Overview of all clients (Go, JavaScript, Python)
- **[JavaScript WebSocket](../client-js/WS_README.md)** - WebSocket client for Node.js
- **[Python WebSocket](../client-py/WS_README.md)** - WebSocket client for Python
- **[API Documentation](https://partner-api-dev.the-one.io/swagger/)** - Full API reference

## Support

- **Issues:** Report via GitLab issues
- **API Docs:** Check Swagger UI at `{baseURL}/swagger/`
- **Examples:** See `ws_example.go` for complete working example

## License

MIT License - see [LICENSE](../LICENSE) for details

---

**Last Updated:** February 2026  
**Go Version:** 1.21+  
**Status:** Production Ready
