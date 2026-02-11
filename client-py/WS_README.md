# WebSocket Client for Python

WebSocket client for TheOne Trading API in Python with asynchronous processing, authentication and real-time support.

> 📖 **Related Documentation:**
> - [Python REST API Client](./README.md) - REST API documentation
> - [Main README](../README.md) - Overview of all clients
> - [Go WebSocket](../client-go/WS_README.md) - Go WebSocket documentation
> - [JavaScript WebSocket](../client-js/WS_README.md) - JavaScript WebSocket documentation

## Installation

```bash
pip install -r requirements.txt
```

## Dependencies

- `websockets` - Asynchronous WebSocket library for Python
- `requests` - For HTTP requests (already installed for REST client)

## Usage

### Simple Connection

```python
import asyncio
from broker_ws_client import BrokerWSClient

async def main():
    # API keys
    api_key = 'your_api_key'
    secret_key = 'your_secret_key'
    ws_url = 'wss://api.example.com/ws'
    
    # Create and connect client
    async with BrokerWSClient(api_key, secret_key, ws_url) as client:
        # Subscribe to balances
        await client.subscribe('balances', lambda msg: print(f"Balances: {msg['data']}"))
        
        # Trading operations
        await client.get_balances()
        await client.estimate_swap('100', 'ETH', 'USDT')
        await client.do_swap('100', 'ETH', 'USDT')
        await client.get_order_status('ord_12345678')
        
        # Wait for activity
        await asyncio.sleep(60)  # Listen for 1 minute

if __name__ == "__main__":
    asyncio.run(main())
```

### Channel Subscriptions

#### User Balances

```python
async def handle_balances(message):
    print(f"Received balances: {message['data']}")

await client.subscribe('balances', handle_balances)
```

#### Order Status

```python
order_id = 'ord_12345678'
channel = f'orders:{order_id}'

async def handle_order_update(message):
    print(f"Order {order_id} status: {message['data']}")

await client.subscribe(channel, handle_order_update)
```

### Trading Operations

#### Get Balances

```python
# Request current balances (signed message)
await client.get_balances()
```

#### Estimate Swap

```python
# Estimate swap cost (signed message)
await client.estimate_swap('100', 'ETH', 'USDT')
```

#### Execute Swap

```python
# Execute swap operation (signed message)
await client.do_swap('100', 'ETH', 'USDT')
```

#### Check Order Status

```python
# Get order status (signed message)
await client.get_order_status('ord_12345678')
```

### Error Handling

```python
async def handle_balances(message):
    if 'error' in message and message['error']:
        print(f"Error: {message['error']}")
        return
    
    # Process data
    print(f"Data: {message['data']}")

try:
    async with BrokerWSClient(api_key, secret_key, ws_url) as client:
        await client.subscribe('balances', handle_balances)
        await asyncio.sleep(60)
except Exception as e:
    print(f"Connection failed: {e}")
```

### Manual Connection Management

```python
client = BrokerWSClient(api_key, secret_key, ws_url)

try:
    await client.connect()
    await client.subscribe('balances', handle_balances)
    
    # Your code...
    
finally:
    await client.close()
```

## API Reference

### BrokerWSClient

```python
class BrokerWSClient:
    def __init__(self, api_key: str, secret_key: str, ws_url: str)
```

#### Methods

##### \_\_init\_\_

```python
def __init__(self, api_key: str, secret_key: str, ws_url: str)
```

Creates a new WebSocket client.

**Parameters:**
- `api_key` - API key
- `secret_key` - Secret key
- `ws_url` - WebSocket server URL

##### connect()

```python
async def connect()
```

Establishes connection and authenticates.

**Returns:** `None`  
**Raises:** `Exception` on connection error

##### subscribe()

```python
async def subscribe(channel: str, handler: Callable[[Dict[str, Any]], None])
```

Subscribes to a channel.

**Parameters:**
- `channel` - Channel name
- `handler` - Message handler function (can be sync or async)

