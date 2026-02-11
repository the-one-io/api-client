# TheOne Trading API - JavaScript/Node.js Client Example

This example demonstrates how to use the TheOne Trading API with a modern JavaScript (Node.js) client supporting both REST and WebSocket connections.

## Features

- ✅ Complete REST API implementation
- ✅ WebSocket API with real-time updates
- ✅ Environment variables support (.env file)
- ✅ Modern ES Modules (import/export)
- ✅ Async/await Promise-based API
- ✅ HMAC-SHA256 authentication
- ✅ Idempotency support for swaps
- ✅ Auto-reconnection for WebSocket
- ✅ Cyclical order status checking
- ✅ Case-insensitive status matching
- ✅ Comprehensive error handling

## Requirements

- **Node.js 16+** (for ES Modules and built-in fetch support)
- **npm** or **yarn** package manager

## Installation

```bash
cd client-js
npm install
```

## Configuration

### Using Environment Variables (Recommended)

Create a `.env` file in the `client-js` directory:

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

### Using npm (Recommended)

```bash
# Install dependencies
npm install

# Run REST API client example
npm start

# Run WebSocket client example
npm run ws
```

### Using Makefile

```bash
# Quick start
make install && make run

# Run REST API client
make run

# Run WebSocket client
make run-ws

# See all available commands
make help
```

## REST API Methods

The client implements all main API methods:

| Method | Description | HTTP Endpoint |
|--------|-------------|---------------|
| `getBalances()` | Get user balances | GET /api/v1/balances |
| `estimateSwap(request)` | Get swap estimation | POST /api/v1/estimate |
| `swap(request, idempotencyKey)` | Execute swap | POST /api/v1/swap |
| `getOrderStatus(orderId, clientOrderId?)` | Get order status | GET /api/v1/orders/{id}/status |

## WebSocket API Methods

The WebSocket client supports real-time updates:

| Method | Description | Purpose |
|--------|-------------|---------|
| `connect()` | Establish WebSocket connection | Authentication and setup |
| `subscribe(channel, callback)` | Subscribe to data channels | Real-time updates |
| `unsubscribe(channel)` | Unsubscribe from channels | Stop receiving updates |
| `close()` | Close WebSocket connection | Cleanup resources |

### Available WebSocket Channels

- `balances` - Real-time balance updates
- `orders:{orderId}` - Real-time order status updates for specific order

## Usage Examples

### REST API

#### Setup Client

```javascript
import { BrokerClient } from './broker-client.js';

// Load API credentials (from .env or directly)
const apiKey = process.env.BROKER_API_KEY || "your_api_key";
const secretKey = process.env.BROKER_SECRET_KEY || "your_secret_key";
const baseURL = process.env.BROKER_BASE_URL || "https://partner-api-dev.the-one.io";

// Create client
const client = new BrokerClient(apiKey, secretKey, baseURL);
```

#### Get Balances

```javascript
try {
    const balances = await client.getBalances();
    console.log('Balances received:', JSON.stringify(balances, null, 2));
    
    // Process each balance
    balances.balances.forEach(balance => {
        console.log(`${balance.asset}: ${balance.total} (locked: ${balance.locked})`);
    });
} catch (error) {
    console.error('Error getting balances:', error.message);
}
```

#### Estimate Swap

```javascript
const estimateRequest = {
    from: "USDT",
    to: "BTC",
    amount: "10",
    filter: ["binance", "gate"] // Optional: specify exchanges
};

try {
    const estimate = await client.estimateSwap(estimateRequest);
    console.log('Estimate received:', JSON.stringify(estimate, null, 2));
    
    console.log(`Price: ${estimate.price}`);
    console.log(`Expected Out: ${estimate.expectedOut}`);
    console.log(`Expires At: ${new Date(estimate.expiresAt)}`);
    
    // Route information
    estimate.route.forEach((step, index) => {
        console.log(`Step ${index + 1}: ${step.exchange} - ${step.from_asset} -> ${step.to_asset}`);
    });
} catch (error) {
    console.error('Error getting estimate:', error.message);
}
```

