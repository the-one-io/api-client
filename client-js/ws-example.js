import { BrokerWSClient } from './ws-client.js';
import dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

/**
 * Example usage of WebSocket client
 */
async function runWebSocketExample() {
    // Load API keys from environment variables
    const apiKey = process.env.BROKER_API_KEY;
    const secretKey = process.env.BROKER_SECRET_KEY;
    const baseURL = process.env.BROKER_BASE_URL;

    // Validate required environment variables
    if (!apiKey || !secretKey || !baseURL) {
        console.error('Error: BROKER_API_KEY, BROKER_SECRET_KEY, and BROKER_BASE_URL must be set in .env file or environment');
        process.exit(1);
    }

    // Convert HTTP URL to WebSocket URL
    const wsURL = baseURL.replace('https://', 'wss://').replace('http://', 'ws://') + '/ws/v1/stream';
    
    console.log(`Connecting to: ${wsURL}`);

    // Create WebSocket client
    const wsClient = new BrokerWSClient(apiKey, secretKey, wsURL);

    // Set up event listeners
    wsClient.on('authenticated', () => {
        console.log('✅ WebSocket client authenticated');
    });

    wsClient.on('disconnected', ({ code, reason }) => {
        console.log(`❌ WebSocket disconnected: ${code} ${reason}`);
    });

    wsClient.on('error', (error) => {
        console.error('❌ WebSocket error:', error.message);
    });

    wsClient.on('subscribed', (channel) => {
        console.log(`✅ Successfully subscribed to channel: ${channel}`);
    });

    wsClient.on('unsubscribed', (channel) => {
        console.log(`✅ Successfully unsubscribed from channel: ${channel}`);
    });

    // Store created order ID for subscription
    let createdOrderId = null;

    // Add response listener to handle operation results
    wsClient.on('response', async (message) => {
        if (message.op === 'estimate') {
            console.log('💰 Estimate response received:');
            console.log(JSON.stringify(message.data, null, 2));
        } else if (message.op === 'balances') {
            console.log('💼 Balances response received:');
            console.log(JSON.stringify(message.data, null, 2));
        } else if (message.op === 'swap') {
            console.log('🔄 Swap response received:');
            console.log(JSON.stringify(message.data, null, 2));
            
            // Subscribe to the created order channel for real-time updates
            if (message.data && message.data.orderId) {
                createdOrderId = message.data.orderId;
                const orderChannel = `orders:${createdOrderId}`;
                console.log(`\n📡 Subscribing to order channel: ${orderChannel}`);
                
                try {
                    await wsClient.subscribe(orderChannel, (orderMessage) => {
                        console.log(`\n🔔 Real-time order update for ${createdOrderId}:`);
                        console.log(JSON.stringify(orderMessage.data, null, 2));
                    });
                    
                    // Wait a bit and then check order status to see current state
                    setTimeout(async () => {
                        console.log(`\n📋 Checking order status for ${createdOrderId}...`);
                        try {
                            await wsClient.getOrderStatus(createdOrderId);
                        } catch (error) {
                            console.error('Failed to get order status:', error.message);
                        }
                    }, 1000);
                } catch (error) {
                    console.error('Failed to subscribe to order channel:', error.message);
                }
            }
        } else if (message.op === 'order_status') {
            console.log('📋 Order status response received:');
            console.log(JSON.stringify(message.data, null, 2));
        }
    });

    try {
        // Connect to WebSocket
        console.log('=== Connecting to WebSocket ===');
        await wsClient.connect();

        // Subscribe to balances channel
        console.log('\n=== Subscribing to balances channel ===');
        await wsClient.subscribe('balances', (message) => {
            console.log(`📊 Received balances update on channel '${message.ch}'`);
            if (message.data) {
                console.log('Data:', JSON.stringify(message.data, null, 2));
            }
        });

        // Subscribe to specific order channel
        console.log('\n=== Subscribing to order updates ===');
        const orderID = 'ord_12345678';
        const orderChannel = `orders:${orderID}`;
        
        await wsClient.subscribe(orderChannel, (message) => {
            console.log(`📦 Received order update for order '${orderID}'`);
            if (message.data) {
                console.log('Data:', JSON.stringify(message.data, null, 2));
            }
        });

        // Demo REST API commands via WebSocket
        console.log('\n=== Testing REST API commands via WebSocket ===');
        
        // Test estimate
        console.log('💰 Testing estimate swap...');
        await wsClient.estimateSwap('10.00', 'USDT', 'ETH');
        
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test balances request
        console.log('💼 Getting account balances...');
        await wsClient.getBalances();
        
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Test swap (creates a real swap order)
        console.log('🔄 Testing swap operation...');
        await wsClient.doSwap('10.00', 'USDT', 'TRX');
        
        // Wait for swap response to get order ID
        await new Promise(resolve => setTimeout(resolve, 3000));
        
        // Note: To test order status, uncomment below and use a real order ID from swap response
        // The order ID will be shown in the swap response above
        // console.log('📋 Getting order status...');
        // await wsClient.getOrderStatus('your_order_id_from_swap_response');
        // await new Promise(resolve => setTimeout(resolve, 1000));

        console.log('\n=== WebSocket client is running ===');
        console.log('Listening for real-time updates...');
        console.log('Press Ctrl+C to exit');

        // Handle graceful shutdown
        const handleShutdown = async () => {
            console.log('\n=== Shutting down ===');
            
            try {
                console.log('Unsubscribing from channels...');
                await wsClient.unsubscribe('balances');
                await wsClient.unsubscribe(orderChannel);
            } catch (error) {
                console.error('Error during unsubscribe:', error.message);
            }

            wsClient.close();
            console.log('WebSocket client stopped');
            process.exit(0);
        };

        // Listen for shutdown signals
        process.on('SIGINT', handleShutdown);
        process.on('SIGTERM', handleShutdown);

        // For demonstration, we'll also listen for Enter key to trigger manual actions
        process.stdin.setRawMode(true);
        process.stdin.resume();
        process.stdin.setEncoding('utf8');
        
        console.log('\nPress:');
        console.log('  b - to manually request balances subscription');
        console.log('  u - to test unsubscribe/resubscribe');
        console.log('  q - to quit');

        process.stdin.on('data', async (key) => {
            if (key === 'q' || key === '\u0003') { // 'q' or Ctrl+C
                await handleShutdown();
            } else if (key === 'b') {
                console.log('\n--- Manual balances subscription test ---');
                try {
                    await wsClient.subscribe('balances', (message) => {
                        console.log('🔄 Manual subscription - Balances update:', message.data);
                    });
                } catch (error) {
                    console.error('Manual subscription failed:', error.message);
                }
            } else if (key === 'u') {
                console.log('\n--- Testing unsubscribe/resubscribe ---');
                try {
                    await wsClient.unsubscribe('balances');
                    console.log('⏳ Waiting 3 seconds before resubscribing...');
                    
                    setTimeout(async () => {
                        try {
                            await wsClient.subscribe('balances', (message) => {
                                console.log('🔄 Resubscribed - Balances update:', message.data);
                            });
                        } catch (error) {
                            console.error('Resubscribe failed:', error.message);
                        }
                    }, 3000);
                } catch (error) {
                    console.error('Unsubscribe test failed:', error.message);
                }
            }
        });

    } catch (error) {
        console.error('❌ WebSocket example failed:', error.message);
        process.exit(1);
    }
}

// Run the example
runWebSocketExample().catch((error) => {
    console.error('❌ Failed to start WebSocket example:', error.message);
    process.exit(1);
});
