package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// BrokerClient represents the client for TheOne Trading API
type BrokerClient struct {
	apiKey    string
	secretKey string
	baseURL   string
	client    HTTPClient
}

// HTTPClient interface for HTTP client
type HTTPClient interface {
	Do(method, url string, headers map[string]string, body []byte) ([]byte, error)
}

// NewBrokerClient creates a new client instance
func NewBrokerClient(apiKey, secretKey, baseURL string, httpClient HTTPClient) *BrokerClient {
	return &BrokerClient{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		client:    httpClient,
	}
}

// Balance represents asset balance
type Balance struct {
	Asset  string `json:"asset"`
	Total  string `json:"total"`
	Locked string `json:"locked"`
}

// BalanceResponse represents response for balances request
type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

// EstimateRequestHTTP represents request for swap estimation (HTTP only)
type EstimateRequestHTTP struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Amount  string   `json:"amount"`
	Network string   `json:"network"`
	Account *string  `json:"account,omitempty"`
	Filter  []string `json:"filter,omitempty"` // Liquidity sources filter (binance, bybit, gate)
}

// RouteStep represents a swap route step
type RouteStep struct {
	Exchange  string `json:"exchange"`
	Pool      string `json:"pool"`
	FromAsset string `json:"from_asset"`
	ToAsset   string `json:"to_asset"`
	AmountIn  string `json:"amount_in"`
	AmountOut string `json:"amount_out"`
}

// EstimateResponse represents response for estimation request
type EstimateResponse struct {
	Route       []RouteStep `json:"route"`
	Price       string      `json:"price"`
	ExpectedOut string      `json:"expectedOut"`
	ExpiresAt   int64       `json:"expiresAt"`
}

// SwapRequestHTTP represents request for executing a swap (HTTP only)
type SwapRequestHTTP struct {
	From          string   `json:"from"`
	To            string   `json:"to"`
	Amount        string   `json:"amount"`
	Account       string   `json:"account"`
	SlippageBps   int      `json:"slippage_bps"`
	ClientOrderID *string  `json:"clientOrderId,omitempty"`
	Filter        []string `json:"filter,omitempty"` // Liquidity sources filter (binance, bybit, gate)
}

// SwapResponse represents response for swap request
type SwapResponse struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

// OrderStatusResponse represents order status response
type OrderStatusResponse struct {
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	FilledOut     string  `json:"filledOut,omitempty"`
	TxHash        string  `json:"txHash,omitempty"`
	UpdatedAt     int64   `json:"updatedAt"`
	ClientOrderID *string `json:"clientOrderId,omitempty"`
}

// APIError represents API error
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

// Error implements error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API Error [%s]: %s (RequestID: %s)", e.Code, e.Message, e.RequestID)
}

// generateSignature creates HMAC-SHA256 signature
func (c *BrokerClient) generateSignature(method, pathWithQuery string, timestamp int64, nonce, bodySHA256 string) string {
	canonicalString := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
		strings.ToUpper(method),
		pathWithQuery,
		timestamp,
		nonce,
		bodySHA256,
	)

	hash := sha256.Sum256([]byte(c.secretKey))
	secretKeyBase64 := base64.URLEncoding.EncodeToString(hash[:])
	hmacKey := []byte(secretKeyBase64)

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write([]byte(canonicalString))
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature
}

// hashBody creates SHA256 hash of request body
func hashBody(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// generateNonce creates unique nonce
func generateNonce() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
}

// makeRequest makes authenticated request to API
func (c *BrokerClient) makeRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	timestamp := time.Now().UnixMilli() // Using UnixMilli() as in the test
	nonce := generateNonce()
	bodySHA256 := hashBody(body)

	signature := c.generateSignature(method, path, timestamp, nonce, bodySHA256)

	headers := map[string]string{
		"Content-Type":    "application/json",
		"X-API-KEY":       c.apiKey,
		"X-API-TIMESTAMP": strconv.FormatInt(timestamp, 10),
		"X-API-NONCE":     nonce,
		"X-API-SIGN":      signature,
	}

	url := c.baseURL + path
	responseBody, err := c.client.Do(method, url, headers, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return responseBody, nil
}

// checkAPIError checks if response is an API error
func checkAPIError(responseBody []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(responseBody, &apiErr); err == nil && apiErr.Code != "" {
		return &apiErr
	}
	return nil
}

