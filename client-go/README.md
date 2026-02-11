# TheOne Trading API - Go Client Example

This example demonstrates how to use the TheOne Trading API with a Go client supporting both REST and WebSocket connections.

## Features

- ✅ Complete REST API implementation
- ✅ WebSocket API with real-time updates
- ✅ Environment variables support (.env file)
- ✅ HMAC-SHA256 authentication
- ✅ Idempotency support for swaps
- ✅ Auto-reconnection for WebSocket
- ✅ Type-safe Go structs
- ✅ Context support for timeouts
- ✅ Comprehensive error handling

## Installation

```bash
cd client-go
go mod download
go mod tidy
```

## Configuration

### Using Environment Variables (Recommended)

Create a `.env` file in the `client-go` directory:

```bash
BROKER_API_KEY=your_api_key_here
BROKER_SECRET_KEY=your_secret_key_here
BROKER_BASE_URL=https://partner-api-dev.the-one.io
```

The client will automatically load these variables on startup.

### Example .env file

```bash
BROKER_API_KEY=ak_WrXiA7I-VFolEYtZxnsqZTn-tB_f2zqSDEl4XQmqHqA
BROKER_SECRET_KEY=NwTdHuVVfHA--40pyq_yqJBbscsbtPbD9jRhcU4tRFFQuYagqatzuhzrDu_-xd_q
BROKER_BASE_URL=https://partner-api-dev.the-one.io
```

> ⚠️ **Security Note**: Never commit `.env` files to version control. Use `.env.example` as a template.

## Quick Start

### Using Makefile (Recommended)

```bash
# Run REST API client example
make run

# Run WebSocket client example
make run-ws

# Install dependencies only
make deps

# See all available commands
make help
```

### Manual Run

```bash
# REST API client
go run main.go http_client.go

# WebSocket client
go run ws_main.go ws_client.go ws_example.go
```

## REST API Methods

The client implements all main API methods:

| Method | Description | HTTP Endpoint |
|--------|-------------|---------------|
| `GetBalances(ctx)` | Get user balances | GET /api/v1/balances |
| `EstimateSwap(ctx, req)` | Get swap estimation | POST /api/v1/estimate |
| `Swap(ctx, req, idempotencyKey)` | Execute swap | POST /api/v1/swap |
| `GetOrderStatus(ctx, orderID, clientOrderID)` | Get order status | GET /api/v1/orders/{id}/status |

## WebSocket API Methods

The WebSocket client supports real-time updates:

| Method | Description | Purpose |
|--------|-------------|---------|
| `Connect()` | Establish WebSocket connection | Authentication and setup |
| `Subscribe(channel, callback)` | Subscribe to data channels | Real-time updates |
| `Unsubscribe(channel)` | Unsubscribe from channels | Stop receiving updates |
| `Close()` | Close WebSocket connection | Cleanup resources |

### Available WebSocket Channels

- `balances` - Real-time balance updates
- `orders:{orderId}` - Real-time order status updates for specific order

## Usage Examples

### REST API