#### Execute Swap

```javascript
const swapRequest = {
    from: "USDT",
    to: "BTC",
    amount: "10",
    slippage_bps: 30, // 0.3% slippage tolerance
    filter: ["binance", "gate"] // Optional
};

// Important: Use unique idempotency key for each swap
const idempotencyKey = `swap_${Date.now()}_${Math.random()}`;

try {
    const swapResponse = await client.swap(swapRequest, idempotencyKey);
    console.log('Swap created:', JSON.stringify(swapResponse, null, 2));
    
    console.log(`Order ID: ${swapResponse.orderId}`);
    console.log(`Status: ${swapResponse.status}`);
} catch (error) {
    console.error('Error executing swap:', error.message);
}
```

#### Check Order Status with Cyclical Polling

The example includes automatic cyclical order status checking (up to 5 attempts):

```javascript
// Example 4: Check order status cyclically
console.log("\n=== Checking Order Status ===");

const maxAttempts = 5;
let attempt = 0;
let orderStatus = null;
let finalStatus = null;

while (attempt < maxAttempts) {
    attempt++;
    console.log(`\nAttempt ${attempt}/${maxAttempts}:`);
    
    try {
        orderStatus = await client.getOrderStatus(swapResponse.orderId);
        console.log("Order status:", JSON.stringify(orderStatus, null, 2));
        
        // Get the status from the response
        finalStatus = orderStatus.status || orderStatus.order_status;
        
        // Check if status is filled or partialfilled (case-insensitive)
        const statusUpper = finalStatus.toUpperCase();
        if (statusUpper === 'FILLED' || statusUpper === 'PARTIAL_FILLED') {
            console.log(`\n✓ Order ${finalStatus}! Stopping status checks.`);
            break;
        }
        
        // If not final status and not last attempt, wait before next check
        if (attempt < maxAttempts) {
            console.log(`Status is '${finalStatus}', waiting 2 seconds before next check...`);
            await sleep(2000);
        }
    } catch (error) {
        console.log("Error getting order status:", error.message);
        
        // If not last attempt, wait before retry
        if (attempt < maxAttempts) {
            console.log("Waiting 2 seconds before retry...");
            await sleep(2000);
        }
    }
}

// Summary after all attempts
const finalStatusUpper = finalStatus ? finalStatus.toUpperCase() : '';
if (attempt >= maxAttempts && finalStatusUpper !== 'FILLED' && finalStatusUpper !== 'PARTIAL_FILLED') {
    console.log(`\n⚠ Maximum attempts (${maxAttempts}) reached. Final status: ${finalStatus || 'unknown'}`);
} else if (finalStatusUpper === 'FILLED' || finalStatusUpper === 'PARTIAL_FILLED') {
    console.log(`\n✓ Order successfully ${finalStatus} after ${attempt} attempt(s).`);
}
```

### WebSocket API

#### Setup WebSocket Client

```javascript
import { WSClient } from './ws-client.js';

const apiKey = process.env.BROKER_API_KEY;
const secretKey = process.env.BROKER_SECRET_KEY;
const baseURL = process.env.BROKER_BASE_URL;

// Convert HTTP URL to WebSocket URL
const wsURL = baseURL.replace('https://', 'wss://').replace('http://', 'ws://') + '/ws/v1/stream';

// Create WebSocket client
const wsClient = new WSClient(apiKey, secretKey, wsURL);

// Connect
await wsClient.connect();
console.log('Connected to WebSocket!');
```

#### Subscribe to Balances

```javascript
wsClient.subscribe('balances', (message) => {
    console.log('📊 Balance Update:', message);
    if (message.data) {
        console.log(JSON.stringify(message.data, null, 2));
    }
});
```

#### Subscribe to Order Updates

```javascript
const orderId = '12345';
const orderChannel = `orders:${orderId}`;

wsClient.subscribe(orderChannel, (message) => {
    console.log(`📦 Order Update for ${orderId}:`, message);
    if (message.data) {
        console.log(JSON.stringify(message.data, null, 2));
    }
});
```

