# WebSocket API Documentation

## Overview

TheOne Trading API provides a WebSocket interface for real-time communication, including event subscriptions and trading operations execution.

**WebSocket URL**: `wss://partner-api.the-one.io/ws/v1/stream`

**Protocol**: WebSocket (RFC 6455)

## Connection

```javascript
const ws = new WebSocket('wss://partner-api.the-one.io/ws/v1/stream');
```

## Authentication

The first message after connection **must** be an authentication message.

### Authentication Message Format

```json
{
  "op": "auth",
  "key": "your_api_key",
  "ts": 1732526400000,
  "nonce": "unique_nonce_123",
  "sig": "hmac_signature"
}
```

### Authentication Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `op` | string | Yes | Operation (always "auth" for authentication) |
| `key` | string | Yes | Your API key |
| `ts` | integer | Yes | Timestamp in milliseconds |
| `nonce` | string | Yes | Unique nonce |
| `sig` | string | Yes | HMAC-SHA256 signature |

### Signature Generation for Authentication

WebSocket authentication uses the same HMAC-SHA256 algorithm as REST API:

**Canonical string**:
```
WS\n/ws/v1/stream\n<TIMESTAMP>\n<NONCE>\n<EMPTY_BODY_HASH>
```

**Example (JavaScript)**:
```javascript
const crypto = require('crypto');

function authenticateWebSocket(apiKey, secret) {
  const timestamp = Date.now();
  const nonce = `ws_${timestamp}_${Math.random()}`;
  
  // Empty body for auth
  const bodyHash = crypto.createHash('sha256').update('').digest('hex');
  
  // Canonical string
  const canonical = `WS\n/ws/v1/stream\n${timestamp}\n${nonce}\n${bodyHash}`;
  
  // Generate signature
  const signature = crypto
    .createHmac('sha256', secret)
    .update(canonical)
    .digest('hex');
  
  return {
    op: 'auth',
    key: apiKey,
    ts: timestamp,
    nonce: nonce,
    sig: signature
  };
}

// Usage
const authMsg = authenticateWebSocket('your_api_key', 'your_secret');
ws.send(JSON.stringify(authMsg));
```

**Example (Python)**:
```python
import hashlib
import hmac
import time
import json

def authenticate_websocket(api_key, secret):
    timestamp = int(time.time() * 1000)
    nonce = f"ws_{timestamp}"
    
    # Empty body for auth
    body_hash = hashlib.sha256(b'').hexdigest()
    
    # Canonical string
    canonical = f"WS\n/ws/v1/stream\n{timestamp}\n{nonce}\n{body_hash}"
    
    # Generate signature
    signature = hmac.new(
        secret.encode('utf-8'),
        canonical.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    return {
        'op': 'auth',
        'key': api_key,
        'ts': timestamp,
        'nonce': nonce,
        'sig': signature
    }

# Usage
auth_msg = authenticate_websocket('your_api_key', 'your_secret')
ws.send(json.dumps(auth_msg))
```

**Example (Go)**:
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

func authenticateWebSocket(apiKey, secret string) ([]byte, error) {
    timestamp := time.Now().UnixMilli()
    nonce := fmt.Sprintf("ws_%d", timestamp)
    
    // Empty body for auth
    bodyHash := sha256.Sum256([]byte{})
    bodyHashHex := hex.EncodeToString(bodyHash[:])
    
    // Canonical string
    canonical := fmt.Sprintf("WS\n/ws/v1/stream\n%d\n%s\n%s",
        timestamp, nonce, bodyHashHex)
    
    // Generate signature
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(canonical))
    signature := hex.EncodeToString(mac.Sum(nil))
    
    authMsg := map[string]interface{}{
        "op":    "auth",
        "key":   apiKey,
        "ts":    timestamp,
        "nonce": nonce,
        "sig":   signature,
    }
    
    return json.Marshal(authMsg)
}

