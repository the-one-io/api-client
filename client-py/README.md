# Broker Trading API - Python Client Example

This example demonstrates how to use the Broker Trading API with a Python client.

## Installation

```bash
cd examples/client-py
pip install -r requirements.txt
```

Or using virtual environment (recommended):

```bash
cd examples/client-py
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
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

### Quick Start

```python
from broker_client import BrokerClient, APIError

# Client setup
api_key = "ak_NCkCBuPwz-76ZuIufZesX3CU0AEtZ_S3yAUDwJtBTsk"
secret_key = "tMLS4vFaKLApc8WL6qvn-3Gu11agkJe31ijrftHUMFYOjnXlUIkkc1sUzHqNSWjt"
base_url = "http://localhost:8080"

client = BrokerClient(api_key, secret_key, base_url)
```

### Getting Balances

```python
try:
    balances = client.get_balances()
    print(f"Found {len(balances)} assets:")
    for balance in balances:
        print(f"  {balance.asset}: {balance.total} (locked: {balance.locked})")
except APIError as e:
    print(f"API Error: {e}")
except Exception as e:
    print(f"Network Error: {e}")
```

### Swap Estimation

```python
try:
    estimate = client.estimate_swap(
        from_asset="ETH",
        to_asset="USDT", 
        amount="1.0",
        network="ETH",  # optional
        account="0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234"  # optional
    )
    
    print("Estimate received:")
    print(f"  Price: {estimate['price']}")
    print(f"  Expected Out: {estimate['expectedOut']}")
    print(f"  Expires At: {estimate['expiresAt']}")
    
    # Route information
    for i, step in enumerate(estimate['route']):
        print(f"  Step {i+1}: {step['exchange']} - {step['fromAsset']} -> {step['toAsset']}")
        
except APIError as e:
    print(f"API Error: {e}")
```

### Executing Swap

```python
import time

try:
    # Unique idempotency key
    idempotency_key = f"swap_{int(time.time() * 1000)}_{hash('unique_string') % 1000000}"
    
    swap_response = client.swap(
        from_asset="ETH",
        to_asset="USDT",
        amount="1.0",
        account="0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
        slippage_bps=30,
        idempotency_key=idempotency_key,
        client_order_id="my_order_123"  # optional
    )
    
    print("Swap created:")
    print(f"  Order ID: {swap_response['orderId']}")
    print(f"  Status: {swap_response['status']}")
    
except APIError as e:
    print(f"API Error: {e}")
```

### Checking Order Status

```python
try:
    order_status = client.get_order_status(
        order_id=swap_response['orderId'],
        client_order_id="my_order_123"  # optional
    )
    
    print("Order status:")
    print(f"  Order ID: {order_status['orderId']}")
    print(f"  Status: {order_status['status']}")
    print(f"  Filled Out: {order_status.get('filledOut', 'N/A')}")
    print(f"  TX Hash: {order_status.get('txHash', 'N/A')}")
    
except APIError as e:
    print(f"API Error: {e}")
```

## Running Example

```bash
python example.py
```

Or:

```bash
python3 example.py
```

## Project Structure

- `example.py` - main file with usage example
- `broker_client.py` - API client class with helper classes
- `requirements.txt` - Python dependencies
- `README.md` - this documentation

## Classes and Exceptions

### BrokerClient

Main client class with the following methods:

- `get_balances()` → `List[Balance]`
- `estimate_swap(from_asset, to_asset, amount, network=None, account=None)` → `Dict`
- `swap(from_asset, to_asset, amount, account, slippage_bps, idempotency_key, client_order_id=None)` → `Dict`
- `get_order_status(order_id, client_order_id=None)` → `Dict`

### Balance

Class for representing asset balance:

```python
class Balance:
    def __init__(self, asset: str, total: str, locked: str):
        self.asset = asset      # Asset name
        self.total = total      # Total balance
        self.locked = locked    # Locked balance
```

### APIError

Exception for API errors:

```python
class APIError(Exception):
    def __init__(self, code: str, message: str, request_id: str = ""):
        self.code = code           # Error code
        self.message = message     # Error message
        self.request_id = request_id  # Request ID
```

## Error Handling

The client automatically handles API and HTTP errors:

```python
try:
    result = client.get_balances()
    print("Success:", result)
except APIError as e:
    # Error from API server
    print(f"API Error [{e.code}]: {e.message}")
    if e.request_id:
        print(f"Request ID: {e.request_id}")
except requests.RequestException as e:
    # Network error
    print(f"Network Error: {e}")
except Exception as e:
    # Other errors
    print(f"Unexpected Error: {e}")
```

## Possible Order Statuses

- `PENDING` - order is processing
- `FILLED` - order is executed
- `PARTIAL` - order is partially executed
- `CANCELED` - order is canceled
- `FAILED` - order failed

## Python Implementation Features

### Type Hints
The code uses type hints for better readability and IDE support.

### Dependencies
- `requests` - for HTTP requests
- Built-in modules: `hashlib`, `hmac`, `json`, `time`, `random`, `base64`

### Compatibility
- Python 3.7+ (for type hints support)
- All major operating systems

## Security

- Never store API keys in code
- Use environment variables for production
- Ensure nonce uniqueness
- Use HTTPS in production

## Environment Variables (recommended)

Create a `.env` file or use environment variables:

```bash
export BROKER_API_KEY="your_api_key_here"
export BROKER_SECRET_KEY="your_secret_key_here"
export BROKER_BASE_URL="https://partner-api-dev.the-one.io"
```

And use in code:

```python
import os

api_key = os.getenv('BROKER_API_KEY')
secret_key = os.getenv('BROKER_SECRET_KEY')
base_url = os.getenv('BROKER_BASE_URL')

if not all([api_key, secret_key, base_url]):
    raise ValueError("Missing required environment variables")

client = BrokerClient(api_key, secret_key, base_url)
```

## Logging

For production, it's recommended to configure logging:

```python
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Add logging to client
try:
    result = client.get_balances()
    logger.info(f"Retrieved balances: {len(result)} assets")
except APIError as e:
    logger.error(f"API Error: {e}")
```

## Asynchronous Version (aiohttp)

For asynchronous usage, you can create a version with `aiohttp`:

```python
import aiohttp
import asyncio

class AsyncBrokerClient:
    async def get_balances(self):
        async with aiohttp.ClientSession() as session:
            # Implementation with aiohttp
            pass
```

## Testing

For testing, you can use mocking:

```python
import unittest
from unittest.mock import patch, MagicMock

class TestBrokerClient(unittest.TestCase):
    @patch('broker_client.requests.Session')
    def test_get_balances(self, mock_session):
        # Testing with mocking
        pass
