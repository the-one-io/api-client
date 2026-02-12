#!/usr/bin/env python3
"""
Example of using TheOne Trading API Python client
"""

import os
import time
from dotenv import load_dotenv
from broker_client import BrokerClient, APIError

# Load environment variables from .env file
load_dotenv()


def main():
    # Load API keys from environment variables
    api_key = os.getenv('BROKER_API_KEY')
    secret_key = os.getenv('BROKER_SECRET_KEY')
    base_url = os.getenv('BROKER_BASE_URL')

    # Validate required environment variables
    if not api_key or not secret_key or not base_url:
        print("Error: BROKER_API_KEY, BROKER_SECRET_KEY, and BROKER_BASE_URL must be set in .env file or environment")
        return

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
                to_asset="XRP",
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
                    to_asset="XRP",
                    amount="10",
                    account="test_account", # In real usage, need valid account
                    slippage_bps=30,
                    idempotency_key=idempotency_key
                )
                print("Swap created:")
                print(f"  Order ID: {swap_response.get('orderId', 'N/A')}")
                print(f"  Status: {swap_response.get('status', 'N/A')}")

                # Example 4: Check order status cyclically (up to 5 attempts)
                print("\n=== Checking Order Status ===")
                
                max_attempts = 5
                attempt = 0
                order_status = None
                final_status = None
                
                while attempt < max_attempts:
                    attempt += 1
                    print(f"\nAttempt {attempt}/{max_attempts}:")
                    
                    try:
                        order_status = client.get_order_status(swap_response['orderId'])
                        print("Order status:")
                        print(f"  Order ID: {order_status.get('orderId', 'N/A')}")
                        print(f"  Status: {order_status.get('status', 'N/A')}")
                        print(f"  Filled Out: {order_status.get('filledOut', 'N/A')}")
                        print(f"  TX Hash: {order_status.get('txHash', 'N/A')}")
                        print(f"  Updated At: {order_status.get('updatedAt', 'N/A')}")
                        
                        # Get the status from response
                        final_status = order_status.get('status', '')
                        
                        # Check if status is filled or partial_filled (case-insensitive)
                        status_upper = final_status.upper()
                        if status_upper == 'FILLED' or status_upper == 'PARTIAL_FILLED':
                            print(f"\n✓ Order {final_status}! Stopping status checks.")
                            break
                        
                        # If not final status and not last attempt, wait before next check
                        if attempt < max_attempts:
                            print(f"Status is '{final_status}', waiting 2 seconds before next check...")
                            time.sleep(2)
                            
                    except APIError as e:
                        print(f"Error getting order status: {e}")
                        
                        # If not last attempt, wait before retry
                        if attempt < max_attempts:
                            print("Waiting 2 seconds before retry...")
                            time.sleep(2)
                    except Exception as e:
                        print(f"Network error getting order status: {e}")
                        
                        # If not last attempt, wait before retry
                        if attempt < max_attempts:
                            print("Waiting 2 seconds before retry...")
                            time.sleep(2)
                
                # Summary after all attempts
                final_status_upper = final_status.upper() if final_status else ''
                if attempt >= max_attempts and final_status_upper not in ['FILLED', 'PARTIAL_FILLED']:
                    print(f"\n⚠ Maximum attempts ({max_attempts}) reached. Final status: {final_status or 'unknown'}")
                elif final_status_upper in ['FILLED', 'PARTIAL_FILLED']:
                    print(f"\n✓ Order successfully {final_status} after {attempt} attempt(s).")
                    
            except APIError as e:
                print(f"Error executing swap: {e}")
            except Exception as e:
                print(f"Network error executing swap: {e}")

    except Exception as e:
        print(f"General error: {e}")


if __name__ == "__main__":
    main()
