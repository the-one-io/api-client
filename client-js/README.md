# Broker Trading API - JavaScript Client Example

This example demonstrates how to use the Broker Trading API with a JavaScript (Node.js) client.

## Installation

```bash
cd examples/client-js
npm install
```

## Description

The client implements all main API methods:
- **GET /api/v1/balances** - get user balances
- **POST /api/v1/estimate** - get swap estimation
- **POST /api/v1/swap** - execute swap (with idempotency support)
- **GET /api/v1/orders/{id}/status** - get order status

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

## Usage

### Setup

```javascript
import { BrokerClient } from './broker-client.js';

const apiKey = "ak_NCkCBuPwz-76ZuIufZesX3CU0AEtZ_S3yAUDwJtBTsk";
const secretKey = "tMLS4vFaKLApc8WL6qvn-3Gu11agkJe31ijrftHUMFYOjnXlUIkkc1sUzHqNSWjt";
const baseURL = "http://localhost:8080";

const client = new BrokerClient(apiKey, secretKey, baseURL);
```

### Getting Balances

```javascript
try {
    const balances = await client.getBalances();
    console.log('Balances:', balances);
} catch (error) {
    console.error('Error:', error.message);
}
```

### Swap Estimation

```javascript
const estimateRequest = {
    from: "ETH",
    to: "USDT", 
    amount: "1.0",
    network: "ETH",
    account: "0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234" // optional
};

try {
    const estimate = await client.estimateSwap(estimateRequest);
    console.log('Estimate:', estimate);
} catch (error) {
    console.error('Error:', error.message);
}
```

### Executing Swap

```javascript
const swapRequest = {
    from: "ETH",
    to: "USDT",
    amount: "1.0",
    account: "0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
    slippage_bps: 30,
    clientOrderId: "my_order_123" // optional
};

// Important: use unique idempotency key for each swap
const idempotencyKey = `swap_${Date.now()}_${Math.random()}`;

try {
    const swapResponse = await client.swap(swapRequest, idempotencyKey);
    console.log('Swap created:', swapResponse);
} catch (error) {
    console.error('Error:', error.message);
}
```

### Checking Order Status

```javascript
try {
    const orderStatus = await client.getOrderStatus(swapResponse.orderId);
    console.log('Order status:', orderStatus);
} catch (error) {
    console.error('Error:', error.message);
}

// Or with clientOrderId
try {
    const orderStatus = await client.getOrderStatus(swapResponse.orderId, "my_order_123");
    console.log('Order status:', orderStatus);
} catch (error) {
    console.error('Error:', error.message);
}
```

## Running Example

```bash
npm start
```

## Project Structure

- `index.js` - main file with usage example
- `broker-client.js` - API client class
- `package.json` - Node.js project configuration

## Error Handling

The client automatically handles API errors and HTTP errors:

```javascript
try {
    const result = await client.getBalances();
    console.log(result);
} catch (error) {
    if (error.code) {
        // API error
        console.error(`API Error [${error.code}]: ${error.message}`);
        console.error(`Request ID: ${error.requestId}`);
    } else {
        // Network or other error
        console.error(`Network/Other Error: ${error.message}`);
    }
}
```

## Possible Order Statuses

- `PENDING` - order is processing
- `FILLED` - order is executed
- `PARTIAL` - order is partially executed
- `CANCELED` - order is canceled
- `FAILED` - order failed

## JavaScript Implementation Features

### Asynchronous
All client methods return Promises and should be used with `async/await` or `.then()/.catch()`.

### Dependencies
- `node-fetch` - for HTTP requests
- `crypto` - built-in Node.js module for cryptographic functions

### ES Modules
The project uses ES Modules (import/export). Make sure your package.json specifies `"type": "module"`.

## Security

- Never store API keys in code
- Use environment variables for production
- Ensure nonce uniqueness
- Use HTTPS in production

## Environment Variables (recommended)

Create a `.env` file:
```
BROKER_API_KEY=your_api_key_here
BROKER_SECRET_KEY=your_secret_key_here
BROKER_BASE_URL=https://partner-api-dev.the-one.io
```

And use in code:
```javascript
import dotenv from 'dotenv';
dotenv.config();

const apiKey = process.env.BROKER_API_KEY;
const secretKey = process.env.BROKER_SECRET_KEY;
const baseURL = process.env.BROKER_BASE_URL;
```

## Compatibility

- Node.js 16+ (for ES Modules and fetch API support)
- Modern browsers (with appropriate polyfills)

## TypeScript

For use with TypeScript, create a types file `types.d.ts`:

```typescript
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
}

// ... other types
