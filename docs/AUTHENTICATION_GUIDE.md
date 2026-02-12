# Посібник з автентифікації API

## Огляд

TheOne Trading API використовує HMAC-SHA256 автентифікацію для захисту всіх приватних ендпоінтів. Цей посібник містить детальні інструкції та приклади коду для різних мов програмування.

## Отримання API ключів

### Тестові ключі (для розробки)

Для тестування API доступні наступні ключі:

| API Key | Secret |
|---------|--------|
| `test_key_1` | `test_secret_1` |
| `test_key_2` | `test_secret_2` |

### Створення продакшн ключів

Використовуйте ендпоінт `/broker/v1/keys` для створення власних API ключів:

```bash
curl -X POST "https://partner-api.the-one.io/broker/v1/keys" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_BROKER_TOKEN" \
  -d '{
    "name": "Production Key",
    "description": "Main production API key",
    "permissions": ["read", "trade"],
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

**⚠️ Важливо**: Секретний ключ показується тільки один раз при створенні. Збережіть його в безпечному місці!

---

## Алгоритм автентифікації

### Крок 1: Підготовка запиту

1. Визначте метод HTTP: `GET`, `POST`, `PUT`, `DELETE`
2. Визначте повний шлях з query параметрами: `/api/v1/estimate?param=value`
3. Підготуйте тіло запиту (для POST/PUT) або порожній рядок (для GET)

### Крок 2: Генерація timestamp та nonce

```
timestamp = поточний час в мілісекундах
nonce = унікальний рядок (UUID, timestamp + random, тощо)
```

### Крок 3: Обчислення хешу тіла

```
body_sha256 = SHA256(request_body)
body_hash_hex = hex_encode(body_sha256)
```

### Крок 4: Формування канонічного рядка

```
canonical_string = "{METHOD}\n{PATH}\n{TIMESTAMP}\n{NONCE}\n{BODY_HASH}"
```

**Приклад**:
```
POST\n/api/v1/estimate\n1732526400000\nnonce_123\ne3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

### Крок 5: Обчислення HMAC підпису

```
signature = HMAC-SHA256(secret_key, canonical_string)
signature_hex = hex_encode(signature)
```

### Крок 6: Додавання заголовків

```
X-API-KEY: your_api_key
X-API-TIMESTAMP: 1732526400000
X-API-NONCE: nonce_123
X-API-SIGN: computed_signature_hex
```

---

## Приклади коду

### Bash / cURL

```bash
#!/bin/bash

# Configuration
API_KEY="test_key_1"
SECRET="test_secret_1"
BASE_URL="https://partner-api.the-one.io"

# Request parameters
METHOD="POST"
PATH="/api/v1/estimate"
BODY='{"from":"ETH","to":"USDT","amount":"1.5"}'

# Generate timestamp and nonce
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)_$$"

# Calculate body hash
BODY_HASH=$(echo -n "$BODY" | sha256sum | awk '{print $1}')

# Create canonical string
CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"

# Generate signature
SIGNATURE=$(echo -ne "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')

# Make request
curl -X "$METHOD" "$BASE_URL$PATH" \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE" \
  -d "$BODY"
```

### Python