#### Graceful Shutdown

```javascript
// Handle process termination
process.on('SIGINT', async () => {
    console.log('\nShutting down...');
    
    // Unsubscribe from channels
    wsClient.unsubscribe('balances');
    wsClient.unsubscribe(orderChannel);
    
    // Close connection
    await wsClient.close();
    
    process.exit(0);
});
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
- `NONCE` - unique value in format `{timestamp}_{random}`
- `BODY_SHA256_HEX` - SHA256 hash of request body in hex format (empty string for GET)

## Project Structure

```
client-js/
├── index.js          # REST API client example with cyclical status checking
├── broker-client.js  # REST API client class
├── ws-example.js     # WebSocket usage example
├── ws-client.js      # WebSocket client class
├── package.json      # Node.js project configuration
├── Makefile          # Build and run commands
├── .env              # Environment variables (create from .env.example)
├── .env.example      # Example environment configuration
└── README.md         # This file
```

## Error Handling

The client automatically handles API errors and HTTP errors:

```javascript
try {
    const result = await client.getBalances();
    console.log('Success:', result);
} catch (error) {
    if (error.code) {
        // API error from server
        console.error(`API Error [${error.code}]: ${error.message}`);
        console.error(`Request ID: ${error.requestId}`);
    } else {
        // Network or other error
        console.error(`Network Error: ${error.message}`);
    }
}
```

### Common Error Codes

- `INSUFFICIENT_BALANCE` - Not enough balance for swap
- `INVALID_AMOUNT` - Amount is invalid or too small
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `UNAUTHORIZED` - Invalid API credentials
- `ORDER_NOT_FOUND` - Order ID not found

## Order Statuses

The cyclical status checker supports case-insensitive matching:

- `PENDING` / `pending` - order is processing
- `FILLED` / `filled` - order is executed ✅ **Auto-stops checking**
- `PARTIAL_FILLED` / `partial_filled` - order is partially executed ✅ **Auto-stops checking**
- `CANCELED` / `canceled` - order is canceled
- `FAILED` / `failed` - order failed

## Package.json Scripts

```json
{
  "scripts": {
    "start": "node index.js",           // Run REST API example
    "ws": "node ws-example.js",         // Run WebSocket example
    "test": "echo \"No tests yet\""
  }
}
```

## Makefile Commands

The Makefile provides convenient commands:

```bash
make install       # Install dependencies
make run           # Run REST API client example
make run-ws        # Run WebSocket client example
make dev           # Run with auto-reload (if nodemon installed)
make clean         # Remove node_modules
make reinstall     # Clean and reinstall dependencies
make help          # Show all available commands
```

## Dependencies

- **Node.js built-in modules**: `crypto`, `https`
- **No external dependencies** for REST client (uses built-in fetch)
- **ws** - WebSocket library (for WebSocket client)

Install dependencies:
```bash
npm install
```

## ES Modules

This project uses ES Modules (`type: "module"` in package.json):

```javascript
// Use import/export instead of require/module.exports
import { BrokerClient } from './broker-client.js';
export class MyClass { }
```

Make sure to include `.js` extension in imports!

## Security Best Practices

1. **Never commit API keys to version control**
   - Use `.env` files (add to `.gitignore`)
   - Use environment variables in production

2. **Use HTTPS in production**
   - Always use `https://` URLs, not `http://`

3. **Validate input data**
   - Check amounts before executing swaps
   - Validate order IDs and other parameters

4. **Handle errors appropriately**
   - Check error types (API vs network errors)
   - Implement retry logic with exponential backoff
   - Log errors with request IDs for debugging

5. **Use unique idempotency keys**
   - Format: `swap_{timestamp}_{random}`
   - Never reuse keys for different swaps

## Advanced Features

### Cyclical Order Status Checking

The client includes built-in cyclical order status checking:

- **Maximum 5 attempts** by default
- **2-second delay** between attempts
- **Case-insensitive** status matching (handles both 'FILLED' and 'filled')
- **Auto-stops** when order is filled or partially filled
- **Detailed logging** of each attempt