**Returns:** `None`  
**Raises:** `Exception` if not authenticated

##### unsubscribe()

```python
async def unsubscribe(channel: str)
```

Unsubscribes from a channel.

**Parameters:**
- `channel` - Channel name

**Returns:** `None`

##### get_balances()

```python
async def get_balances()
```

Requests current account balances (signed message).

**Returns:** `None`  
**Raises:** `Exception` if not authenticated

##### estimate_swap()

```python
async def estimate_swap(amount_in: str, asset_in: str, asset_out: str)
```

Estimates swap cost (signed message).

**Parameters:**
- `amount_in` - Amount of input asset
- `asset_in` - Input asset symbol
- `asset_out` - Output asset symbol

**Returns:** `None`  
**Raises:** `Exception` if not authenticated

##### do_swap()

```python
async def do_swap(amount_in: str, asset_in: str, asset_out: str)
```

Executes swap operation (signed message).

**Parameters:**
- `amount_in` - Amount of input asset
- `asset_in` - Input asset symbol
- `asset_out` - Output asset symbol

**Returns:** `None`  
**Raises:** `Exception` if not authenticated

##### get_order_status()

```python
async def get_order_status(order_id: str)
```

Gets order status (signed message).

**Parameters:**
- `order_id` - Order ID to query

**Returns:** `None`  
**Raises:** `Exception` if not authenticated

##### close()

```python
async def close()
```

Closes WebSocket connection.

##### is_connected()

```python
def is_connected() -> bool
```

Checks connection status.

**Returns:** `bool` - True if connected and authenticated

#### Context Manager

Client supports async context manager:

```python
async with BrokerWSClient(api_key, secret_key, ws_url) as client:
    # Automatic connection
    await client.subscribe('balances', handler)
    # Automatic disconnection on exit
```

### Message Handlers

Handler functions can be synchronous or asynchronous:

```python
# Synchronous handler
def sync_handler(message):
    print("Sync handler:", message['data'])

# Asynchronous handler  
async def async_handler(message):
    await some_async_operation()
    print("Async handler:", message['data'])

await client.subscribe('balances', sync_handler)
await client.subscribe('orders:123', async_handler)
```

### Message Structure

```python
message = {
    'op': 'string',          # Operation (auth, subscribe, unsubscribe, estimate, swap, order_status, balances)
    'ch': 'string',          # Channel  
    'key': 'string',         # API key (auth only)
    'ts': 1640995200000,     # Timestamp in milliseconds
    'nonce': 'string',       # Unique nonce
    'sig': 'string',         # HMAC signature
    'data': {},              # Message data
    'error': 'string'        # Error description
}
```

## Available Channels

### balances
Receives user balance updates.

**Example data:**
```python
{
  'ch': 'balances',
  'data': [
    {
      'asset': 'USDT',
      'total': '1000.00', 
      'locked': '0'
    }
  ]
}
```

### orders:{orderId}
Receives status updates for specific orders.

**Example data:**
```python
{
  'ch': 'orders:ord_12345678',
  'data': {
    'status': 'FILLED',
    'txHash': '0xabc123...',
    'filledOut': '100.00',
    'updatedAt': 1640995200000
  }
}
```

## Running the Example

```bash
# Run WebSocket example
python ws_example.py

# With virtual environment
python -m venv venv
source venv/bin/activate  # or venv\Scripts\activate on Windows
pip install -r requirements.txt
python ws_example.py
```

## Features

### Automatic Reconnection

Client automatically reconnects on connection loss with exponential backoff:

```python
# Reconnection settings (defaults)
client.max_reconnect_attempts = 5
client.reconnect_delay = 1.0  # 1 second
```

All subscriptions are automatically restored on reconnection.

### Logging

Client uses standard `logging` library:

```python
import logging

# Enable debug logs
logging.basicConfig(level=logging.DEBUG)

# Or for specific client
client = BrokerWSClient(api_key, secret_key, ws_url)
client.logger.setLevel(logging.DEBUG)
```

### Asynchronous Architecture