// Usage
authMsg, _ := authenticateWebSocket("your_api_key", "your_secret")
ws.WriteMessage(websocket.TextMessage, authMsg)
```

### Authentication Response

**Successful authentication**:
```json
{
  "op": "auth",
  "data": {
    "status": "authenticated"
  }
}
```

**Authentication error**:
```json
{
  "error": "Invalid signature"
}
```

---

## Subscriptions

After successful authentication, you can subscribe to event channels.

### Subscribe to Channel

```json
{
  "op": "subscribe",
  "ch": "channel_name"
}
```

**Response**:
```json
{
  "op": "subscribe",
  "ch": "channel_name",
  "data": {
    "status": "subscribed"
  }
}
```

### Unsubscribe from Channel

```json
{
  "op": "unsubscribe",
  "ch": "channel_name"
}
```

**Відповідь**:
```json
{
  "op": "unsubscribe",
  "ch": "channel_name",
  "data": {
    "status": "unsubscribed"
  }
}
```

---

## Available Channels

### 1. Balances (`balances`)

Receive real-time balance updates.

**Subscribe**:
```json
{
  "op": "subscribe",
  "ch": "balances"
}
```

**Update notification**:
```json
{
  "ch": "balances",
  "data": [
    {
      "asset": "USDT",
      "total": "101.23",
      "locked": "0"
    },
    {
      "asset": "ETH",
      "total": "2.6",
      "locked": "0.4"
    },
    {
      "asset": "BTC",
      "total": "0.11",
      "locked": "0"
    }
  ]
}
```

### 2. Order Status (`orders:<orderId>`)

Receive real-time updates about specific order status.

**Subscribe**:
```json
{
  "op": "subscribe",
  "ch": "orders:ord_123456"
}
```

**Update message**:
```json
{
  "ch": "orders:ord_123456",
  "data": {
    "orderId": "ord_123456",
    "status": "FILLED",
    "filledOut": "3245.67",
    "txHash": "0xabc123def456789abc123def456789abc123def456789abc123def456789abc123de",
    "updatedAt": 1732526402000
  }
}
```

**Possible statuses**:
- `PENDING` - Awaiting execution
- `PROCESSING` - Being executed
- `FILLED` - Completed
- `FAILED` - Failed
- `CANCELLED` - Cancelled

---

## Executing Operations via WebSocket

WebSocket API supports executing trading operations with authentication of each message.

### Signing Operation Messages

For operations (estimate, swap, order_status, balances), each message must be signed:

**Canonical string**:
```
WS\n/ws/v1/<OPERATION>\n<TIMESTAMP>\n<NONCE>\n<DATA_BODY_SHA256>
```

**Example sign generation (JavaScript)**:
```javascript
function signMessage(secret, operation, data, timestamp, nonce) {
  const dataStr = JSON.stringify(data);
  const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
  
  const canonical = `WS\n/ws/v1/${operation}\n${timestamp}\n${nonce}\n${bodyHash}`;
  
  return crypto
    .createHmac('sha256', secret)
    .update(canonical)
    .digest('hex');
}
```

### 1. Swap Estimate

**Request**:
```json
{
  "op": "estimate",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "from": "ETH",
    "to": "USDT",
    "amount": "1.5",
    "network": "ETH",
    "filter": ["binance", "gate"]
  }
}
```

**Response**:
```json
{
  "op": "estimate",
  "data": {
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
}
```

### 2. Execute Swap

**Request**:
```json
{
  "op": "swap",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "from": "ETH",
    "to": "USDT",
    "amount": "1.5",
    "slippage_bps": 30,
    "clientOrderId": "client_ord_123",
    "filter": ["binance", "gate"]
  }
}
```

**Response**:
```json
{
  "op": "swap",
  "data": {
    "orderId": "ord_123456",
    "status": "PENDING"
  }
}
```

### 3. Order Status

**Request**:
```json
{
  "op": "order_status",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "id": "ord_123456"
  }
}
```

**Response**:
```json
{
  "op": "order_status",
  "data": {
    "orderId": "ord_123456",
    "status": "FILLED",
    "filledOut": "3245.67",
    "txHash": "0xabc123...",
    "updatedAt": 1732526402000
  }
}
```

### 4. Get Balances

**Request**:
```json
{
  "op": "balances",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {}
}
```

**Response**:
```json
{
  "op": "balances",
  "data": {
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
      }
    ]
  }
}
```

---

## Error Format

```json
{
  "error": "Error message description"
}
```

**Common errors**:
- `"Not authenticated"` - Attempt to execute operation without authentication
- `"Invalid signature"` - Invalid message signature
- `"Invalid API key"` - Invalid API key
- `"Request timestamp outside allowed window"` - Expired timestamp
- `"Nonce has already been used or is expired"` - Nonce already used
- `"Unknown operation: <op>"` - Unknown operation
- `"Invalid channel format"` - Invalid channel format

---

## Complete Example (JavaScript)

```javascript
const WebSocket = require('ws');
const crypto = require('crypto');

