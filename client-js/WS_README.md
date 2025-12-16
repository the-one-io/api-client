# WebSocket Client for JavaScript/Node.js

WebSocket client for Broker Trading API in JavaScript/Node.js with authentication and real-time support.

## Installation

```bash
npm install
```

## Dependencies

- `ws` - WebSocket library for Node.js
- `node-fetch` - For HTTP requests (already installed for REST client)

## Usage

### Simple Connection

```javascript
import { BrokerWSClient } from './ws-client.js';

async function main() {
    // API keys
    const apiKey = 'your_api_key';
    const secretKey = 'your_secret_key';
    const wsURL = 'wss://api.example.com/ws';
    
    // Create client
    const client = new BrokerWSClient(apiKey, secretKey, wsURL);
    
    // Connect
    await client.connect();
    
    // Subscribe to balances
    await client.subscribe('balances', (message) => {
        console.log('Balances update:', message.data);
    });
    
    // Trading operations
    await client.getBalances();
    await client.estimateSwap('100', 'ETH', 'USDT');
    await client.doSwap('100', 'ETH', 'USDT');
    await client.getOrderStatus('ord_12345678');
    
    // Graceful shutdown
    process.on('SIGINT', () => {
        client.close();
        process.exit(0);
    });
}

main().catch(console.error);
```

### Channel Subscriptions

#### User Balances

```javascript
await client.subscribe('balances', (message) => {
    console.log('Received balances:', message.data);
});
```

#### Order Status

```javascript
const orderID = 'ord_12345678';
const channel = `orders:${orderID}`;

await client.subscribe(channel, (message) => {
    console.log(`Order ${orderID} status:`, message.data);
});
```

### Trading Operations

#### Get Balances

```javascript
// Request current balances (signed message)
await client.getBalances();
```

#### Estimate Swap

```javascript
// Estimate swap cost (signed message)
await client.estimateSwap('100', 'ETH', 'USDT');
```

#### Execute Swap

```javascript
// Execute swap operation (signed message)
await client.doSwap('100', 'ETH', 'USDT');
```

#### Check Order Status

```javascript
// Get order status (signed message)
await client.getOrderStatus('ord_12345678');
```

### Event Handling

```javascript
// Authentication
client.on('authenticated', () => {
    console.log('✅ Authenticated successfully');
});

// Disconnection
client.on('disconnected', ({ code, reason }) => {
    console.log(`❌ Disconnected: ${code} ${reason}`);
});

// Errors
client.on('error', (error) => {
    console.error('❌ WebSocket error:', error.message);
});

// Subscriptions
client.on('subscribed', (channel) => {
    console.log(`✅ Subscribed to: ${channel}`);
});

client.on('unsubscribed', (channel) => {
    console.log(`✅ Unsubscribed from: ${channel}`);
});
```

### Error Handling

```javascript
try {
    await client.connect();
    await client.subscribe('balances', handleBalances);
} catch (error) {
    console.error('Connection failed:', error);
}

function handleBalances(message) {
    if (message.error) {
        console.error('Message error:', message.error);
        return;
    }
    
    // Process data
    console.log('Data:', message.data);
}
```

## API Reference

### BrokerWSClient

```javascript
class BrokerWSClient extends EventEmitter {
    constructor(apiKey, secretKey, wsURL)
}
```

#### Methods

##### constructor

```javascript
new BrokerWSClient(apiKey, secretKey, wsURL)
```

Creates a new WebSocket client.

**Parameters:**
- `apiKey` - API key
- `secretKey` - Secret key 
- `wsURL` - WebSocket server URL

##### connect()

```javascript
async connect()
```

Establishes connection and authenticates.

**Returns:** `Promise<void>`

##### subscribe()

```javascript
async subscribe(channel, handler)
```

Subscribes to a channel.

**Parameters:**
- `channel` - Channel name
- `handler` - Message handler function

**Returns:** `Promise<void>`

##### unsubscribe()

```javascript
async unsubscribe(channel)
```

Unsubscribes from a channel.

**Parameters:**
- `channel` - Channel name

**Returns:** `Promise<void>`

##### getBalances()

```javascript
async getBalances()
```

Requests current account balances (signed message).

**Returns:** `Promise<void>`

##### estimateSwap()

```javascript
async estimateSwap(amountIn, assetIn, assetOut)
```

Estimates swap cost (signed message).