Client is fully asynchronous and uses `asyncio`:

```python
# All operations are asynchronous
await client.connect()
await client.subscribe('balances', handler)
await client.unsubscribe('balances')
await client.close()

# Support for multiple clients
clients = [
    BrokerWSClient(key1, secret1, url),
    BrokerWSClient(key2, secret2, url)
]

await asyncio.gather(*[client.connect() for client in clients])
```

## Integration with Existing Projects

### FastAPI

```python
from fastapi import FastAPI
from broker_ws_client import BrokerWSClient

app = FastAPI()
ws_client = None

@app.on_event("startup")
async def startup():
    global ws_client
    ws_client = BrokerWSClient(api_key, secret_key, ws_url)
    await ws_client.connect()

@app.on_event("shutdown") 
async def shutdown():
    if ws_client:
        await ws_client.close()

@app.get("/status")
async def get_status():
    return {"websocket_connected": ws_client.is_connected() if ws_client else False}

@app.post("/swap")
async def execute_swap(amount: str, from_asset: str, to_asset: str):
    if not ws_client or not ws_client.is_connected():
        return {"error": "WebSocket not connected"}
    
    try:
        await ws_client.do_swap(amount, from_asset, to_asset)
        return {"success": True}
    except Exception as e:
        return {"error": str(e)}
```

### Django (with django-channels)

```python
from channels.generic.websocket import AsyncWebsocketConsumer
from broker_ws_client import BrokerWSClient

class TradingConsumer(AsyncWebsocketConsumer):
    async def connect(self):
        self.ws_client = BrokerWSClient(api_key, secret_key, ws_url)
        await self.ws_client.connect()
        await self.accept()
    
    async def disconnect(self, close_code):
        if hasattr(self, 'ws_client'):
            await self.ws_client.close()
```

### As a Service

```python
class WebSocketService:
    def __init__(self, api_key, secret_key, ws_url):
        self.client = BrokerWSClient(api_key, secret_key, ws_url)
        self.handlers = {}
    
    async def start(self):
        await self.client.connect()
    
    async def stop(self):
        await self.client.close()
    
    async def add_subscription(self, channel, callback):
        self.handlers[channel] = callback
        await self.client.subscribe(channel, callback)
    
    async def remove_subscription(self, channel):
        if channel in self.handlers:
            await self.client.unsubscribe(channel)
            del self.handlers[channel]
    
    async def get_balances(self):
        return await self.client.get_balances()
    
    async def estimate_swap(self, amount_in, asset_in, asset_out):
        return await self.client.estimate_swap(amount_in, asset_in, asset_out)
    
    async def execute_swap(self, amount_in, asset_in, asset_out):
        return await self.client.do_swap(amount_in, asset_in, asset_out)
```

## Security

- All trading operation messages are signed using HMAC-SHA256
- API keys should be stored securely
- Environment variables are recommended for keys

```python
import os

api_key = os.getenv('BROKER_API_KEY')
secret_key = os.getenv('BROKER_SECRET_KEY')

# Or with python-dotenv
from dotenv import load_dotenv
load_dotenv()

api_key = os.getenv('BROKER_API_KEY')
```

## Debugging

### Enable Debug Logs

```python
import logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
```

### Connection Monitoring

```python
async def monitor_connection():
    while True:
        print(f"Connected: {client.is_connected()}")
        await asyncio.sleep(10)

# Run in background
asyncio.create_task(monitor_connection())
```

### Handle All Messages

```python
async def debug_handler(message):
    print(f"DEBUG: Received message: {message}")

# Subscribe to all channels for debugging
await client.subscribe('balances', debug_handler)
await client.subscribe('orders:test_order', debug_handler)
```

## Requirements

- Python 3.7+
- `asyncio` support
- `websockets` library
- Stable internet connection

## Limitations

- Maximum 100 active subscriptions per connection
- Authentication timeout: 10 seconds  
- Maximum message size: 64KB
- Heartbeat interval: 30 seconds
