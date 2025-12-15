"""
Example usage of WebSocket client for Broker Trading API
"""

import asyncio
import json
import logging
import signal
from broker_ws_client import BrokerWSClient


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


async def main():
    """Main WebSocket client example"""
    setup_logging()
    
    # API keys (obtained from server)
    api_key = 'ak_X0sw1Dr97hcXkFsDM9nXbD2gn2ZkCIptjtpqQ-MvAnc'
    secret_key = 'L2pcViLs7uGdFZJd3wYKmnSgoSv-UYx8oF4c6lX95NSk3Ejm-T5eWproVlRcvQn1'
    ws_url = 'ws://localhost:8080/ws/v1/stream'

    # Create WebSocket client
    ws_client = BrokerWSClient(api_key, secret_key, ws_url)
    
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

            # Subscribe to specific order channel
            print("\n=== Subscribing to order updates ===")
            order_id = 'ord_12345678'
            order_channel = f'orders:{order_id}'
            await ws_client.subscribe(order_channel, order_handler)

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
