# TheOne Trading API - Client Examples

This directory contains client examples for the TheOne Trading API in various programming languages.

## Available Clients

### [Go Client](./client-go/)
- Complete REST and WebSocket API implementation
- HTTP client with timeouts
- WebSocket client with auto-reconnection
- Detailed debug information
- Gorilla WebSocket library

### [JavaScript/Node.js Client](client-js/)
- Modern ES Modules JavaScript client
- Asynchronous REST and WebSocket methods with Promise/async-await
- Node.js 16+ support
- Event-driven WebSocket client with auto-reconnection
- Dependencies: node-fetch, ws

### [Python Client](client-py/)
- Python client with type hints
- Object-oriented design for REST and WebSocket
- Async/await WebSocket client
- Python 3.7+ support
- Uses requests (REST) and websockets (WS)

## Common Features

### REST API Methods
All clients implement the same REST API methods:

| Method | Description | HTTP |
|--------|-------------|------|
| `getBalances()` / `get_balances()` | Get user balances | GET /api/v1/balances |
| `estimateSwap()` / `estimate_swap()` | Get swap estimation | POST /api/v1/estimate |
| `swap()` | Execute swap | POST /api/v1/swap |
| `getOrderStatus()` / `get_order_status()` | Get order status | GET /api/v1/orders/{id}/status |

### WebSocket API Methods
All clients also support real-time WebSocket connections:

| Method | Description | Purpose |
|--------|-------------|---------|
| `connect()` / `Connect()` | Establish WebSocket connection | Authentication and setup |
| `subscribe()` / `Subscribe()` | Subscribe to data channels | Real-time updates |
| `unsubscribe()` / `Unsubscribe()` | Unsubscribe from channels | Stop receiving updates |
| `close()` / `Close()` | Close WebSocket connection | Cleanup resources |

### Available WebSocket Channels
- `balances` - Real-time balance updates
- `orders:{orderId}` - Real-time order status updates

## Authentication

All clients use the same HMAC-SHA256 authentication scheme:

### Headers
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
- `PATH_WITH_QUERY` - path with query parameters
- `TIMESTAMP_MS` - timestamp in milliseconds
- `NONCE` - unique value in format `{nanoseconds}_{random}`
- `BODY_SHA256_HEX` - SHA256 hash of request body in hex format

## Quick Start

### Go
```bash
cd examples/client
go mod tidy
go run .
```

### JavaScript
```bash
cd examples/client-js
npm install
npm start
```

### Python
```bash
cd examples/client-py
pip install -r requirements.txt
python example.py
```

## Configuration

All clients require similar configuration:

```
API Key: ak_WrXiA7I-VFolEYtZxnsqZTn-tB_f2zqSDEl4XQmqHqA
Secret Key: NwTdHuVVfHA--40pyq_yqJBbscsbtPbD9jRhcU4tRFFQuYagqatzuhzrDu_-xd_q
Base URL: https://partner-api-dev.the-one.io
```

> ⚠️ **Important**: These keys are for testing only. In production, use your own API keys and environment variables.

## Usage Examples

### Getting Balances

**Go:**
```go
balances, err := client.GetBalances(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Balances: %+v\n", balances)
```

**JavaScript:**
```javascript
try {
    const balances = await client.getBalances();
    console.log('Balances:', balances);
} catch (error) {
    console.error('Error:', error.message);
}
```

**Python:**
```python
try:
    balances = client.get_balances()
    print(f"Balances: {balances}")
except APIError as e:
    print(f"Error: {e}")
```

### Executing Swap

**Go:**
```go
swapReq := &SwapRequest{
    From:        "ETH",
    To:          "USDT",
    Amount:      "1.0",
    SlippageBps: 30,
}
idempotencyKey := fmt.Sprintf("swap_%d", time.Now().UnixNano())
response, err := client.Swap(ctx, swapReq, idempotencyKey)
```

**JavaScript:**
```javascript
const swapRequest = {
    from: "ETH",
    to: "USDT",
    amount: "1.0",
    slippage_bps: 30
};
const idempotencyKey = `swap_${Date.now()}_${Math.random()}`;
const response = await client.swap(swapRequest, idempotencyKey);
```

**Python:**
```python
idempotency_key = f"swap_{int(time.time() * 1000)}_{random.randint(0, 999999)}"
response = client.swap(
    from_asset="ETH",
    to_asset="USDT",
    amount="1.0",
    account="0x...",
    slippage_bps=30,
    idempotency_key=idempotency_key
)
```

## Error Handling

All clients provide structured error handling:

### API Errors
```json
{
    "code": "INSUFFICIENT_BALANCE",
    "message": "Insufficient balance for swap",
    "requestId": "req_123456789"
}
```

### Order Statuses
- `PENDING` - order is processing
- `FILLED` - order is executed
- `PARTIAL` - order is partially executed
- `CANCELED` - order is canceled
- `FAILED` - order failed

## Security

### Recommendations
1. **Never store API keys in code**
2. **Use environment variables**
3. **Use HTTPS in production**
4. **Ensure nonce uniqueness**
5. **Implement proper timeout and retry mechanisms**

### Environment Variables

Create a `.env` file:
```
BROKER_API_KEY=your_api_key_here
BROKER_SECRET_KEY=your_secret_key_here
BROKER_BASE_URL=https://partner-api-dev.the-one.io
```

## Client Differences

| Feature | Go | JavaScript | Python |
|---------|----|-----------|\--------|
| Typing | Strong (structs) | Dynamic (JSDoc) | Type hints |
| Async | Context/goroutines | Promises/async-await | Synchronous |
| Dependencies | Stdlib only | node-fetch | requests |
| Debug | Printf | console.log | print |
| Testing | testing package | Jest/Mocha | unittest/pytest |

## Testing

For API testing, it's recommended to:

1. **Use test API keys**
2. **Test with small amounts**
3. **Verify idempotency keys**
4. **Test error handling**

### Test Environment
```
Base URL: https://partner-api-dev.the-one.io
Network: Testnet
Minimum amount: 0.001
```

## Support and Documentation

- **API Documentation**: [Swagger UI](https://partner-api-dev.the-one.io/docs)
- **OpenAPI Spec**: `docs/swagger.yaml`
- **Postman Collection**: Available on request
- **Support**: Through GitLab issues

## License

All clients are provided under MIT license. See LICENSE file for details.

---

**Note**: These examples are for API demonstration purposes. In production, additional error handling, logging, monitoring, and other production-ready features should be added.
