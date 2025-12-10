#!/usr/bin/env python3
"""
Example of using Broker Trading API Python client
"""

import time
from broker_client import BrokerClient, APIError


def main():
    # API keys (obtained from server)
    api_key = "key"
    secret_key = "secret"
    base_url = "https://partner-api-dev.the-one.io"

    # Create API client
    client = BrokerClient(api_key, secret_key, base_url)

    try:
        # Example 1: Get balances
        print("=== Getting Balances ===")
        try:
            balances = client.get_balances()
            print(f"Balances received ({len(balances)} assets):")
            for balance in balances:
                print(f"  {balance}")
        except APIError as e:
            print(f"Error getting balances: {e}")
        except Exception as e:
            print(f"Network error getting balances: {e}")

        # Example 2: Estimate swap
        print("\n=== Estimating Swap ===")
        estimate = None
        try:
            estimate = client.estimate_swap(
                from_asset="USDT",
                to_asset="BTC", 
                amount="10"
            )
            print("Estimate received:")
            print(f"  Price: {estimate.get('price', 'N/A')}")
            print(f"  Expected Out: {estimate.get('expectedOut', 'N/A')}")
            print(f"  Expires At: {estimate.get('expiresAt', 'N/A')}")
            
            if 'route' in estimate and estimate['route']:
                print(f"  Route ({len(estimate['route'])} steps):")
                for i, step in enumerate(estimate['route']):
                    print(f"    {i+1}. {step['exchange']}: {step['fromAsset']} -> {step['toAsset']}")
                    
        except APIError as e:
            print(f"Error getting estimate: {e}")
        except Exception as e:
            print(f"Network error getting estimate: {e}")

        # Example 3: Execute swap (only if estimate exists)
        if estimate:
            print("\n=== Executing Swap ===")
            idempotency_key = f"swap_{int(time.time() * 1000)}_{int(time.time() % 1000000)}"
            
            try:
                swap_response = client.swap(
                    from_asset="USDT",
                    to_asset="BTC",
                    amount="10",
                    account="test_account", # In real usage, need valid account
                    slippage_bps=30,
                    idempotency_key=idempotency_key
                )
                print("Swap created:")
                print(f"  Order ID: {swap_response.get('orderId', 'N/A')}")
                print(f"  Status: {swap_response.get('status', 'N/A')}")

                # Example 4: Check order status
                time.sleep(1)  # Wait 1 second
                print("\n=== Checking Order Status ===")
                
                try:
                    order_status = client.get_order_status(swap_response['orderId'])
                    print("Order status:")
                    print(f"  Order ID: {order_status.get('orderId', 'N/A')}")
                    print(f"  Status: {order_status.get('status', 'N/A')}")
                    print(f"  Filled Out: {order_status.get('filledOut', 'N/A')}")
                    print(f"  TX Hash: {order_status.get('txHash', 'N/A')}")
                    print(f"  Updated At: {order_status.get('updatedAt', 'N/A')}")
                    
                except APIError as e:
                    print(f"Error getting order status: {e}")
                except Exception as e:
                    print(f"Network error getting order status: {e}")
                    
            except APIError as e:
                print(f"Error executing swap: {e}")
            except Exception as e:
                print(f"Network error executing swap: {e}")

    except Exception as e:
        print(f"General error: {e}")


if __name__ == "__main__":
    main()
