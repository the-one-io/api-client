"""
Example usage of WebSocket client for TheOne Trading API
"""

import asyncio
import json
import logging
import os
import signal
from dotenv import load_dotenv
from broker_ws_client import BrokerWSClient

# Load environment variables from .env file
load_dotenv()


def setup_logging():
    """Setup logging configuration"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )


async def balances_handler(message):
    """Handle balances updates"""
    print(f"📊 Received balances update on channel '{message['ch']}'")
    if 'data' in message:
        print("Data:", json.dumps(message['data'], indent=2))


async def order_handler(message):
    """Handle order updates"""
    print(f"📦 Received order update")
    if 'data' in message:
        print("Order Data:", json.dumps(message['data'], indent=2))


async def check_order_status_cyclically(ws_client, order_id, max_attempts=5):
    """
    Cyclically check order status until it's filled or max attempts reached
    
    Args:
        ws_client: WebSocket client instance
        order_id: Order ID to check
        max_attempts: Maximum number of attempts (default: 5)
    """
    attempt = 0
    
    while attempt < max_attempts:
        attempt += 1
        print(f"\nAttempt {attempt}/{max_attempts} - Checking status for order {order_id}...")
        
        try:
            await ws_client.get_order_status(order_id)
            
            # Wait 2 seconds before next check
            if attempt < max_attempts:
                await asyncio.sleep(2)
                
        except Exception as e:
            print(f"Error checking order status: {e}")
            if attempt < max_attempts:
                await asyncio.sleep(2)
    
    print(f"\n⚠ Completed {max_attempts} status checks for order {order_id}")


async def main():

    """Main WebSocket client example"""
    setup_logging()
    
    # Load API keys from environment variables
    api_key = os.getenv('BROKER_API_KEY')
    secret_key = os.getenv('BROKER_SECRET_KEY')
    base_url = os.getenv('BROKER_BASE_URL')

    # Validate required environment variables
    if not api_key or not secret_key or not base_url:
        print("Error: BROKER_API_KEY, BROKER_SECRET_KEY, and BROKER_BASE_URL must be set in .env file or environment")
        return

    # Convert HTTP URL to WebSocket URL
    ws_url = base_url.replace('https://', 'wss://').replace('http://', 'ws://') + '/ws/v1/stream'
    
    print(f"Connecting to: {ws_url}")

    # Create WebSocket client
    ws_client = BrokerWSClient(api_key, secret_key, ws_url)
    
    # Store created order ID for subscription
    created_order_id = None
    
    # Response handler for operation results
    async def response_handler(message):
        nonlocal created_order_id
        op = message.get('op', '')
        
        if op == 'estimate':
            print("💰 Estimate response received:")
            print(json.dumps(message.get('data', {}), indent=2))
        elif op == 'balances':
            print("💼 Balances response received:")
            print(json.dumps(message.get('data', {}), indent=2))
        elif op == 'swap':
            print("🔄 Swap response received:")
            print(json.dumps(message.get('data', {}), indent=2))
            
            # Subscribe to the created order channel for real-time updates
            if message.get('data') and message['data'].get('orderId'):
                created_order_id = message['data']['orderId']
                order_channel = f"orders:{created_order_id}"
                print(f"\n📡 Subscribing to order channel: {order_channel}")
                
                try:
                    async def order_update_handler(order_message):
                        print(f"\n🔔 Real-time order update for {created_order_id}:")
                        print(json.dumps(order_message.get('data', {}), indent=2))
                    
                    await ws_client.subscribe(order_channel, order_update_handler)
                    
                    # Cyclically check order status (up to 5 attempts)
                    print(f"\n📋 Checking order status for {created_order_id} (cyclical check)...")
                    asyncio.create_task(check_order_status_cyclically(ws_client, created_order_id))
                except Exception as e:
                    print(f"Failed to subscribe to order channel: {e}")
                    
        elif op == 'order_status':
            print("📋 Order status response received:")
            print(json.dumps(message.get('data', {}), indent=2))
    
    # Set response handler
    ws_client.response_handler = response_handler
    
    # Setup graceful shutdown
    shutdown_event = asyncio.Event()
    
    def signal_handler():
        print("\n=== Shutting down ===")
        shutdown_event.set()
    
    # Register signal handlers
    for sig in [signal.SIGINT, signal.SIGTERM]:
        signal.signal(sig, lambda s, f: signal_handler())

    try:
        # Connect to WebSocket using context manager
        print("=== Connecting to WebSocket ===")
        async with ws_client:
            # Subscribe to balances channel
            print("\n=== Subscribing to balances channel ===")
            await ws_client.subscribe('balances', balances_handler)

            # Subscribe to specific order channel (example)
            print("\n=== Subscribing to order updates (example) ===")
            order_id = 'ord_12345678'
            order_channel = f'orders:{order_id}'
            await ws_client.subscribe(order_channel, order_handler)

            # Demo REST API commands via WebSocket
            print("\n=== Testing REST API commands via WebSocket ===")
            
            # Test estimate
            print("💰 Testing estimate swap...")
            await ws_client.estimate_swap('10.00', 'USDT', 'ETH')
            await asyncio.sleep(1)
            
            # Test balances request
            print("💼 Getting account balances...")
            await ws_client.get_balances()
            await asyncio.sleep(2)
            
            # Test swap (creates a real swap order)
            print("🔄 Testing swap operation...")
            await ws_client.do_swap('10.00', 'USDT', 'TRX')
            await asyncio.sleep(3)

            print("\n=== WebSocket client is running ===")
            print("Listening for real-time updates...")
            print("Press Ctrl+C to exit")

            # Create input handler task for interactive commands
            input_task = asyncio.create_task(handle_user_input(ws_client))

            # Wait for shutdown signal or input task completion
            done, pending = await asyncio.wait(
                [input_task, asyncio.create_task(shutdown_event.wait())],
                return_when=asyncio.FIRST_COMPLETED
            )
            
            # Cancel pending tasks
            for task in pending:
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass

            # Cleanup subscriptions
            print("\nUnsubscribing from channels...")
            try:
                await ws_client.unsubscribe('balances')
                await ws_client.unsubscribe(order_channel)
            except Exception as e:
                print(f"Error during unsubscribe: {e}")

        print("WebSocket client stopped")

    except Exception as e:
        print(f"❌ WebSocket example failed: {e}")


async def handle_user_input(ws_client):
    """Handle user input for interactive commands"""
    print("\nPress:")
    print("  'b' + Enter - to manually test balances subscription")
    print("  'u' + Enter - to test unsubscribe/resubscribe")
    print("  'q' + Enter - to quit")
    
    loop = asyncio.get_event_loop()
    
    while True:
        try:
            # Read input from stdin in a non-blocking way
            user_input = await loop.run_in_executor(None, input, "\n> ")
            user_input = user_input.strip().lower()
            
            if user_input == 'q':
                break
            elif user_input == 'b':
                print("--- Manual balances subscription test ---")
                try:
                    def manual_balances_handler(message):
                        print("🔄 Manual subscription - Balances update:", message.get('data', {}))
                    
                    await ws_client.subscribe('balances', manual_balances_handler)
                except Exception as e:
                    print(f"Manual subscription failed: {e}")
                    
            elif user_input == 'u':
                print("--- Testing unsubscribe/resubscribe ---")
                try:
                    await ws_client.unsubscribe('balances')
                    print("⏳ Waiting 3 seconds before resubscribing...")
                    
                    await asyncio.sleep(3)
                    
                    def resubscribed_handler(message):
                        print("🔄 Resubscribed - Balances update:", message.get('data', {}))
                    
                    await ws_client.subscribe('balances', resubscribed_handler)
                except Exception as e:
                    print(f"Unsubscribe test failed: {e}")
            else:
                print("Unknown command. Use 'b', 'u', or 'q'.")
                
        except KeyboardInterrupt:
            break
        except EOFError:
            break
        except Exception as e:
            print(f"Input error: {e}")
            break


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nProgram interrupted")
    except Exception as e:
        print(f"❌ Failed to start WebSocket example: {e}")