class TradingWebSocketClient {
  constructor(apiKey, secret, url = 'wss://partner-api.the-one.io/ws/v1/stream') {
    this.apiKey = apiKey;
    this.secret = secret;
    this.url = url;
    this.ws = null;
    this.authenticated = false;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      
      this.ws.on('open', () => {
        console.log('WebSocket connected');
        this.authenticate()
          .then(() => resolve())
          .catch(err => reject(err));
      });
      
      this.ws.on('message', (data) => {
        this.handleMessage(JSON.parse(data.toString()));
      });
      
      this.ws.on('error', (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      });
      
      this.ws.on('close', () => {
        console.log('WebSocket disconnected');
        this.authenticated = false;
      });
    });
  }

  authenticate() {
    return new Promise((resolve, reject) => {
      const timestamp = Date.now();
      const nonce = `ws_auth_${timestamp}`;
      const bodyHash = crypto.createHash('sha256').update('').digest('hex');
      
      const canonical = `WS\n/ws/v1/stream\n${timestamp}\n${nonce}\n${bodyHash}`;
      const signature = crypto
        .createHmac('sha256', this.secret)
        .update(canonical)
        .digest('hex');
      
      const authMsg = {
        op: 'auth',
        key: this.apiKey,
        ts: timestamp,
        nonce: nonce,
        sig: signature
      };
      
      this.ws.send(JSON.stringify(authMsg));
      
      // Wait for auth response
      const handler = (msg) => {
        if (msg.op === 'auth') {
          if (msg.data?.status === 'authenticated') {
            this.authenticated = true;
            console.log('Authenticated successfully');
            resolve();
          } else if (msg.error) {
            reject(new Error(msg.error));
          }
        }
      };
      
      this.ws.once('message', (data) => {
        handler(JSON.parse(data.toString()));
      });
    });
  }

  subscribe(channel) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    this.ws.send(JSON.stringify({
      op: 'subscribe',
      ch: channel
    }));
  }

  unsubscribe(channel) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    this.ws.send(JSON.stringify({
      op: 'unsubscribe',
      ch: channel
    }));
  }

  async estimate(from, to, amount, options = {}) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    const timestamp = Date.now();
    const nonce = `estimate_${timestamp}`;
    
    const data = {
      from,
      to,
      amount,
      ...options
    };
    
    const dataStr = JSON.stringify(data);
    const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
    const canonical = `WS\n/ws/v1/estimate\n${timestamp}\n${nonce}\n${bodyHash}`;
    const signature = crypto
      .createHmac('sha256', this.secret)
      .update(canonical)
      .digest('hex');
    
    this.ws.send(JSON.stringify({
      op: 'estimate',
      ts: timestamp,
      nonce: nonce,
      sig: signature,
      data: data
    }));
  }

  async swap(from, to, amount, options = {}) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    const timestamp = Date.now();
    const nonce = `swap_${timestamp}`;
    
    const data = {
      from,
      to,
      amount,
      ...options
    };
    
    const dataStr = JSON.stringify(data);
    const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
    const canonical = `WS\n/ws/v1/swap\n${timestamp}\n${nonce}\n${bodyHash}`;
    const signature = crypto
      .createHmac('sha256', this.secret)
      .update(canonical)
      .digest('hex');
    
    this.ws.send(JSON.stringify({
      op: 'swap',
      ts: timestamp,
      nonce: nonce,
      sig: signature,
      data: data
    }));
  }

  handleMessage(msg) {
    console.log('Received message:', msg);
    
    if (msg.error) {
      console.error('Error:', msg.error);
      return;
    }
    
    // Handle different message types
    if (msg.ch) {
      // Channel update
      console.log(`Update from ${msg.ch}:`, msg.data);
    } else if (msg.op) {
      // Operation response
      console.log(`Response for ${msg.op}:`, msg.data);
    }
  }

  close() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// Usage example
(async () => {
  const client = new TradingWebSocketClient('your_api_key', 'your_secret');
  
  try {
    await client.connect();
    
    // Subscribe to balances
    client.subscribe('balances');
    
    // Subscribe to specific order updates
    client.subscribe('orders:ord_123456');
    
    // Get estimate
    await client.estimate('ETH', 'USDT', '1.5');
    
    // Execute swap
    await client.swap('ETH', 'USDT', '1.5', {
      slippage_bps: 30,
      clientOrderId: 'my_order_123'
    });
    
    // Keep connection open
    setTimeout(() => {
      client.close();
    }, 60000); // Close after 1 minute
    
  } catch (error) {
    console.error('Error:', error);
  }
})();
```

---

## Best Practices

### Connection

1. **Reconnection Logic** - Implement automatic reconnection
2. **Heartbeat/Ping** - Use WebSocket ping/pong to check connection
3. **Timeout** - Set timeout for operations
4. **Error Handling** - Handle all error types

### Security

1. **Secure Connection** - Always use WSS (WebSocket Secure)
2. **Nonce Uniqueness** - Use unique nonce for each message
3. **Timestamp Validation** - Synchronize time with server
4. **Secret Protection** - Never transmit secret over connection

### Performance

1. **Batching** - Combine subscriptions when possible
2. **Selective Subscriptions** - Subscribe only to needed channels
3. **Message Queuing** - Use queue for sending messages
4. **Connection Pooling** - Reuse connections

---

## Limitations

- Maximum **100 concurrent WebSocket connections** per API key
- Maximum **50 subscriptions** per connection
- Maximum **100 messages/second** per connection
- Automatic disconnect after **5 minutes of inactivity**
- Maximum message size: **1 MB**

---

## Support

For assistance:
- Email: support@the-one.io
- Telegram: [@TheOneLiveSupportBot](https://t.me/TheOneLiveSupportBot)
- Документація: [https://partner-api.the-one.io/swagger/index.html](https://partner-api.the-one.io/swagger/index.html)
