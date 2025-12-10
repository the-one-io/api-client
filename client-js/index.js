import { BrokerClient } from './broker-client.js';

// Sleep function for delay
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

async function main() {
    // API keys (obtained from server)
    const apiKey = "ak_WrXiA7I-VFolEYtZxnsqZTn-tB_f2zqSDEl4XQmqHqA";
    const secretKey = "NwTdHuVVfHA--40pyq_yqJBbscsbtPbD9jRhcU4tRFFQuYagqatzuhzrDu_-xd_q";
    const baseURL = "https://partner-api-dev.the-one.io";

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
            to: "BTC",
            amount: "10"
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
                await sleep(1000); // Wait 1 second
                console.log("\n=== Checking Order Status ===");
                
                try {
                    const orderStatus = await client.getOrderStatus(swapResponse.orderId);
                    console.log("Order status:", JSON.stringify(orderStatus, null, 2));
                } catch (error) {
                    console.log("Error getting order status:", error.message);
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
