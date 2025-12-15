"""
WebSocket Client for Broker Trading API
"""

import asyncio
import json
import hashlib
import hmac
import time
import random
import base64
from typing import Optional, Callable, Dict, Any
import logging
import websockets
from websockets.exceptions import ConnectionClosed, WebSocketException


class BrokerWSClient:
    """
    WebSocket client for Broker Trading API
    
    Provides real-time data streaming with authentication and subscription management.
    """

    def __init__(self, api_key: str, secret_key: str, ws_url: str):
        """
        Initialize WebSocket client
        
        Args:
            api_key: API key for authentication
            secret_key: Secret key for signature generation
            ws_url: WebSocket server URL
        """
        self.api_key = api_key
        self.secret_key = secret_key
        self.ws_url = ws_url
        self.websocket = None
        self.authenticated = False
        self.subscriptions = {}
        self.reconnect_attempts = 0
        self.max_reconnect_attempts = 5
        self.reconnect_delay = 1.0
        self.running = False
        
        # Setup logging
        self.logger = logging.getLogger(__name__)
        if not self.logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
            handler.setFormatter(formatter)
            self.logger.addHandler(handler)
            self.logger.setLevel(logging.INFO)

    def generate_nonce(self) -> str:
        """Generate unique nonce"""
        timestamp_ns = int(time.time() * 1_000_000_000)
        random_part = random.randint(0, 999_999)
        return f"{timestamp_ns}_{random_part}"

    def generate_signature(self, timestamp: int, nonce: str) -> str:
        """
        Generate HMAC-SHA256 signature for WebSocket authentication
        
        Args:
            timestamp: Timestamp in milliseconds
            nonce: Unique nonce string
            
        Returns:
            HMAC-SHA256 signature in hex format
        """
        # For WebSocket auth, use the same signature format as REST API:
        # Method: "WS", Path: "/ws/v1/stream", Body: empty (SHA256 of empty bytes)
        method = "WS"
        path_with_query = "/ws/v1/stream"
        body_sha256 = hashlib.sha256(b'').hexdigest()
        
        canonical_string = "\n".join([method, path_with_query, str(timestamp), nonce, body_sha256])
        
        # Hash secret key and encode in base64url
        secret_key_hash = hashlib.sha256(self.secret_key.encode('utf-8')).digest()
        secret_key_base64 = base64.urlsafe_b64encode(secret_key_hash).decode('utf-8')
        
        # Create HMAC signature
        signature = hmac.new(
            secret_key_base64.encode('utf-8'),
            canonical_string.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        
        return signature

    async def connect(self):
        """Connect to WebSocket server and authenticate"""
        try:
            self.logger.info(f"Connecting to WebSocket: {self.ws_url}")
            self.websocket = await websockets.connect(self.ws_url)
            self.logger.info("WebSocket connected successfully")
            
            # Start message handler
            self.running = True
            asyncio.create_task(self._message_handler())
            
            # Authenticate
            await self._authenticate()
            self.logger.info("Authentication successful")
            self.reconnect_attempts = 0
            
        except Exception as e:
            self.logger.error(f"Failed to connect: {e}")
            raise

    async def _authenticate(self):
        """Send authentication message"""
        timestamp = int(time.time() * 1000)  # milliseconds
        nonce = self.generate_nonce()
        signature = self.generate_signature(timestamp, nonce)
        
        auth_message = {
            'op': 'auth',
            'key': self.api_key,
            'ts': timestamp,
            'nonce': nonce,
            'sig': signature
        }
        
        # Send authentication message
        await self._send_message(auth_message)
        
        # Wait for authentication response
        auth_timeout = 10.0
        start_time = time.time()
        
        while not self.authenticated and (time.time() - start_time) < auth_timeout:
            await asyncio.sleep(0.1)
            
        if not self.authenticated:
            raise Exception("Authentication timeout")

    async def subscribe(self, channel: str, handler: Callable[[Dict[str, Any]], None]):
        """
        Subscribe to a channel
        
        Args:
            channel: Channel name to subscribe to
            handler: Callback function to handle messages from this channel
        """
        if not self.authenticated:
            raise Exception("Not authenticated")
            
        # Store handler
        if channel not in self.subscriptions:
            self.subscriptions[channel] = []
        self.subscriptions[channel].append(handler)
        
        subscribe_message = {
            'op': 'subscribe',
            'ch': channel
        }
        
        await self._send_message(subscribe_message)
        self.logger.info(f"Subscription request sent for channel: {channel}")

    async def unsubscribe(self, channel: str):
        """
        Unsubscribe from a channel
        
        Args:
            channel: Channel name to unsubscribe from
        """
        if not self.authenticated:
            raise Exception("Not authenticated")
            
        # Remove handlers
        if channel in self.subscriptions:
            del self.subscriptions[channel]
        
        unsubscribe_message = {
            'op': 'unsubscribe',
            'ch': channel
        }
        
        await self._send_message(unsubscribe_message)
        self.logger.info(f"Unsubscribe request sent for channel: {channel}")

    async def _send_message(self, message: Dict[str, Any]):
        """Send message to WebSocket"""
        if not self.websocket:
            raise Exception("WebSocket not connected")
            
        message_json = json.dumps(message, separators=(',', ':'))
        await self.websocket.send(message_json)

    async def _message_handler(self):
        """Handle incoming WebSocket messages"""
        try:
            async for message in self.websocket:
                try:
                    data = json.loads(message)
                    await self._handle_message(data)
                except json.JSONDecodeError:
                    self.logger.error(f"Failed to parse message: {message}")
                except Exception as e:
                    self.logger.error(f"Error handling message: {e}")
                    
        except ConnectionClosed:
            self.logger.warning("WebSocket connection closed")
            self.authenticated = False
            if self.running:
                await self._handle_reconnect()
        except Exception as e:
            self.logger.error(f"Message handler error: {e}")
            self.authenticated = False
            if self.running:
                await self._handle_reconnect()

    async def _handle_message(self, data: Dict[str, Any]):
        """Process individual message"""
        # Handle error messages
        if 'error' in data and data['error']:
            self.logger.error(f"WebSocket message error: {data['error']}")
            return
            
        # Handle response messages (auth, subscribe, unsubscribe)
        if 'op' in data:
            op = data['op']
            self.logger.info(f"Received response for operation: {op}")
            
            if op == 'auth':
                if 'error' in data and data['error']:
                    self.logger.error(f"Authentication failed: {data['error']}")
                else:
                    self.authenticated = True
                    self.logger.info("Authentication successful")
                    
            elif op == 'subscribe':
                channel = data.get('ch', '')
                if 'error' in data and data['error']:
                    self.logger.error(f"Subscription failed for {channel}: {data['error']}")
                else:
                    self.logger.info(f"Successfully subscribed to channel: {channel}")
                    
            elif op == 'unsubscribe':
                channel = data.get('ch', '')
                if 'error' in data and data['error']:
                    self.logger.error(f"Unsubscribe failed for {channel}: {data['error']}")
                else:
                    self.logger.info(f"Successfully unsubscribed from channel: {channel}")
            return
            
        # Handle data messages
        if 'ch' in data:
            channel = data['ch']
            self.logger.debug(f"Received data for channel: {channel}")
            
            # Call all handlers for this channel
            if channel in self.subscriptions:
                for handler in self.subscriptions[channel]:
                    try:
                        if asyncio.iscoroutinefunction(handler):
                            await handler(data)
                        else:
                            handler(data)
                    except Exception as e:
                        self.logger.error(f"Error in message handler: {e}")

    async def _handle_reconnect(self):
        """Handle reconnection with exponential backoff"""
        if self.reconnect_attempts >= self.max_reconnect_attempts:
            self.logger.error("Max reconnection attempts reached")
            return
            
        self.reconnect_attempts += 1
        delay = self.reconnect_delay * (2 ** (self.reconnect_attempts - 1))
        
        self.logger.info(f"Attempting to reconnect in {delay}s (attempt {self.reconnect_attempts}/{self.max_reconnect_attempts})")
        
        await asyncio.sleep(delay)
        
        try:
            await self.connect()
            
            # Re-subscribe to all channels
            channels_to_resubscribe = list(self.subscriptions.keys())
            for channel in channels_to_resubscribe:
                handlers = self.subscriptions[channel].copy()
                # Clear and re-subscribe
                self.subscriptions[channel] = []
                for handler in handlers:
                    await self.subscribe(channel, handler)
                    
        except Exception as e:
            self.logger.error(f"Reconnection failed: {e}")
            await self._handle_reconnect()

    async def close(self):
        """Close WebSocket connection"""
        self.running = False
        self.authenticated = False
        self.subscriptions.clear()
        
        if self.websocket:
            self.logger.info("Closing WebSocket connection")
            await self.websocket.close()
            self.websocket = None

    def is_connected(self) -> bool:
        """Check if WebSocket is connected and authenticated"""
        return (
            self.websocket is not None and
            not self.websocket.closed and
            self.authenticated
        )

    async def __aenter__(self):
        """Async context manager entry"""
        await self.connect()
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        await self.close()