```python
import hashlib
import hmac
import time
import json
import requests
from typing import Optional, Dict, Any

class TradingAPIClient:
    def __init__(self, api_key: str, secret: str, base_url: str = "https://partner-api.the-one.io"):
        self.api_key = api_key
        self.secret = secret
        self.base_url = base_url
    
    def _generate_signature(
        self,
        method: str,
        path: str,
        timestamp: int,
        nonce: str,
        body: str = ""
    ) -> str:
        """Generate HMAC-SHA256 signature."""
        # Calculate body hash
        body_hash = hashlib.sha256(body.encode('utf-8')).hexdigest()
        
        # Create canonical string
        canonical = f"{method.upper()}\n{path}\n{timestamp}\n{nonce}\n{body_hash}"
        
        # Generate signature
        signature = hmac.new(
            self.secret.encode('utf-8'),
            canonical.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        
        return signature
    
    def _make_request(
        self,
        method: str,
        path: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Make authenticated API request."""
        # Generate timestamp and nonce
        timestamp = int(time.time() * 1000)
        nonce = f"nonce_{timestamp}"
        
        # Prepare body
        body = ""
        if data:
            body = json.dumps(data, separators=(',', ':'))
        
        # Add query params to path
        full_path = path
        if params:
            query_string = "&".join([f"{k}={v}" for k, v in params.items()])
            full_path = f"{path}?{query_string}"
        
        # Generate signature
        signature = self._generate_signature(method, full_path, timestamp, nonce, body)
        
        # Prepare headers
        headers = {
            "Content-Type": "application/json",
            "X-API-KEY": self.api_key,
            "X-API-TIMESTAMP": str(timestamp),
            "X-API-NONCE": nonce,
            "X-API-SIGN": signature
        }
        
        # Make request
        url = f"{self.base_url}{full_path}"
        response = requests.request(
            method=method,
            url=url,
            headers=headers,
            json=data if data else None
        )
        response.raise_for_status()
        
        return response.json()
    
    def estimate(self, from_asset: str, to_asset: str, amount: str, **kwargs) -> Dict[str, Any]:
        """Get swap estimate."""
        data = {
            "from": from_asset,
            "to": to_asset,
            "amount": amount,
            **kwargs
        }
        return self._make_request("POST", "/api/v1/estimate", data=data)
    
    def swap(self, from_asset: str, to_asset: str, amount: str, **kwargs) -> Dict[str, Any]:
        """Execute swap."""
        data = {
            "from": from_asset,
            "to": to_asset,
            "amount": amount,
            **kwargs
        }
        return self._make_request("POST", "/api/v1/swap", data=data)
    
    def get_order_status(self, order_id: str) -> Dict[str, Any]:
        """Get order status."""
        return self._make_request("GET", f"/api/v1/orders/{order_id}/status")
    
    def get_balances(self) -> Dict[str, Any]:
        """Get balances."""
        return self._make_request("GET", "/api/v1/balances")

# Usage example
if __name__ == "__main__":
    client = TradingAPIClient("test_key_1", "test_secret_1")
    
    # Get estimate
    estimate = client.estimate("ETH", "USDT", "1.5")
    print("Estimate:", estimate)
    
    # Execute swap
    swap_result = client.swap("ETH", "USDT", "1.5", slippage_bps=30)
    print("Swap:", swap_result)
    
    # Get balances
    balances = client.get_balances()
    print("Balances:", balances)
```

### JavaScript / Node.js

```javascript
const crypto = require('crypto');
const axios = require('axios');

class TradingAPIClient {
  constructor(apiKey, secret, baseURL = 'https://partner-api.the-one.io') {
    this.apiKey = apiKey;
    this.secret = secret;
    this.baseURL = baseURL;
  }

  generateSignature(method, path, timestamp, nonce, body = '') {
    // Calculate body hash
    const bodyHash = crypto
      .createHash('sha256')
      .update(body)
      .digest('hex');

    // Create canonical string
    const canonical = `${method.toUpperCase()}\n${path}\n${timestamp}\n${nonce}\n${bodyHash}`;

    // Generate signature
    const signature = crypto
      .createHmac('sha256', this.secret)
      .update(canonical)
      .digest('hex');

    return signature;
  }

  async makeRequest(method, path, data = null, params = null) {
    // Generate timestamp and nonce
    const timestamp = Date.now();
    const nonce = `nonce_${timestamp}_${Math.random()}`;

    // Prepare body
    const body = data ? JSON.stringify(data) : '';

    // Add query params to path
    let fullPath = path;
    if (params) {
      const queryString = Object.entries(params)
        .map(([k, v]) => `${k}=${v}`)
        .join('&');
      fullPath = `${path}?${queryString}`;
    }

    // Generate signature
    const signature = this.generateSignature(method, fullPath, timestamp, nonce, body);

    // Prepare headers
    const headers = {
      'Content-Type': 'application/json',
      'X-API-KEY': this.apiKey,
      'X-API-TIMESTAMP': timestamp.toString(),
      'X-API-NONCE': nonce,
      'X-API-SIGN': signature
    };

    // Make request
    const url = `${this.baseURL}${fullPath}`;
    const response = await axios({
      method: method,
      url: url,
      headers: headers,
      data: data
    });

    return response.data;
  }

  async estimate(fromAsset, toAsset, amount, options = {}) {
    const data = {
      from: fromAsset,
      to: toAsset,
      amount: amount,
      ...options
    };
    return await this.makeRequest('POST', '/api/v1/estimate', data);
  }

  async swap(fromAsset, toAsset, amount, options = {}) {
    const data = {
      from: fromAsset,
      to: toAsset,
      amount: amount,
      ...options
    };
    return await this.makeRequest('POST', '/api/v1/swap', data);
  }

  async getOrderStatus(orderId) {
    return await this.makeRequest('GET', `/api/v1/orders/${orderId}/status`);
  }

  async getBalances() {
    return await this.makeRequest('GET', '/api/v1/balances');
  }
}

// Usage example
(async () => {
  const client = new TradingAPIClient('test_key_1', 'test_secret_1');

  try {
    // Get estimate
    const estimate = await client.estimate('ETH', 'USDT', '1.5');
    console.log('Estimate:', estimate);

    // Execute swap
    const swapResult = await client.swap('ETH', 'USDT', '1.5', { slippage_bps: 30 });
    console.log('Swap:', swapResult);

    // Get balances
    const balances = await client.getBalances();
    console.log('Balances:', balances);
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
  }
})();
```