**Parameters:**
- `amountIn` - Amount of input asset
- `assetIn` - Input asset symbol
- `assetOut` - Output asset symbol

**Returns:** `Promise<void>`

##### doSwap()

```javascript
async doSwap(amountIn, assetIn, assetOut)
```

Executes swap operation (signed message).

**Parameters:**
- `amountIn` - Amount of input asset
- `assetIn` - Input asset symbol
- `assetOut` - Output asset symbol

**Returns:** `Promise<void>`

##### getOrderStatus()

```javascript
async getOrderStatus(orderID)
```

Gets order status (signed message).

**Parameters:**
- `orderID` - Order ID to query

**Returns:** `Promise<void>`

##### close()

```javascript
close()
```

Closes WebSocket connection.

##### isConnected()

```javascript
isConnected()
```

Checks connection status.

**Returns:** `boolean`

#### Events

- `authenticated` - When client is authenticated
- `disconnected` - When connection is lost
- `error` - On error occurrence 
- `subscribed` - On successful channel subscription
- `unsubscribed` - On channel unsubscription

### Message Structure

```javascript
const message = {
    op: 'string',        // Operation (auth, subscribe, unsubscribe, estimate, swap, order_status, balances) 
    ch: 'string',        // Channel
    key: 'string',       // API key (auth only)
    ts: 1640995200000,   // Timestamp in milliseconds
    nonce: 'string',     // Unique nonce
    sig: 'string',       // HMAC signature
    data: {},            // Message data
    error: 'string'      // Error description
};
```

## Available Channels

### balances
Receives user balance updates.

**Example data:**
```javascript
{
  ch: 'balances',
  data: [
    {
      asset: 'USDT',
      total: '1000.00',
      locked: '0'
    }
  ]
}
```

### orders:{orderId}
Receives status updates for specific orders.

**Example data:**
```javascript
{
  ch: 'orders:ord_12345678',
  data: {
    status: 'FILLED',
    txHash: '0xabc123...',
    filledOut: '100.00',
    updatedAt: 1640995200000
  }
}
```

## Running the Example

```bash
# Run WebSocket example
node ws-example.js

# Or with ES modules
node --experimental-modules ws-example.js
```

## Features

### Automatic Reconnection

Client automatically reconnects on connection loss with exponential backoff:

```javascript
// Reconnection settings (defaults)
client.maxReconnectAttempts = 5;
client.reconnectDelay = 1000; // 1 second
```

### Asynchronous Processing

All methods return promises:

```javascript
try {
    await client.connect();
    await client.subscribe('balances', handler);
} catch (error) {
    console.error('Error:', error);
}
```

### Event-driven Architecture

Client extends EventEmitter for convenient event handling:

```javascript
client.on('error', (error) => {
    console.error('WebSocket error:', error);
});
```

## Security

- All trading operation messages are signed using HMAC-SHA256
- API keys should be stored securely
- Environment variables are recommended for keys

```javascript
const apiKey = process.env.BROKER_API_KEY;
const secretKey = process.env.BROKER_SECRET_KEY;
```

## Debugging

Enable debug logs for diagnostics:

```javascript
// Set logging level
client.logger.setLevel('debug');

// Or listen to all events
client.on('response', (message) => {
    console.log('Server response:', message);
});
```

## Integration with Existing Projects

### Express.js

```javascript
import express from 'express';
import { BrokerWSClient } from './ws-client.js';

const app = express();
const wsClient = new BrokerWSClient(apiKey, secretKey, wsURL);

app.get('/status', (req, res) => {
    res.json({
        websocket_connected: wsClient.isConnected()
    });
});

app.post('/swap', async (req, res) => {
    try {
        await wsClient.doSwap(req.body.amount, req.body.from, req.body.to);
        res.json({ success: true });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// On server start
wsClient.connect().then(() => {
    console.log('WebSocket client ready');
});
```

### As Middleware

```javascript
export function createWSMiddleware(apiKey, secretKey, wsURL) {
    const client = new BrokerWSClient(apiKey, secretKey, wsURL);
    
    return {
        async init() {
            await client.connect();
        },
        
        subscribe: client.subscribe.bind(client),
        unsubscribe: client.unsubscribe.bind(client),
        getBalances: client.getBalances.bind(client),
        estimateSwap: client.estimateSwap.bind(client),
        doSwap: client.doSwap.bind(client),
        getOrderStatus: client.getOrderStatus.bind(client),
        
        close() {
            client.close();
        }
    };
}
