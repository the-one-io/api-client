# REST API Documentation

## Overview

TheOne Trading API provides a RESTful interface for executing trading operations with HMAC-SHA256 authentication support.

**Base URL**: `https://partner-api.the-one.io`

**API Version**: `v1`

## Authentication

All private endpoints require HMAC-SHA256 authentication with the following headers:

| Header | Type | Description |
|--------|------|-------------|
| `X-API-KEY` | string | Your API key |
| `X-API-SECRET` | string | Your secret key (used for signing) |
| `X-API-TIMESTAMP` | string | Request timestamp in milliseconds |
| `X-API-NONCE` | string | Unique string (UUID or timestamp-based) |
| `X-API-SIGN` | string | HMAC-SHA256 signature |

### Canonical String Format

The signature is generated from a canonical string in this format:

```
<HTTP_METHOD>\n<PATH_WITH_QUERY>\n<TIMESTAMP_MS>\n<NONCE>\n<BODY_SHA256_HEX>
```

**Example canonical string**:
```
POST\n/api/v1/estimate\n1732526400000\nnonce_123\nsha256_hash_of_body
```

### Signature Generation Algorithm

1. Create SHA256 hash of request body (empty string for GET requests)
2. Form canonical string
3. Calculate HMAC-SHA256 signature using secret key
4. Convert result to hex format

### Code Examples (Go)

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

func generateSignature(secret, method, path string, timestamp int64, nonce, bodyHash string) string {
    canonical := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
        strings.ToUpper(method),
        path,
        timestamp,
        nonce,
        bodyHash,
    )
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(canonical))
    return hex.EncodeToString(mac.Sum(nil))
}

func hashBody(body []byte) string {
    hash := sha256.Sum256(body)
    return hex.EncodeToString(hash[:])
}
```

### Code Examples (Python)

```python
import hashlib
import hmac
import time
import json

