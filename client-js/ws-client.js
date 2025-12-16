import crypto from 'crypto';
import { WebSocket } from 'ws';
import { EventEmitter } from 'events';

/**
 * WebSocket Client for Broker Trading API
 */
export class BrokerWSClient extends EventEmitter {
    /**
     * @param {string} apiKey - API key
     * @param {string} secretKey - Secret key
     * @param {string} wsURL - WebSocket URL
     */
    constructor(apiKey, secretKey, wsURL) {
        super();
        this.apiKey = apiKey;
        this.secretKey = secretKey;
        this.wsURL = wsURL;
        this.ws = null;
        this.authenticated = false;
        this.subscriptions = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
    }

    /**
     * Generates unique nonce
     * @returns {string}
     */
    generateNonce() {
        const timestamp = Date.now() * 1000000; // nanoseconds
        const random = Math.floor(Math.random() * 1000000);
        return `${timestamp}_${random}`;
    }

    /**
     * Generates HMAC-SHA256 signature for WebSocket authentication
     * @param {number} timestamp - Timestamp in milliseconds
     * @param {string} nonce - Nonce
     * @param {string} operation - Operation name (optional, for signed messages)
     * @param {object} data - Message data (optional, for signed messages)
     * @returns {string}
     */
    generateSignature(timestamp, nonce, operation = null, data = null) {
        // For WebSocket auth, use the same signature format as REST API:
        // Method: "WS", Path: "/ws/v1/stream" or "/ws/v1/{operation}", Body: SHA256 of data
        const method = 'WS';
        const pathWithQuery = operation ? `/ws/v1/${operation}` : '/ws/v1/stream';
        
        let bodyBytes = Buffer.from('');
        if (data !== null) {
            bodyBytes = Buffer.from(JSON.stringify(data));
        }
        const bodySHA256 = crypto.createHash('sha256').update(bodyBytes).digest('hex');
        
        const canonicalString = [method, pathWithQuery, timestamp.toString(), nonce, bodySHA256].join('\n');

        // Hash the secret key with SHA256
        const secretKeyHash = crypto.createHash('sha256').update(this.secretKey).digest();
        // Encode as base64url to match Go implementation
        const secretKeyBase64 = Buffer.from(secretKeyHash).toString('base64')
            .replace(/\+/g, '-')
            .replace(/\//g, '_');
        
        const hmacKey = Buffer.from(secretKeyBase64, 'utf8');
        
        const signature = crypto
            .createHmac('sha256', hmacKey)
            .update(canonicalString)
            .digest('hex');

        return signature;
    }

    /**
     * Creates a signed message for operations that require authentication
     * @param {string} operation - Operation name
     * @param {object} data - Message data
     * @returns {object}
     */
    createSignedMessage(operation, data) {
        const timestamp = Date.now();
        const nonce = this.generateNonce();
        const signature = this.generateSignature(timestamp, nonce, operation, data);

        return {
            op: operation,
            ts: timestamp,
            nonce: nonce,
            sig: signature,
            data: data
        };
    }

    /**
     * Connects to WebSocket and authenticates
     * @returns {Promise<void>}
     */
    async connect() {
        return new Promise((resolve, reject) => {
            console.log(`Connecting to WebSocket: ${this.wsURL}`);
            
            this.ws = new WebSocket(this.wsURL);

            this.ws.on('open', async () => {
                console.log('WebSocket connected');
                try {
                    await this.authenticate();
                    resolve();
                } catch (error) {
                    reject(error);
                }
            });

            this.ws.on('message', (data) => {
                this.handleMessage(data);
            });

            this.ws.on('close', (code, reason) => {
                console.log(`WebSocket closed: ${code} ${reason}`);
                this.authenticated = false;
                this.emit('disconnected', { code, reason });
                this.handleReconnect();
            });

            this.ws.on('error', (error) => {
                console.error('WebSocket error:', error);
                this.emit('error', error);
                reject(error);
            });
        });
    }

    /**
     * Authenticates with the WebSocket server
     * @returns {Promise<void>}
     */
    async authenticate() {
        return new Promise((resolve, reject) => {
            const timestamp = Date.now();
            const nonce = this.generateNonce();
            const signature = this.generateSignature(timestamp, nonce);

            const authMessage = {
                op: 'auth',
                key: this.apiKey,
                ts: timestamp,
                nonce: nonce,
                sig: signature
            };

            // Set up one-time listener for auth response
            const authTimeout = setTimeout(() => {
                reject(new Error('Authentication timeout'));
            }, 10000);

            const authHandler = (msg) => {
                if (msg.op === 'auth') {
                    clearTimeout(authTimeout);
                    if (msg.error) {
                        reject(new Error(`Authentication failed: ${msg.error}`));
                    } else {
                        console.log('Authentication successful');
                        this.authenticated = true;
                        this.reconnectAttempts = 0;
                        this.emit('authenticated');
                        resolve();
                    }
                }
            };

            this.once('response', authHandler);
            this.send(authMessage);
        });
    }

    /**
     * Subscribes to a channel
     * @param {string} channel - Channel name
     * @param {function} handler - Message handler function
     * @returns {Promise<void>}
     */
    async subscribe(channel, handler) {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        return new Promise((resolve, reject) => {
            // Store handler
            if (!this.subscriptions.has(channel)) {
                this.subscriptions.set(channel, []);
            }
            this.subscriptions.get(channel).push(handler);

            const subscribeMessage = {
                op: 'subscribe',
                ch: channel
            };

            const subTimeout = setTimeout(() => {
                reject(new Error(`Subscription timeout for channel: ${channel}`));
            }, 5000);

            const subHandler = (msg) => {
                if (msg.op === 'subscribe' && msg.ch === channel) {
                    clearTimeout(subTimeout);
                    if (msg.error) {
                        reject(new Error(`Subscription failed: ${msg.error}`));
                    } else {
                        console.log(`Subscribed to channel: ${channel}`);
                        this.emit('subscribed', channel);
                        resolve();
                    }
                }
            };

            this.once('response', subHandler);
            this.send(subscribeMessage);
        });
    }

    /**
     * Unsubscribes from a channel
     * @param {string} channel - Channel name
     * @returns {Promise<void>}
     */
    async unsubscribe(channel) {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        return new Promise((resolve, reject) => {
            // Remove handlers
            this.subscriptions.delete(channel);

            const unsubscribeMessage = {
                op: 'unsubscribe',
                ch: channel
            };

            const unsubTimeout = setTimeout(() => {
                reject(new Error(`Unsubscribe timeout for channel: ${channel}`));
            }, 5000);

            const unsubHandler = (msg) => {
                if (msg.op === 'unsubscribe' && msg.ch === channel) {
                    clearTimeout(unsubTimeout);
                    if (msg.error) {
                        reject(new Error(`Unsubscribe failed: ${msg.error}`));
                    } else {
                        console.log(`Unsubscribed from channel: ${channel}`);
                        this.emit('unsubscribed', channel);
                        resolve();
                    }
                }
            };

            this.once('response', unsubHandler);
            this.send(unsubscribeMessage);
        });
    }

    /**
     * Sends a message to WebSocket
     * @param {object} message - Message to send
     */
    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            throw new Error('WebSocket is not open');
        }
    }

    /**
     * Handles incoming WebSocket messages
     * @param {Buffer} data - Raw message data
     */
    handleMessage(data) {
        try {
            const message = JSON.parse(data.toString());

            // Handle error messages
            if (message.error) {
                console.error('WebSocket message error:', message.error);
                this.emit('error', new Error(message.error));
                return;
            }

            // Handle response messages (auth, subscribe, unsubscribe)
            if (message.op) {
                console.log(`Received response for operation: ${message.op}`);
                this.emit('response', message);
                return;
            }

            // Handle data messages
            if (message.ch) {
                console.log(`Received data for channel: ${message.ch}`);
                const handlers = this.subscriptions.get(message.ch) || [];
                handlers.forEach(handler => {
                    try {
                        handler(message);
                    } catch (error) {
                        console.error('Error in message handler:', error);
                        this.emit('error', error);
                    }
                });
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
            this.emit('error', error);
        }
    }

    /**
     * Handles reconnection logic
     */
    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(async () => {
                try {
                    await this.connect();
                    // Re-subscribe to all channels
                    const channelsToResubscribe = Array.from(this.subscriptions.keys());
                    for (const channel of channelsToResubscribe) {
                        const handlers = this.subscriptions.get(channel);
                        // Clear handlers temporarily to avoid duplicates
                        this.subscriptions.delete(channel);
                        // Re-subscribe with first handler (others will be added back)
                        if (handlers && handlers.length > 0) {
                            await this.subscribe(channel, handlers[0]);
                            // Add back remaining handlers
                            for (let i = 1; i < handlers.length; i++) {
                                this.subscriptions.get(channel).push(handlers[i]);
                            }
                        }
                    }
                } catch (error) {
                    console.error('Reconnection failed:', error);
                    this.emit('error', error);
                }
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
            this.emit('error', new Error('Max reconnection attempts reached'));
        }
    }

    /**
     * Closes WebSocket connection
     */
    close() {
        if (this.ws) {
            console.log('Closing WebSocket connection');
            this.ws.close();
            this.ws = null;
        }
        this.authenticated = false;
        this.subscriptions.clear();
    }

    /**
     * Checks if WebSocket is connected and authenticated
     * @returns {boolean}
     */
    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN && this.authenticated;
    }

    /**
     * Estimates a swap operation via WebSocket
     * @param {string} amountIn - Amount of input asset
     * @param {string} assetIn - Input asset
     * @param {string} assetOut - Output asset
     * @returns {Promise<void>}
     */
    async estimateSwap(amountIn, assetIn, assetOut) {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        const estimateData = {
            amountIn: amountIn,
            assetIn: assetIn,
            assetOut: assetOut
        };

        const estimateMessage = this.createSignedMessage('estimate', estimateData);
        this.send(estimateMessage);
    }

    /**
     * Executes a swap operation via WebSocket
     * @param {string} amountIn - Amount of input asset
     * @param {string} assetIn - Input asset
     * @param {string} assetOut - Output asset
     * @returns {Promise<void>}
     */
    async doSwap(amountIn, assetIn, assetOut) {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        const swapData = {
            amountIn: amountIn,
            assetIn: assetIn,
            assetOut: assetOut
        };

        const swapMessage = this.createSignedMessage('swap', swapData);
        this.send(swapMessage);
    }

    /**
     * Gets order status via WebSocket
     * @param {string} orderID - Order ID
     * @returns {Promise<void>}
     */
    async getOrderStatus(orderID) {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        const orderData = {
            id: orderID
        };

        const orderMessage = this.createSignedMessage('order_status', orderData);
        this.send(orderMessage);
    }

    /**
     * Gets account balances via WebSocket
     * @returns {Promise<void>}
     */
    async getBalances() {
        if (!this.authenticated) {
            throw new Error('Not authenticated');
        }

        const balancesMessage = this.createSignedMessage('balances', null);
        this.send(balancesMessage);
    }
}
