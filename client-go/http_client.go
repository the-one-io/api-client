package client_go

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultHTTPClient реализует интерфейс HTTPClient
type DefaultHTTPClient struct {
	client *http.Client
}

// NewDefaultHTTPClient создает новый HTTP клиент
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Do выполняет HTTP запрос
func (c *DefaultHTTPClient) Do(method, url string, headers map[string]string, body []byte) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем статус код
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseBody, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