// GetBalances gets user balances
func (c *BrokerClient) GetBalances(ctx context.Context) (*BalanceResponse, error) {
	responseBody, err := c.makeRequest(ctx, "GET", "/api/v1/balances", nil)
	if err != nil {
		return nil, err
	}

	if apiErr := checkAPIError(responseBody); apiErr != nil {
		return nil, apiErr
	}

	var response BalanceResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// EstimateSwap gets swap estimation
func (c *BrokerClient) EstimateSwap(ctx context.Context, req *EstimateRequestHTTP) (*EstimateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	responseBody, err := c.makeRequest(ctx, "POST", "/api/v1/estimate", body)
	if err != nil {
		return nil, err
	}

	if apiErr := checkAPIError(responseBody); apiErr != nil {
		return nil, apiErr
	}

	var response EstimateResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// Swap executes swap
func (c *BrokerClient) Swap(ctx context.Context, req *SwapRequestHTTP, idempotencyKey string) (*SwapResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	timestamp := time.Now().UnixMilli() // Using UnixMilli() as in the test
	nonce := generateNonce()
	bodySHA256 := hashBody(body)

	signature := c.generateSignature("POST", "/api/v1/swap", timestamp, nonce, bodySHA256)

	headers := map[string]string{
		"Content-Type":    "application/json",
		"X-API-KEY":       c.apiKey,
		"X-API-TIMESTAMP": strconv.FormatInt(timestamp, 10),
		"X-API-NONCE":     nonce,
		"X-API-SIGN":      signature,
		"Idempotency-Key": idempotencyKey,
	}

	url := c.baseURL + "/api/v1/swap"
	responseBody, err := c.client.Do("POST", url, headers, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if apiErr := checkAPIError(responseBody); apiErr != nil {
		return nil, apiErr
	}

	var response SwapResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetOrderStatus gets order status
func (c *BrokerClient) GetOrderStatus(ctx context.Context, orderID string, clientOrderID *string) (*OrderStatusResponse, error) {
	path := fmt.Sprintf("/api/v1/orders/%s/status", orderID)
	if clientOrderID != nil {
		path += fmt.Sprintf("?clientOrderId=%s", *clientOrderID)
	}

	responseBody, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if apiErr := checkAPIError(responseBody); apiErr != nil {
		return nil, apiErr
	}

	var response OrderStatusResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// Example of client usage
func main() {
	// API keys (obtained from server)
	apiKey := "key"
	secretKey := "secret"
	baseURL := "http://127.0.0.1:8080"

	// Create HTTP client
	httpClient := NewDefaultHTTPClient()

	// Create API client
	client := NewBrokerClient(apiKey, secretKey, baseURL, httpClient)

	ctx := context.Background()

	// Example 1: Getting balances
	fmt.Println("=== Getting balances ===")
	balances, err := client.GetBalances(ctx)
	if err != nil {
		log.Printf("Error getting balances: %v\n", err)
	} else {
		fmt.Printf("Balances received: %+v\n", balances)
	}

	// Example 2: Swap estimation
	fmt.Println("\n=== Swap estimation ===")
	estimateReq := &EstimateRequestHTTP{
		From:   "TRX",
		To:     "USDT",
		Amount: "10",
		Filter: []string{"binance", "gate"}, // Use specific liquidity sources
	}

	estimate, err := client.EstimateSwap(ctx, estimateReq)
	if err != nil {
		log.Printf("Error getting estimation: %v\n", err)
	} else {
		fmt.Printf("Estimation received: %+v\n", estimate)
	}

	// Example 3: Executing swap (only if estimation exists)
	if estimate != nil {
		fmt.Println("\n=== Executing swap ===")
		swapReq := &SwapRequestHTTP{
			From:        estimateReq.From,
			To:          estimateReq.To,
			Amount:      estimateReq.Amount,
			SlippageBps: 30,
			Filter:      estimateReq.Filter, // Use same filter
		}

		idempotencyKey := fmt.Sprintf("swap_%d", time.Now().UnixNano())
		swapResponse, err := client.Swap(ctx, swapReq, idempotencyKey)
		if err != nil {
			log.Printf("Error executing swap: %v\n", err)
		} else {
			fmt.Printf("Swap created: %+v\n", swapResponse)

			// Example 4: Checking order status
			fmt.Println("\n=== Checking order status ===")
			for i := 0; i < 10; i++ {
				orderStatus, err := client.GetOrderStatus(ctx, swapResponse.OrderID, nil)
				if err != nil {
					log.Printf("Error getting order status: %v\n", err)
				} else {
					if orderStatus.Status == "PENDING" {
						fmt.Printf("Order %s is pending\n", orderStatus.OrderID)
						<-time.After(5 * time.Second)
						continue
					}

					fmt.Printf("Order status: %+v\n", orderStatus)
					break
				}
			}
		}
	}
}