#### Setup Client

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }

    // Get credentials from environment
    apiKey := os.Getenv("BROKER_API_KEY")
    secretKey := os.Getenv("BROKER_SECRET_KEY")
    baseURL := os.Getenv("BROKER_BASE_URL")

    // Validate
    if apiKey == "" || secretKey == "" || baseURL == "" {
        log.Fatal("Missing required environment variables")
    }

    // Create clients
    httpClient := NewDefaultHTTPClient()
    client := NewBrokerClient(apiKey, secretKey, baseURL, httpClient)
    
    ctx := context.Background()
    
    // Use client...
}
```

#### Get Balances

```go
balances, err := client.GetBalances(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Balances:")
for _, balance := range balances.Balances {
    fmt.Printf("  %s: %s (locked: %s)\n", 
        balance.Asset, balance.Total, balance.Locked)
}
```

#### Estimate Swap

```go
estimateReq := &EstimateRequestHTTP{
    From:   "USDT",
    To:     "TRX",
    Amount: "10",
    Filter: []string{"binance", "gate"}, // Optional: specify exchanges
}

estimate, err := client.EstimateSwap(ctx, estimateReq)
if err != nil {
    log.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Price: %s\n", estimate.Price)
fmt.Printf("Expected Out: %s\n", estimate.ExpectedOut)
fmt.Printf("Expires At: %d\n", estimate.ExpiresAt)

// Route information
for i, step := range estimate.Route {
    fmt.Printf("Step %d: %s - %s -> %s\n", 
        i+1, step.Exchange, step.FromAsset, step.ToAsset)
}
```

#### Execute Swap

```go
swapReq := &SwapRequestHTTP{
    From:        "USDT",
    To:          "TRX",
    Amount:      "10",
    SlippageBps: 30, // 0.3% slippage tolerance
    Filter:      []string{"binance", "gate"},
}

// Important: Use unique idempotency key for each swap
idempotencyKey := fmt.Sprintf("swap_%d", time.Now().UnixNano())

swapResponse, err := client.Swap(ctx, swapReq, idempotencyKey)
if err != nil {
    log.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Swap created:\n")
fmt.Printf("  Order ID: %s\n", swapResponse.OrderID)
fmt.Printf("  Status: %s\n", swapResponse.Status)
```

#### Check Order Status

```go
// Poll order status until filled or failed
for i := 0; i < 10; i++ {
    orderStatus, err := client.GetOrderStatus(ctx, swapResponse.OrderID, nil)
    if err != nil {
        log.Printf("Error: %v\n", err)
        time.Sleep(2 * time.Second)
        continue
    }

    fmt.Printf("Order Status: %s\n", orderStatus.Status)
    
    if orderStatus.Status == "FILLED" || orderStatus.Status == "filled" {
        fmt.Printf("Order filled! Amount: %s\n", orderStatus.FilledOut)
        break
    }
    
    if orderStatus.Status == "FAILED" {
        fmt.Println("Order failed!")
        break
    }
    
    time.Sleep(2 * time.Second)
}
```

### WebSocket API

#### Setup WebSocket Client

```go
import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }

    apiKey := os.Getenv("BROKER_API_KEY")
    secretKey := os.Getenv("BROKER_SECRET_KEY")
    baseURL := os.Getenv("BROKER_BASE_URL")

    // Convert HTTP URL to WebSocket URL
    wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
    wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
    wsURL += "/ws/v1/stream"

    // Create WebSocket client
    wsClient := NewWSClient(apiKey, secretKey, wsURL)

    // Connect
    if err := wsClient.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // Use client...
}
```

#### Subscribe to Balances

```go
err := wsClient.Subscribe("balances", func(msg *WSMessage) {
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

#### Subscribe to Order Updates

```go
orderID := "12345"
orderChannel := "orders:" + orderID

err := wsClient.Subscribe(orderChannel, func(msg *WSMessage) {
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

#### Graceful Shutdown

```go
// Set up signal handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Wait for interrupt
select {
case <-sigChan:
    fmt.Println("\nShutting down...")
case <-wsClient.ctx.Done():
    fmt.Println("\nConnection lost...")
}

// Unsubscribe and close
wsClient.Unsubscribe("balances")
wsClient.Unsubscribe(orderChannel)
wsClient.Close()
```

## Authentication

API uses HMAC-SHA256 authentication with the following headers:

- `X-API-KEY` - API key
- `X-API-TIMESTAMP` - timestamp in milliseconds
- `X-API-NONCE` - unique nonce value
- `X-API-SIGN` - HMAC-SHA256 signature

### Signature Format

The signature is created from a canonical string:
```
<HTTP_METHOD>\n<PATH_WITH_QUERY>\n<TIMESTAMP_MS>\n<NONCE>\n<BODY_SHA256_HEX>
```

Where:
- `HTTP_METHOD` - uppercase method (GET, POST)
- `PATH_WITH_QUERY` - path with query parameters (e.g., `/api/v1/balances`)
- `TIMESTAMP_MS` - current timestamp in milliseconds
- `NONCE` - unique value in format `{nanoseconds}_{random}`
- `BODY_SHA256_HEX` - SHA256 hash of request body in hex format (empty string for GET)

## Project Structure

```
client-go/
├── main.go           # REST API client and example
├── http_client.go    # HTTP client implementation
├── ws_main.go        # WebSocket entry point
├── ws_client.go      # WebSocket client implementation
├── ws_example.go     # WebSocket usage example
├── go.mod            # Go module definition
├── go.sum            # Dependency checksums
├── Makefile          # Build and run commands
├── .env              # Environment variables (create from .env.example)
├── .env.example      # Example environment configuration
└── README.md         # This file
```

## Error Handling

The client provides structured error handling:

### API Errors

```go
balances, err := client.GetBalances(ctx)
if err != nil {
    if apiErr, ok := err.(*APIError); ok {
        // API error from server
        fmt.Printf("API Error [%s]: %s\n", apiErr.Code, apiErr.Message)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)
    } else {
        // Network or other error
        fmt.Printf("Network Error: %v\n", err)
    }
    return
}
```

### Common Error Codes

- `INSUFFICIENT_BALANCE` - Not enough balance for swap
- `INVALID_AMOUNT` - Amount is invalid or too small
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `UNAUTHORIZED` - Invalid API credentials
- `ORDER_NOT_FOUND` - Order ID not found

## Order Statuses

- `PENDING` - order is processing
- `FILLED` / `filled` - order is executed (case-insensitive)
- `PARTIAL_FILLED` / `partial_filled` - order is partially executed
- `CANCELED` - order is canceled
- `FAILED` - order failed

## Makefile Commands

The Makefile provides convenient commands:

```bash
make deps          # Download and tidy dependencies
make run           # Run REST API client example
make run-ws        # Run WebSocket client example
make clean         # Clean build directory
make build-all     # Build for all platforms
make build-windows # Build for Windows
make build-macos   # Build for macOS (amd64 and arm64)
make build-linux   # Build for Linux
make help          # Show all available commands
```

## Dependencies

- **Go 1.21+** - Programming language
- **github.com/gorilla/websocket** - WebSocket library
- **github.com/joho/godotenv** - Environment variables loader
- Standard library only for REST client

Install dependencies:
```bash
go mod download
go mod tidy
```

## Security Best Practices

1. **Never commit API keys to version control**
   - Use `.env` files (add to `.gitignore`)
   - Use environment variables in production

2. **Use HTTPS in production**
   - Always use `https://` URLs, not `http://`

3. **Implement proper timeout handling**
   - Use context with timeout for all API calls
   - Set reasonable timeout values (e.g., 30 seconds)

4. **Ensure nonce uniqueness**
   - The client automatically generates unique nonces
   - Format: `{timestamp_nanoseconds}_{random_number}`

5. **Validate idempotency keys**
   - Use unique keys for each swap operation
   - Format: `swap_{timestamp}_{unique_id}`

6. **Handle errors appropriately**
   - Check error types (API vs network errors)
   - Implement retry logic with exponential backoff
   - Log errors with request IDs for debugging

## Testing

### Test Environment

```
Base URL: https://partner-api-dev.the-one.io
Network: Testnet
Minimum swap amount: Varies by asset
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

## Troubleshooting

### Common Issues

1. **"missing LC_UUID" error**
   - Solution: Use `CGO_ENABLED=0` (already configured in Makefile)

2. **"unauthorized" error**
   - Check API keys in `.env` file
   - Verify signature generation
   - Check system clock synchronization

3. **"connection refused" error**
   - Verify `BROKER_BASE_URL` is correct
   - Check network connectivity
   - Ensure API server is running

4. **WebSocket disconnect**
   - Check internet connection
   - Verify WebSocket URL format
   - The client has auto-reconnection built-in

### Debug Mode

Enable debug logging:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Println("Debug info:", variable)
```

## Support

- **Documentation**: See main [README.md](../README.md)
- **API Docs**: Check Swagger UI at `{baseURL}/swagger/`
- **Issues**: Report via GitLab issues

## License

MIT License - see [LICENSE](../LICENSE) for details

## Additional Resources

- [WebSocket README](./WS_README.md) - Detailed WebSocket documentation
- [Main README](../README.md) - Overview of all clients
- [API Documentation](https://partner-api-dev.the-one.io/swagger/) - Full API reference

---

**Last Updated**: February 2026  
**Go Version**: 1.21+  
**Status**: Production Ready
