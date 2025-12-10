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

// BrokerClient представляет клиент для Broker Trading API
type BrokerClient struct {
	apiKey    string
	secretKey string
	baseURL   string
	client    HTTPClient
}

// HTTPClient интерфейс для HTTP клиента
type HTTPClient interface {
	Do(method, url string, headers map[string]string, body []byte) ([]byte, error)
}

// NewBrokerClient создает новый экземпляр клиента
func NewBrokerClient(apiKey, secretKey, baseURL string, httpClient HTTPClient) *BrokerClient {
	return &BrokerClient{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		client:    httpClient,
	}
}

// Balance представляет баланс активов
type Balance struct {
	Asset  string `json:"asset"`
	Total  string `json:"total"`
	Locked string `json:"locked"`
}

// BalanceResponse представляет ответ на запрос балансов
type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

// EstimateRequest представляет запрос на оценку свопа
type EstimateRequest struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	Amount  string  `json:"amount"`
	Network string  `json:"network"`
	Account *string `json:"account,omitempty"`
}

// RouteStep представляет шаг маршрута свопа
type RouteStep struct {
	Exchange  string `json:"exchange"`
	Pool      string `json:"pool"`
	FromAsset string `json:"from_asset"`
	ToAsset   string `json:"to_asset"`
	AmountIn  string `json:"amount_in"`
	AmountOut string `json:"amount_out"`
}

// EstimateResponse представляет ответ на запрос оценки
type EstimateResponse struct {
	Route       []RouteStep `json:"route"`
	Price       string      `json:"price"`
	ExpectedOut string      `json:"expectedOut"`
	ExpiresAt   int64       `json:"expiresAt"`
}

// SwapRequest представляет запрос на выполнение свопа
type SwapRequest struct {
	From          string  `json:"from"`
	To            string  `json:"to"`
	Amount        string  `json:"amount"`
	Account       string  `json:"account"`
	SlippageBps   int     `json:"slippage_bps"`
	ClientOrderID *string `json:"clientOrderId,omitempty"`
}

// SwapResponse представляет ответ на запрос свопа
type SwapResponse struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

// OrderStatusResponse представляет ответ статуса ордера
type OrderStatusResponse struct {
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	FilledOut     string  `json:"filledOut,omitempty"`
	TxHash        string  `json:"txHash,omitempty"`
	UpdatedAt     int64   `json:"updatedAt"`
	ClientOrderID *string `json:"clientOrderId,omitempty"`
}

// APIError представляет ошибку API
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

// Error реализует интерфейс error
func (e *APIError) Error() string {
	return fmt.Sprintf("API Error [%s]: %s (RequestID: %s)", e.Code, e.Message, e.RequestID)
}

// generateSignature создает HMAC-SHA256 подпись
func (c *BrokerClient) generateSignature(method, pathWithQuery string, timestamp int64, nonce, bodySHA256 string) string {
	canonicalString := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
		strings.ToUpper(method),
		pathWithQuery,
		timestamp,
		nonce,
		bodySHA256,
	)

	// Отладочная информация
	fmt.Printf("Creating signature:\n")
	fmt.Printf("Method: %s\n", strings.ToUpper(method))
	fmt.Printf("Path: %s\n", pathWithQuery)
	fmt.Printf("Timestamp: %d\n", timestamp)
	fmt.Printf("Nonce: %s\n", nonce)
	fmt.Printf("BodySHA256: %s\n", bodySHA256)
	fmt.Printf("CanonicalString: %q\n", canonicalString)

	hash := sha256.Sum256([]byte(c.secretKey))
	mac := hmac.New(sha256.New, []byte(base64.URLEncoding.EncodeToString(hash[:])))
	mac.Write([]byte(canonicalString))
	signature := hex.EncodeToString(mac.Sum(nil))
	fmt.Printf("Signature: %s\n", signature)
	fmt.Printf("---\n")

	return signature
}

// hashBody создает SHA256 хеш тела запроса
func hashBody(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// generateNonce создает уникальный nonce
func generateNonce() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
}

// makeRequest выполняет аутентифицированный запрос к API
func (c *BrokerClient) makeRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	timestamp := time.Now().UnixMilli() // Используем UnixMilli() как в тесте
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

// checkAPIError проверяет, является ли ответ ошибкой API
func checkAPIError(responseBody []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(responseBody, &apiErr); err == nil && apiErr.Code != "" {
		return &apiErr
	}
	return nil
}

// GetBalances получает балансы пользователя
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

// EstimateSwap получает оценку свопа
func (c *BrokerClient) EstimateSwap(ctx context.Context, req *EstimateRequest) (*EstimateResponse, error) {
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

// Swap выполняет своп
func (c *BrokerClient) Swap(ctx context.Context, req *SwapRequest, idempotencyKey string) (*SwapResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	timestamp := time.Now().UnixMilli() // Используем UnixMilli() как в тесте
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

// GetOrderStatus получает статус ордера
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

// Пример использования клиента
func main() {
	// API ключи (полученные от сервера)
	apiKey := ""
	secretKey := ""
	baseURL := "https://partner-api-dev.the-one.io"

	// Создаем HTTP клиент
	httpClient := NewDefaultHTTPClient()

	// Создаем клиент API
	client := NewBrokerClient(apiKey, secretKey, baseURL, httpClient)

	ctx := context.Background()

	// Пример 1: Получение балансов
	fmt.Println("=== Получение балансов ===")
	balances, err := client.GetBalances(ctx)
	if err != nil {
		log.Printf("Ошибка при получении балансов: %v\n", err)
	} else {
		fmt.Printf("Балансы получены: %+v\n", balances)
	}

	// Пример 2: Оценка свопа
	fmt.Println("\n=== Оценка свопа ===")
	estimateReq := &EstimateRequest{
		From:   "USDT",
		To:     "BTC",
		Amount: "10",
	}

	estimate, err := client.EstimateSwap(ctx, estimateReq)
	if err != nil {
		log.Printf("Ошибка при получении оценки: %v\n", err)
	} else {
		fmt.Printf("Оценка получена: %+v\n", estimate)
	}

	// Пример 3: Выполнение свопа (только если есть оценка)
	if estimate != nil {
		fmt.Println("\n=== Выполнение свопа ===")
		swapReq := &SwapRequest{
			From:        estimateReq.From,
			To:          estimateReq.To,
			Amount:      estimateReq.Amount,
			SlippageBps: 30,
		}

		idempotencyKey := fmt.Sprintf("swap_%d", time.Now().UnixNano())
		swapResponse, err := client.Swap(ctx, swapReq, idempotencyKey)
		if err != nil {
			log.Printf("Ошибка при выполнении свопа: %v\n", err)
		} else {
			fmt.Printf("Своп создан: %+v\n", swapResponse)

			// Пример 4: Проверка статуса ордера
			<-time.After(time.Second)
			fmt.Println("\n=== Проверка статуса ордера ===")
			orderStatus, err := client.GetOrderStatus(ctx, swapResponse.OrderID, nil)
			if err != nil {
				log.Printf("Ошибка при получении статуса ордера: %v\n", err)
			} else {
				fmt.Printf("Статус ордера: %+v\n", orderStatus)
			}
		}
	}
}
