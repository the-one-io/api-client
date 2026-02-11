import { BrokerClient } from './broker-client.js';
import dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

// Sleep function for delay
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

async function main() {
    // Load API keys from environment variables
    const apiKey = process.env.BROKER_API_KEY;
    const secretKey = process.env.BROKER_SECRET_KEY;
    const baseURL = process.env.BROKER_BASE_URL;

    // Validate required environment variables
    if (!apiKey || !secretKey || !baseURL) {
        console.error('Error: BROKER_API_KEY, BROKER_SECRET_KEY, and BROKER_BASE_URL must be set in .env file or environment');
        process.exit(1);
    }

    // Create API client
    const client = new BrokerClient(apiKey, secretKey, baseURL);

    try {
        // Example 1: Get balances
        console.log("=== Getting Balances ===");
        try {
            const balances = await client.getBalances();
            console.log("Balances received:", JSON.stringify(balances, null, 2));
        } catch (error) {
            console.log("Error getting balances:", error.message);
        }

        // Example 2: Estimate swap
        console.log("\n=== Estimating Swap ===");
        const estimateRequest = {
            from: "USDT",
            to: "XRP",
            amount: "100"
        };

        let estimate = null;
        try {
            estimate = await client.estimateSwap(estimateRequest);
            console.log("Estimate received:", JSON.stringify(estimate, null, 2));
        } catch (error) {
            console.log("Error getting estimate:", error.message);
        }

        // Example 3: Execute swap (only if estimate exists)
        if (estimate) {
            console.log("\n=== Executing Swap ===");
            const swapRequest = {
                from: estimateRequest.from,
                to: estimateRequest.to,
                amount: estimateRequest.amount,
                slippage_bps: 30
            };

            const idempotencyKey = `swap_${Date.now()}_${Math.random()}`;
            
            try {
                const swapResponse = await client.swap(swapRequest, idempotencyKey);
                console.log("Swap created:", JSON.stringify(swapResponse, null, 2));

                // Example 4: Check order status
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
            } catch (error) {
                console.log("Error executing swap:", error.message);
            }
        }

    } catch (error) {
        console.error("General error:", error.message);
        
        // Additional information for API errors
        if (error.code) {
            console.error("API error code:", error.code);
            console.error("Request ID:", error.requestId);
        }
    }
}

// Run example
main().catch(console.error);
