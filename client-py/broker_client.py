"""
Broker Trading API Client for Python
"""

import hashlib
import hmac
import json
import time
import random
import base64
from typing import Optional, Dict, Any, List
import requests


class APIError(Exception):
    """API Error"""
    def __init__(self, code: str, message: str, request_id: str = ""):
        super().__init__(f"API Error [{code}]: {message} (RequestID: {request_id})")
        self.code = code
        self.message = message
        self.request_id = request_id


class Balance:
    """Asset balance"""
    def __init__(self, asset: str, total: str, locked: str):
        self.asset = asset
        self.total = total
        self.locked = locked

    def __repr__(self):
        return f"Balance(asset='{self.asset}', total='{self.total}', locked='{self.locked}')"


class RouteStep:
    """Swap route step"""
    def __init__(self, exchange: str, pool: str, from_asset: str, to_asset: str, amount_in: str, amount_out: str):
        self.exchange = exchange
        self.pool = pool
        self.from_asset = from_asset
        self.to_asset = to_asset
        self.amount_in = amount_in
        self.amount_out = amount_out

    def __repr__(self):
        return f"RouteStep(exchange='{self.exchange}', from_asset='{self.from_asset}', to_asset='{self.to_asset}')"


class BrokerClient:
    """Client for Broker Trading API"""

    def __init__(self, api_key: str, secret_key: str, base_url: str):
        """
        Initialize API client

        Args:
            api_key: API key
            secret_key: Secret key
            base_url: Base API URL
        """
        self.api_key = api_key
        self.secret_key = secret_key
        self.base_url = base_url
        self.session = requests.Session()
        self.session.timeout = 30

    def generate_nonce(self) -> str:
        """Generate unique nonce"""
        timestamp_ns = int(time.time() * 1_000_000_000)
        random_part = random.randint(0, 999_999)
        return f"{timestamp_ns}_{random_part}"

    def hash_body(self, body: str) -> str:
        """Calculate SHA256 hash of request body"""
        if not body:
            body = ""
        return hashlib.sha256(body.encode('utf-8')).hexdigest()

    def generate_signature(self, method: str, path_with_query: str, timestamp: int, nonce: str, body_sha256: str) -> str:
        """
        Generate HMAC-SHA256 signature

        Args:
            method: HTTP method
            path_with_query: Path with query parameters
            timestamp: Timestamp in milliseconds
            nonce: Nonce
            body_sha256: SHA256 hash of request body

        Returns:
            HMAC-SHA256 signature in hex format
        """
        canonical_string = "\n".join([
            method.upper(),
            path_with_query,
            str(timestamp),
            nonce,
            body_sha256
        ])

        # Debug information
        print("Creating signature:")
        print(f"Method: {method.upper()}")
        print(f"Path: {path_with_query}")
        print(f"Timestamp: {timestamp}")
        print(f"Nonce: {nonce}")
        print(f"BodySHA256: {body_sha256}")
        print(f"CanonicalString: {repr(canonical_string)}")

        # Hash secret key and encode in base64url
        secret_key_hash = hashlib.sha256(self.secret_key.encode('utf-8')).digest()
        secret_key_base64 = base64.urlsafe_b64encode(secret_key_hash).decode('utf-8')
        
        # Create HMAC signature
        signature = hmac.new(
            secret_key_base64.encode('utf-8'),
            canonical_string.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()

        print(f"Signature: {signature}")
        print("---")

        return signature

    def make_request(self, method: str, path: str, body: Optional[Dict[str, Any]] = None, 
                     additional_headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        """
        Make authenticated request to API

        Args:
            method: HTTP method
            path: API path
            body: Request body
            additional_headers: Additional headers

        Returns:
            API response as dictionary

        Raises:
            APIError: API error
            requests.RequestException: Network error
        """
        timestamp = int(time.time() * 1000)  # milliseconds
        nonce = self.generate_nonce()
        
        body_string = json.dumps(body, separators=(',', ':')) if body else ""
        body_sha256 = self.hash_body(body_string)

        signature = self.generate_signature(method, path, timestamp, nonce, body_sha256)

        headers = {
            'Content-Type': 'application/json',
            'X-API-KEY': self.api_key,
            'X-API-TIMESTAMP': str(timestamp),
            'X-API-NONCE': nonce,
            'X-API-SIGN': signature,
        }

        if additional_headers:
            headers.update(additional_headers)

        url = self.base_url + path
        
        response = self.session.request(
            method=method,
            url=url,
            headers=headers,
            data=body_string if body_string else None
        )

        response.raise_for_status()

        # Parse response
        try:
            response_data = response.json()
        except json.JSONDecodeError:
            raise APIError("PARSE_ERROR", f"Failed to parse response: {response.text}")

        # Check if response is an API error
        if isinstance(response_data, dict) and 'code' in response_data and 'message' in response_data:
            raise APIError(
                code=response_data['code'],
                message=response_data['message'],
                request_id=response_data.get('requestId', '')
            )

        return response_data

    def get_balances(self) -> List[Balance]:
        """
        Get user balances

        Returns:
            List of asset balances
        """
        response = self.make_request('GET', '/api/v1/balances')
        balances_data = response.get('balances', [])
        return [Balance(
            asset=b['asset'],
            total=b['total'],
            locked=b['locked']
        ) for b in balances_data]

    def estimate_swap(self, from_asset: str, to_asset: str, amount: str, 
                      network: Optional[str] = None, account: Optional[str] = None) -> Dict[str, Any]:
        """
        Get swap estimation

        Args:
            from_asset: Source asset
            to_asset: Target asset
            amount: Amount
            network: Network (optional)
            account: Account (optional)

        Returns:
            Estimation data
        """
        request_data = {
            'from': from_asset,
            'to': to_asset,
            'amount': amount
        }
        
        if network:
            request_data['network'] = network
        if account:
            request_data['account'] = account

        return self.make_request('POST', '/api/v1/estimate', request_data)

    def swap(self, from_asset: str, to_asset: str, amount: str, account: str,
             slippage_bps: int, idempotency_key: str, client_order_id: Optional[str] = None) -> Dict[str, Any]:
        """
        Execute swap

        Args:
            from_asset: Source asset
            to_asset: Target asset
            amount: Amount
            account: Account
            slippage_bps: Slippage in basis points
            idempotency_key: Idempotency key
            client_order_id: Client order ID (optional)

        Returns:
            Created swap data
        """
        request_data = {
            'from': from_asset,
            'to': to_asset,
            'amount': amount,
            'account': account,
            'slippage_bps': slippage_bps
        }
        
        if client_order_id:
            request_data['clientOrderId'] = client_order_id

        headers = {'Idempotency-Key': idempotency_key}
        
        return self.make_request('POST', '/api/v1/swap', request_data, headers)

    def get_order_status(self, order_id: str, client_order_id: Optional[str] = None) -> Dict[str, Any]:
        """
        Get order status

        Args:
            order_id: Order ID
            client_order_id: Client order ID (optional)

        Returns:
            Order status
        """
        path = f'/api/v1/orders/{order_id}/status'
        if client_order_id:
            path += f'?clientOrderId={client_order_id}'

        return self.make_request('GET', path)
