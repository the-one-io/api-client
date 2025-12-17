import crypto from 'crypto';
import fetch from 'node-fetch';

/**
 * TheOne Trading API Client for JavaScript
 */
export class BrokerClient {
    /**
     * @param {string} apiKey - API key
     * @param {string} secretKey - Secret key
     * @param {string} baseURL - Base API URL
     */
    constructor(apiKey, secretKey, baseURL) {
        this.apiKey = apiKey;
        this.secretKey = secretKey;
        this.baseURL = baseURL;
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
     * Calculates SHA256 hash of request body
     * @param {string|null} body - Request body
     * @returns {string}
     */
    hashBody(body) {
        if (!body) {
            body = '';
        }
        return crypto.createHash('sha256').update(body).digest('hex');
    }

    /**
     * Generates HMAC-SHA256 signature
     * @param {string} method - HTTP method
     * @param {string} pathWithQuery - Path with query parameters
     * @param {number} timestamp - Timestamp in milliseconds
     * @param {string} nonce - Nonce
     * @param {string} bodySHA256 - SHA256 hash of request body
     * @returns {string}
     */
    generateSignature(method, pathWithQuery, timestamp, nonce, bodySHA256) {
        const canonicalString = [
            method.toUpperCase(),
            pathWithQuery,
            timestamp.toString(),
            nonce,
            bodySHA256
        ].join('\n');

        // Hash the secret key with SHA256
        const secretKeyHash = crypto.createHash('sha256').update(this.secretKey).digest();
        // Encode as base64 (not base64url) to match Go implementation that includes padding
        const secretKeyBase64 = Buffer.from(secretKeyHash).toString('base64')
            .replace(/\+/g, '-')
            .replace(/\//g, '_');
        // In Go: []byte(base64.URLEncoding.EncodeToString(hash[:])) - this means bytes of the string, not decode
        const hmacKey = Buffer.from(secretKeyBase64, 'utf8');
        
        const signature = crypto
            .createHmac('sha256', hmacKey)
            .update(canonicalString)
            .digest('hex');

        return signature;
    }

    /**
     * Makes authenticated request to API
     * @param {string} method - HTTP method
     * @param {string} path - API path
     * @param {object|null} body - Request body
     * @param {object} additionalHeaders - Additional headers
     * @returns {Promise<object>}
     */
    async makeRequest(method, path, body = null, additionalHeaders = {}) {
        const timestamp = Date.now();
        const nonce = this.generateNonce();
        
        const bodyString = body ? JSON.stringify(body) : '';
        const bodySHA256 = this.hashBody(bodyString);

        const signature = this.generateSignature(method, path, timestamp, nonce, bodySHA256);

        const headers = {
            'Content-Type': 'application/json',
            'X-API-KEY': this.apiKey,
            'X-API-TIMESTAMP': timestamp.toString(),
            'X-API-NONCE': nonce,
            'X-API-SIGN': signature,
            ...additionalHeaders
        };

        const url = this.baseURL + path;
        const options = {
            method,
            headers,
            body: bodyString || undefined
        };

        const response = await fetch(url, options);
        const responseText = await response.text();

        if (!response.ok) {
            throw new Error(`HTTP error ${response.status}: ${responseText}`);
        }

        // Check if response is an API error
        let responseData;
        try {
            responseData = JSON.parse(responseText);
            if (responseData.code && responseData.message) {
                const error = new Error(`API Error [${responseData.code}]: ${responseData.message}`);
                error.code = responseData.code;
                error.requestId = responseData.requestId;
                throw error;
            }
        } catch (parseError) {
            if (parseError.code) {
                throw parseError; // This is an API error, re-throw
            }
            throw new Error(`Failed to parse response: ${parseError.message}`);
        }

        return responseData;
    }

    /**
     * Gets user balances
     * @returns {Promise<object>}
     */
    async getBalances() {
        return await this.makeRequest('GET', '/api/v1/balances');
    }

    /**
     * Gets swap estimation
     * @param {object} request - Estimation parameters
     * @param {string} request.from - Source asset
     * @param {string} request.to - Target asset
     * @param {string} request.amount - Amount
     * @param {string} [request.network] - Network
     * @param {string} [request.account] - Account
     * @returns {Promise<object>}
     */
    async estimateSwap(request) {
        return await this.makeRequest('POST', '/api/v1/estimate', request);
    }

    /**
     * Executes swap
     * @param {object} request - Swap parameters
     * @param {string} request.from - Source asset
     * @param {string} request.to - Target asset
     * @param {string} request.amount - Amount
     * @param {string} request.account - Account
     * @param {number} request.slippage_bps - Slippage in basis points
     * @param {string} [request.clientOrderId] - Client order ID
     * @param {string} idempotencyKey - Idempotency key
     * @returns {Promise<object>}
     */
    async swap(request, idempotencyKey) {
        return await this.makeRequest('POST', '/api/v1/swap', request, {
            'Idempotency-Key': idempotencyKey
        });
    }

    /**
     * Gets order status
     * @param {string} orderId - Order ID
     * @param {string} [clientOrderId] - Client order ID
     * @returns {Promise<object>}
     */
    async getOrderStatus(orderId, clientOrderId = null) {
        let path = `/api/v1/orders/${orderId}/status`;
        if (clientOrderId) {
            path += `?clientOrderId=${encodeURIComponent(clientOrderId)}`;
        }
        return await this.makeRequest('GET', path);
    }
}
