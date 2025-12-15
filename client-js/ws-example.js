import { BrokerWSClient } from './ws-client.js';

/**
 * Example usage of WebSocket client
 */
async function runWebSocketExample() {
    // API keys (obtained from server)
    const apiKey = 'ak_X0sw1Dr97hcXkFsDM9nXbD2gn2ZkCIptjtpqQ-MvAnc';
    const secretKey = 'L2pcViLs7uGdFZJd3wYKmnSgoSv-UYx8oF4c6lX95NSk3Ejm-T5eWproVlRcvQn1';
    const wsURL = 'ws://localhost:8080/ws/v1/stream';

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
        await wsClient.estimateSwap('100.00', 'USDT', 'ETH');
        
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test balances request
        console.log('💼 Getting account balances...');
        await wsClient.getBalances();
        
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test order status
        console.log('📋 Getting order status...');
        await wsClient.getOrderStatus('ord_12345678');
        
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test swap (creates new order)
        console.log('🔄 Testing swap operation...');
        await wsClient.doSwap('100.00', 'USDT', 'ETH');

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