### Custom Sleep Function

```javascript
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Usage
await sleep(2000); // Wait 2 seconds
```

## Troubleshooting

### Common Issues

1. **"Cannot find module" error**
   - Make sure to include `.js` extension in imports
   - Check that `package.json` has `"type": "module"`

2. **"unauthorized" error**
   - Check API keys in `.env` file
   - Verify signature generation
   - Check system clock synchronization

3. **"fetch is not defined" error (Node.js < 18)**
   - Upgrade to Node.js 18+ or install `node-fetch`
   - Or use `--experimental-fetch` flag

4. **WebSocket connection errors**
   - Check WebSocket URL format (wss:// or ws://)
   - Verify network connectivity
   - Check firewall settings

### Debug Mode

Enable debug logging:

```javascript
console.log('Debug info:', JSON.stringify(variable, null, 2));

// Or use a logger library
import winston from 'winston';
const logger = winston.createLogger({ /* config */ });
logger.info('Info message');
```

## Testing

### Manual Testing

```bash
# Run the example
npm start

# Check output for any errors
# Verify balances are retrieved
# Check swap execution
# Monitor order status updates
```

### Integration with Testing Frameworks

For integration with Jest or Mocha:

```javascript
import { BrokerClient } from './broker-client.js';
import { describe, it, expect } from '@jest/globals';

describe('BrokerClient', () => {
    it('should get balances', async () => {
        const client = new BrokerClient(apiKey, secretKey, baseURL);
        const balances = await client.getBalances();
        expect(balances).toHaveProperty('balances');
        expect(Array.isArray(balances.balances)).toBe(true);
    });
});
```

## Environment Variables

### Using dotenv (Optional)

If you want to use `.env` files:

```bash
npm install dotenv
```

```javascript
import dotenv from 'dotenv';
dotenv.config();

const apiKey = process.env.BROKER_API_KEY;
```

### Production Environment

Set environment variables in your deployment platform:

```bash
# Heroku
heroku config:set BROKER_API_KEY=your_key

# Docker
docker run -e BROKER_API_KEY=your_key ...

# PM2
pm2 start index.js --env production
```

## TypeScript Support

For TypeScript projects, create type definitions:

```typescript
// types.d.ts
export interface Balance {
    asset: string;
    total: string;
    locked: string;
}

export interface BalanceResponse {
    balances: Balance[];
}

export interface EstimateRequest {
    from: string;
    to: string;
    amount: string;
    network?: string;
    account?: string;
    filter?: string[];
}

export interface SwapRequest {
    from: string;
    to: string;
    amount: string;
    account?: string;
    slippage_bps: number;
    clientOrderId?: string;
    filter?: string[];
}

export interface OrderStatusResponse {
    orderId: string;
    status: string;
    filledOut?: string;
    txHash?: string;
    updatedAt: number;
    clientOrderId?: string;
}

export class BrokerClient {
    constructor(apiKey: string, secretKey: string, baseURL: string);
    getBalances(): Promise<BalanceResponse>;
    estimateSwap(request: EstimateRequest): Promise<any>;
    swap(request: SwapRequest, idempotencyKey: string): Promise<any>;
    getOrderStatus(orderId: string, clientOrderId?: string): Promise<OrderStatusResponse>;
}
```

## Support

- **Documentation**: See main [README.md](../README.md)
- **API Docs**: Check Swagger UI at `{baseURL}/swagger/`
- **WebSocket Docs**: See [WS_README.md](./WS_README.md)
- **Issues**: Report via GitLab issues

## License

MIT License - see [LICENSE](../LICENSE) for details

## Additional Resources

- [WebSocket README](./WS_README.md) - Detailed WebSocket documentation
- [Main README](../README.md) - Overview of all clients
- [API Documentation](https://partner-api-dev.the-one.io/swagger/) - Full API reference
- [Node.js Documentation](https://nodejs.org/docs/) - Node.js official docs

---

**Last Updated**: February 2026  
**Node.js Version**: 16+  
**Status**: Production Ready