### Go

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

type TradingAPIClient struct {
    apiKey  string
    secret  string
    baseURL string
    client  *http.Client
}

func NewTradingAPIClient(apiKey, secret string) *TradingAPIClient {
    return &TradingAPIClient{
        apiKey:  apiKey,
        secret:  secret,
        baseURL: "https://partner-api.the-one.io",
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *TradingAPIClient) generateSignature(method, path string, timestamp int64, nonce, body string) string {
    // Calculate body hash
    bodyHash := sha256.Sum256([]byte(body))
    bodyHashHex := hex.EncodeToString(bodyHash[:])
    
    // Create canonical string
    canonical := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
        strings.ToUpper(method),
        path,
        timestamp,
        nonce,
        bodyHashHex,
    )
    
    // Generate signature
    mac := hmac.New(sha256.New, []byte(c.secret))
    mac.Write([]byte(canonical))
    signature := hex.EncodeToString(mac.Sum(nil))
    
    return signature
}

func (c *TradingAPIClient) makeRequest(method, path string, data interface{}) ([]byte, error) {
    // Generate timestamp and nonce
    timestamp := time.Now().UnixMilli()
    nonce := fmt.Sprintf("nonce_%d", timestamp)
    
    // Prepare body
    var body string
    var bodyReader io.Reader
    if data != nil {
        bodyBytes, err := json.Marshal(data)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal data: %w", err)
        }
        body = string(bodyBytes)
        bodyReader = strings.NewReader(body)
    } else {
        body = ""
    }
    
    // Generate signature
    signature := c.generateSignature(method, path, timestamp, nonce, body)
    
    // Create request
    url := c.baseURL + path
    req, err := http.NewRequest(method, url, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-KEY", c.apiKey)
    req.Header.Set("X-API-TIMESTAMP", fmt.Sprintf("%d", timestamp))
    req.Header.Set("X-API-NONCE", nonce)
    req.Header.Set("X-API-SIGN", signature)
    
    // Make request
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Read response
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Check status code
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return respBody, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
    }
    
    return respBody, nil
}