def generate_signature(secret, method, path, timestamp, nonce, body_hash):
    canonical = f"{method.upper()}\n{path}\n{timestamp}\n{nonce}\n{body_hash}"
    signature = hmac.new(
        secret.encode('utf-8'),
        canonical.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    return signature

def hash_body(body):
    if isinstance(body, dict):
        body = json.dumps(body, separators=(',', ':')).encode('utf-8')
    return hashlib.sha256(body).hexdigest()
```

### Code Examples (JavaScript)

```javascript
const crypto = require('crypto');

function generateSignature(secret, method, path, timestamp, nonce, bodyHash) {
    const canonical = `${method.toUpperCase()}\n${path}\n${timestamp}\n${nonce}\n${bodyHash}`;
    return crypto
        .createHmac('sha256', secret)
        .update(canonical)
        .digest('hex');
}

function hashBody(body) {
    const bodyStr = typeof body === 'object' ? JSON.stringify(body) : body;
    return crypto
        .createHash('sha256')
        .update(bodyStr)
        .digest('hex');
}
```

## Rate Limiting

| Endpoint Type | Limit |
|--------------|-------|
| Public | 100 requests/minute |
| Private | 1000 requests/minute |

## Error Format

```json
{
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "request_id": "uuid-request-id"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Invalid authentication |
| `INVALID_REQUEST` | 400 | Invalid request |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMIT` | 429 | Rate limit exceeded |
| `INVALID_SIGNATURE` | 401 | Invalid signature |
| `EXPIRED_REQUEST` | 401 | Expired request (timestamp) |
| `INVALID_NONCE` | 401 | Nonce already used or expired |
| `MISSING_API_KEY` | 401 | Missing API key |
| `DUPLICATE_REQUEST` | 409 | Duplicate request |
| `SWAP_ERROR` | 400 | Swap execution error |

---

## Public Endpoints

### Health Check

**GET** `/healthz`

Returns service health status.

**Response**:
```json
{
  "status": "ok",
  "time": 1732526400000
}
```

### Server Time

**GET** `/api/v1/time`

Returns current server time in milliseconds (for timestamp synchronization).

**Response**:
```json
{
  "serverTime": 1732526400000
}
```

### Version Information

**GET** `/version`

Returns API version information.

**Response**:
```json
{
  "version": "1.0.0",
  "commit": "abc123def",
  "buildTime": "2024-01-15T10:30:00Z",
  "goVersion": "go1.22.0"
}
```

---

## Private Endpoints

### 1. Swap Estimate

**POST** `/api/v1/estimate`

Get a price estimate for a swap operation.

**Headers**:
- `Content-Type: application/json`
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Request Body**:
```json
{
  "from": "ETH",
  "to": "USDT",
  "amount": "1.5",
  "network": "ETH",
  "account": "0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
  "filter": ["binance", "gate"]
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from` | string | Yes | Asset to swap from |
| `to` | string | Yes | Asset to receive |
| `amount` | string | Yes | Amount to swap |
| `network` | string | No | Network (ETH, BSC, etc.) |
| `account` | string | No | Wallet address |
| `filter` | []string | No | Exchange/provider filter |

**Response** (200 OK):
```json
{
  "route": [
    {
      "from_asset": "ETH",
      "to_asset": "USDT",
      "amount_in": "1.5",
      "amount_out": "3245.67",
      "pool": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
    }
  ],
  "price": "2163.78",
  "expectedOut": "3245.67",
  "fee": "0.003",
  "expiresAt": 1732526400000
}
```

**cURL Example**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
METHOD="POST"
PATH="/api/v1/estimate"
BODY='{"from":"ETH","to":"USDT","amount":"1.5"}'
BODY_HASH=$(echo -n "$BODY" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X POST "https://partner-api.the-one.io/api/v1/estimate" \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE" \
  -d "$BODY"
```

---

### 2. Execute Swap

**POST** `/api/v1/swap`

Execute a swap operation. This endpoint is idempotent.

**Headers**:
- `Content-Type: application/json`
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`
- `Idempotency-Key: {unique_key}` - for duplicate prevention

**Request Body**:
```json
{
  "from": "ETH",
  "to": "USDT",
  "amount": "1.5",
  "slippage_bps": 30,
  "clientOrderId": "client_ord_123",
  "filter": ["binance", "gate"]
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from` | string | Yes | Asset to swap from |
| `to` | string | Yes | Asset to receive |
| `amount` | string | Yes | Amount to swap |
| `slippage_bps` | integer | No | Allowed slippage in basis points (0-10000) |
| `clientOrderId` | string | No | Your order ID for tracking |
| `filter` | []string | No | Exchange/provider filter |

**Response** (200 OK):
```json
{
  "orderId": "ord_123456",
  "status": "PENDING"
}
```

**Possible Statuses**:
- `PENDING` - Order created, awaiting execution
- `PROCESSING` - Order is being executed
- `FILLED` - Order completed
- `FAILED` - Order failed
- `CANCELLED` - Order cancelled

**cURL Example**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
IDEMPOTENCY_KEY="swap_$(uuidgen)"
METHOD="POST"
PATH="/api/v1/swap"
BODY='{"from":"ETH","to":"USDT","amount":"1.5","slippage_bps":30}'
BODY_HASH=$(echo -n "$BODY" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X POST "https://partner-api.the-one.io/api/v1/swap" \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d "$BODY"
```

---

### 3. Order Status

**GET** `/api/v1/orders/{id}/status`

Get order status by ID or clientOrderId.

**Path Parameters**:
- `id` - Order ID or Client Order ID

**Query Parameters**:
- `clientOrderId` (optional) - Client Order ID as alternative to path parameter

**Headers**:
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Response** (200 OK):
```json
{
  "orderId": "ord_123456",
  "clientOrderId": "client_ord_123",
  "status": "FILLED",
  "filledOut": "3245.67",
  "txHash": "0xabc123def456789abc123def456789abc123def456789abc123def456789abc123de",
  "updatedAt": 1732526402000
}
```

**cURL Example**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
ORDER_ID="ord_123456"
METHOD="GET"
PATH="/api/v1/orders/$ORDER_ID/status"
BODY_HASH=$(echo -n "" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X GET "https://partner-api.the-one.io/api/v1/orders/$ORDER_ID/status" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE"
```

---

### 4. Balances

**GET** `/api/v1/balances`

Get user balances for all assets.

**Headers**:
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Response** (200 OK):
```json
{
  "balances": [
    {
      "asset": "USDT",
      "total": "100.00",
      "locked": "0"
    },
    {
      "asset": "ETH",
      "total": "2.5",
      "locked": "0.5"
    },
    {
      "asset": "BTC",
      "total": "0.1",
      "locked": "0"
    }
  ]
}
```

**Balance Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `asset` | string | Asset symbol |
| `total` | string | Total balance |
| `locked` | string | Locked balance (in active orders) |

**cURL Example**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
METHOD="GET"
PATH="/api/v1/balances"
BODY_HASH=$(echo -n "" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X GET "https://partner-api.the-one.io/api/v1/balances" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE"
```

---

## API Key Management

### List API Keys

**GET** `/broker/v1/keys`

Get list of all user API keys.

**Query Parameters**:
- `limit` (optional, default: 20, max: 100) - Number of records
- `offset` (optional, default: 0) - Offset for pagination

**Response** (200 OK):
```json
{
  "api_keys": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "key": "api_key_abc123",
      "name": "Production Key",
      "description": "Main production API key",
      "is_active": true,
      "permissions": ["read", "trade"],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "last_used_at": "2024-01-20T14:25:00Z",
      "expires_at": "2025-01-15T10:30:00Z",
      "metadata": {
        "ip_whitelist": ["192.168.1.1", "10.0.0.1"]
      }
    }
  ],
  "pagination": {
    "count": 1,
    "limit": 20,
    "offset": 0
  }
}
```

### Create API Key

**POST** `/broker/v1/keys`

Create a new API key.

**Request Body**:
```json
{
  "name": "Production Key",
  "description": "Main production API key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "expires_at": "2025-01-15T10:30:00Z",
  "metadata": {
    "ip_whitelist": ["192.168.1.1"]
  }
}
```

**Response** (201 Created):
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "key": "api_key_abc123",
  "secret": "api_secret_xyz789",
  "name": "Production Key",
  "description": "Main production API key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2025-01-15T10:30:00Z"
}
```

**⚠️ Important**: The secret key (`secret`) is only returned once during creation!

### Get API Key

**GET** `/broker/v1/keys/{uuid}`

Get information about specific API key.

**Response** (200 OK):
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "key": "api_key_abc123",
  "name": "Production Key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Delete API Key

**DELETE** `/broker/v1/keys/{uuid}`

Delete an API key.

**Response** (204 No Content)

---

## Best Practices

### Security

1. **Never expose secret key** - store it securely
2. **Use HTTPS** - always use secure connection
3. **Key rotation** - regularly update API keys
4. **IP Whitelist** - restrict API access to specific IP addresses
5. **Minimal permissions** - grant only necessary permissions

### Error Handling

1. **Check error code** - use `code` for programmatic handling
2. **Log request_id** - use for issue tracking
3. **Request retries** - use exponential backoff for retries
4. **Handle 429** - implement waiting logic for rate limiting

### Performance

1. **Caching** - cache server time and other static data
2. **Batch requests** - combine requests when possible
3. **WebSocket** - use for real-time updates instead of polling
4. **Idempotency Keys** - use for critical operations

### Testing

1. **Test environment** - use test keys for development
2. **Logging** - log all requests and responses
3. **Monitoring** - track rate limits and errors
4. **Alerts** - set up notifications for critical errors

---

## Client Examples

Ready-to-use clients are available for different programming languages:

- **Go**: `examples/client/client-go/`
- **Python**: `examples/client/client-py/`
- **JavaScript**: `examples/client/client-js/`

Each client includes full authentication implementation and all endpoints.

---

## Support

For assistance:
- Email: support@the-one.io
- Telegram: [@TheOneLiveSupportBot](https://t.me/TheOneLiveSupportBot)
- Documentation: [https://partner-api.the-one.io/swagger/index.html](https://partner-api.the-one.io/swagger/index.html)