func (c *TradingAPIClient) Estimate(fromAsset, toAsset, amount string) (map[string]interface{}, error) {
    data := map[string]interface{}{
        "from":   fromAsset,
        "to":     toAsset,
        "amount": amount,
    }
    
    respBody, err := c.makeRequest("POST", "/api/v1/estimate", data)
    if err != nil {
        return nil, err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return result, nil
}

func (c *TradingAPIClient) Swap(fromAsset, toAsset, amount string, slippageBps int) (map[string]interface{}, error) {
    data := map[string]interface{}{
        "from":         fromAsset,
        "to":           toAsset,
        "amount":       amount,
        "slippage_bps": slippageBps,
    }
    
    respBody, err := c.makeRequest("POST", "/api/v1/swap", data)
    if err != nil {
        return nil, err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return result, nil
}

func (c *TradingAPIClient) GetBalances() (map[string]interface{}, error) {
    respBody, err := c.makeRequest("GET", "/api/v1/balances", nil)
    if err != nil {
        return nil, err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return result, nil
}

func main() {
    client := NewTradingAPIClient("test_key_1", "test_secret_1")
    
    // Get estimate
    estimate, err := client.Estimate("ETH", "USDT", "1.5")
    if err != nil {
        fmt.Printf("Error getting estimate: %v\n", err)
    } else {
        fmt.Printf("Estimate: %+v\n", estimate)
    }
    
    // Get balances
    balances, err := client.GetBalances()
    if err != nil {
        fmt.Printf("Error getting balances: %v\n", err)
    } else {
        fmt.Printf("Balances: %+v\n", balances)
    }
}
```

---

## Тестування автентифікації

### Перевірка підпису

Використайте ендпоінт `/api/v1/time` для отримання серверного часу:

```bash
curl https://partner-api.the-one.io/api/v1/time
```

Потім використайте цей час для створення підпису та тестування автентифікації на ендпоінті `/api/v1/balances`:

```bash
#!/bin/bash
API_KEY="test_key_1"
SECRET="test_secret_1"

# Get server time
SERVER_TIME=$(curl -s https://partner-api.the-one.io/api/v1/time | jq -r '.serverTime')

TIMESTAMP=$SERVER_TIME
NONCE="test_nonce_$(date +%s)"
METHOD="GET"
PATH="/api/v1/balances"
BODY_HASH=$(echo -n "" | sha256sum | awk '{print $1}')

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -ne "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')

curl -X GET "https://partner-api.the-one.io/api/v1/balances" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE"
```

---

## Поширені помилки та рішення

### 1. Invalid Signature

**Причина**: Підпис не співпадає з очікуваним значенням.

**Рішення**:
- Перевірте, що ви використовуєте правильний секретний ключ
- Переконайтеся, що канонічний рядок сформовано правильно
- Перевірте порядок елементів у канонічному рядку
- Переконайтеся, що використовується `\n` (не `\\n`) для розділювача

### 2. Request Timestamp Outside Allowed Window

**Причина**: Різниця між серверним часом та timestamp у запиті більша за дозволену (30 секунд).

**Рішення**:
- Синхронізуйте час на вашому сервері
- Отримайте серверний час через `/api/v1/time`
- Використовуйте мілісекунди, не секунди

### 3. Nonce Has Already Been Used

**Причина**: Той самий nonce був використаний двічі.

**Рішення**:
- Використовуйте унікальний nonce для кожного запиту
- Додайте випадковий компонент до nonce
- Використовуйте UUID або timestamp + random

### 4. Missing API Key

**Причина**: Відсутній заголовок `X-API-KEY`.

**Рішення**:
- Переконайтеся, що всі заголовки додано правильно
- Перевірте написання заголовків (case-sensitive)

---

## Безпека

### Найкращі практики

1. **Ніколи не передавайте secret у запиті** - використовуйте його тільки для генерації підпису
2. **Зберігайте ключі безпечно** - використовуйте змінні оточення або секретні менеджери
3. **Використовуйте HTTPS** - ніколи не використовуйте HTTP для запитів
4. **Ротація ключів** - регулярно оновлюйте API ключі
5. **Моніторинг** - відстежуйте невдалі спроби автентифікації

### Приклад зберігання ключів (.env файл)

```env
TRADING_API_KEY=your_api_key_here
TRADING_API_SECRET=your_secret_here
TRADING_API_BASE_URL=https://partner-api.the-one.io
```

### Завантаження з середовища (Python)

```python
import os
from dotenv import load_dotenv

load_dotenv()

api_key = os.getenv('TRADING_API_KEY')
secret = os.getenv('TRADING_API_SECRET')
base_url = os.getenv('TRADING_API_BASE_URL')

client = TradingAPIClient(api_key, secret, base_url)
```

---

## Підтримка

Якщо у вас виникли проблеми з автентифікацією:

- Email: support@the-one.io
- Telegram: [@TheOneLiveSupportBot](https://t.me/TheOneLiveSupportBot)
- Документація: [https://partner-api.the-one.io/swagger/index.html](https://partner-api.the-one.io/swagger/index.html)
