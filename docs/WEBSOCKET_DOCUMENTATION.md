# WebSocket API Documentation

## Загальна інформація

TheOne Trading API надає WebSocket інтерфейс для real-time комунікації, включаючи підписки на події та виконання торгових операцій.

**WebSocket URL**: `wss://partner-api.the-one.io/ws/v1/stream`

**Protocol**: WebSocket (RFC 6455)

## Підключення

```javascript
const ws = new WebSocket('wss://partner-api.the-one.io/ws/v1/stream');
```

## Автентифікація

Перше повідомлення після підключення **обов'язково** має бути повідомленням автентифікації.

### Формат повідомлення автентифікації

```json
{
  "op": "auth",
  "key": "your_api_key",
  "ts": 1732526400000,
  "nonce": "unique_nonce_123",
  "sig": "hmac_signature"
}
```

### Поля автентифікації

| Поле | Тип | Обов'язкове | Опис |
|------|-----|-------------|------|
| `op` | string | Так | Операція (завжди "auth" для автентифікації) |
| `key` | string | Так | Ваш API ключ |
| `ts` | integer | Так | Timestamp в мілісекундах |
| `nonce` | string | Так | Унікальний nonce |
| `sig` | string | Так | HMAC-SHA256 підпис |

### Генерація підпису для автентифікації

Для WebSocket автентифікації використовується той самий алгоритм HMAC-SHA256, що і для REST API:

**Канонічний рядок**:
```
WS\n/ws/v1/stream\n<TIMESTAMP>\n<NONCE>\n<EMPTY_BODY_HASH>
```

**Приклад (JavaScript)**:
```javascript
const crypto = require('crypto');

function authenticateWebSocket(apiKey, secret) {
  const timestamp = Date.now();
  const nonce = `ws_${timestamp}_${Math.random()}`;
  
  // Empty body for auth
  const bodyHash = crypto.createHash('sha256').update('').digest('hex');
  
  // Canonical string
  const canonical = `WS\n/ws/v1/stream\n${timestamp}\n${nonce}\n${bodyHash}`;
  
  // Generate signature
  const signature = crypto
    .createHmac('sha256', secret)
    .update(canonical)
    .digest('hex');
  
  return {
    op: 'auth',
    key: apiKey,
    ts: timestamp,
    nonce: nonce,
    sig: signature
  };
}

// Usage
const authMsg = authenticateWebSocket('your_api_key', 'your_secret');
ws.send(JSON.stringify(authMsg));
```

**Приклад (Python)**:
```python
import hashlib
import hmac
import time
import json

def authenticate_websocket(api_key, secret):
    timestamp = int(time.time() * 1000)
    nonce = f"ws_{timestamp}"
    
    # Empty body for auth
    body_hash = hashlib.sha256(b'').hexdigest()
    
    # Canonical string
    canonical = f"WS\n/ws/v1/stream\n{timestamp}\n{nonce}\n{body_hash}"
    
    # Generate signature
    signature = hmac.new(
        secret.encode('utf-8'),
        canonical.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    return {
        'op': 'auth',
        'key': api_key,
        'ts': timestamp,
        'nonce': nonce,
        'sig': signature
    }

# Usage
auth_msg = authenticate_websocket('your_api_key', 'your_secret')
ws.send(json.dumps(auth_msg))
```

**Приклад (Go)**:
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

func authenticateWebSocket(apiKey, secret string) ([]byte, error) {
    timestamp := time.Now().UnixMilli()
    nonce := fmt.Sprintf("ws_%d", timestamp)
    
    // Empty body for auth
    bodyHash := sha256.Sum256([]byte{})
    bodyHashHex := hex.EncodeToString(bodyHash[:])
    
    // Canonical string
    canonical := fmt.Sprintf("WS\n/ws/v1/stream\n%d\n%s\n%s",
        timestamp, nonce, bodyHashHex)
    
    // Generate signature
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(canonical))
    signature := hex.EncodeToString(mac.Sum(nil))
    
    authMsg := map[string]interface{}{
        "op":    "auth",
        "key":   apiKey,
        "ts":    timestamp,
        "nonce": nonce,
        "sig":   signature,
    }
    
    return json.Marshal(authMsg)
}

// Usage
authMsg, _ := authenticateWebSocket("your_api_key", "your_secret")
ws.WriteMessage(websocket.TextMessage, authMsg)
```

### Відповідь на автентифікацію

**Успішна автентифікація**:
```json
{
  "op": "auth",
  "data": {
    "status": "authenticated"
  }
}
```

**Помилка автентифікації**:
```json
{
  "error": "Invalid signature"
}
```

---

## Підписки (Subscriptions)

Після успішної автентифікації можна підписуватися на канали подій.

### Підписка на канал

```json
{
  "op": "subscribe",
  "ch": "channel_name"
}
```

**Відповідь**:
```json
{
  "op": "subscribe",
  "ch": "channel_name",
  "data": {
    "status": "subscribed"
  }
}
```

### Відписка від каналу

```json
{
  "op": "unsubscribe",
  "ch": "channel_name"
}
```

**Відповідь**:
```json
{
  "op": "unsubscribe",
  "ch": "channel_name",
  "data": {
    "status": "unsubscribed"
  }
}
```

---

## Доступні канали

### 1. Баланси (`balances`)

Отримання real-time оновлень балансів.

**Підписка**:
```json
{
  "op": "subscribe",
  "ch": "balances"
}
```

**Повідомлення про оновлення**:
```json
{
  "ch": "balances",
  "data": [
    {
      "asset": "USDT",
      "total": "101.23",
      "locked": "0"
    },
    {
      "asset": "ETH",
      "total": "2.6",
      "locked": "0.4"
    },
    {
      "asset": "BTC",
      "total": "0.11",
      "locked": "0"
    }
  ]
}
```

### 2. Статус замовлень (`orders:<orderId>`)

Отримання real-time оновлень про статус конкретного замовлення.

**Підписка**:
```json
{
  "op": "subscribe",
  "ch": "orders:ord_123456"
}
```

**Повідомлення про оновлення**:
```json
{
  "ch": "orders:ord_123456",
  "data": {
    "orderId": "ord_123456",
    "status": "FILLED",
    "filledOut": "3245.67",
    "txHash": "0xabc123def456789abc123def456789abc123def456789abc123def456789abc123de",
    "updatedAt": 1732526402000
  }
}
```

**Можливі статуси**:
- `PENDING` - Очікування виконання
- `PROCESSING` - В процесі виконання
- `FILLED` - Виконано
- `FAILED` - Не виконано
- `CANCELLED` - Скасовано

---

## Виконання операцій через WebSocket

WebSocket API підтримує виконання торгових операцій з автентифікацією кожного повідомлення.

### Підпис повідомлень операцій

Для операцій (estimate, swap, order_status, balances) потрібно підписувати кожне повідомлення:

**Канонічний рядок**:
```
WS\n/ws/v1/<OPERATION>\n<TIMESTAMP>\n<NONCE>\n<DATA_BODY_SHA256>
```

**Приклад генерації підпису (JavaScript)**:
```javascript
function signMessage(secret, operation, data, timestamp, nonce) {
  const dataStr = JSON.stringify(data);
  const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
  
  const canonical = `WS\n/ws/v1/${operation}\n${timestamp}\n${nonce}\n${bodyHash}`;
  
  return crypto
    .createHmac('sha256', secret)
    .update(canonical)
    .digest('hex');
}
```

### 1. Оцінка свопу (Estimate)

**Запит**:
```json
{
  "op": "estimate",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "from": "ETH",
    "to": "USDT",
    "amount": "1.5",
    "network": "ETH",
    "filter": ["binance", "gate"]
  }
}
```

**Відповідь**:
```json
{
  "op": "estimate",
  "data": {
    "route": [
      {
        "from_asset": "ETH",
        "to_asset": "USDT",
        "amount_in": "1.5",
        "amount_out": "3245.67",
        "pool": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
      }
    ],
    "price": "2163.78",
    "expectedOut": "3245.67",
    "fee": "0.003",
    "expiresAt": 1732526400000
  }
}
```

### 2. Виконання свопу (Swap)

**Запит**:
```json
{
  "op": "swap",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "from": "ETH",
    "to": "USDT",
    "amount": "1.5",
    "slippage_bps": 30,
    "clientOrderId": "client_ord_123",
    "filter": ["binance", "gate"]
  }
}
```

**Відповідь**:
```json
{
  "op": "swap",
  "data": {
    "orderId": "ord_123456",
    "status": "PENDING"
  }
}
```

### 3. Статус замовлення (Order Status)

**Запит**:
```json
{
  "op": "order_status",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {
    "id": "ord_123456"
  }
}
```

**Відповідь**:
```json
{
  "op": "order_status",
  "data": {
    "orderId": "ord_123456",
    "status": "FILLED",
    "filledOut": "3245.67",
    "txHash": "0xabc123...",
    "updatedAt": 1732526402000
  }
}
```

### 4. Отримання балансів (Balances)

**Запит**:
```json
{
  "op": "balances",
  "ts": 1732526400000,
  "nonce": "unique_nonce",
  "sig": "hmac_signature",
  "data": {}
}
```

**Відповідь**:
```json
{
  "op": "balances",
  "data": {
    "balances": [
      {
        "asset": "USDT",
        "total": "100.00",
        "locked": "0"
      },
      {
        "asset": "ETH",
        "total": "2.5",
        "locked": "0.5"
      }
    ]
  }
}
```

---

## Формат помилок

```json
{
  "error": "Error message description"
}
```

**Типові помилки**:
- `"Not authenticated"` - Спроба виконати операцію без автентифікації
- `"Invalid signature"` - Невалідний підпис повідомлення
- `"Invalid API key"` - Невалідний API ключ
- `"Request timestamp outside allowed window"` - Застарілий timestamp
- `"Nonce has already been used or is expired"` - Nonce вже використаний
- `"Unknown operation: <op>"` - Невідома операція
- `"Invalid channel format"` - Невалідний формат каналу

---

## Повний приклад (JavaScript)

```javascript
const WebSocket = require('ws');
const crypto = require('crypto');

class TradingWebSocketClient {
  constructor(apiKey, secret, url = 'wss://partner-api.the-one.io/ws/v1/stream') {
    this.apiKey = apiKey;
    this.secret = secret;
    this.url = url;
    this.ws = null;
    this.authenticated = false;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      
      this.ws.on('open', () => {
        console.log('WebSocket connected');
        this.authenticate()
          .then(() => resolve())
          .catch(err => reject(err));
      });
      
      this.ws.on('message', (data) => {
        this.handleMessage(JSON.parse(data.toString()));
      });
      
      this.ws.on('error', (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      });
      
      this.ws.on('close', () => {
        console.log('WebSocket disconnected');
        this.authenticated = false;
      });
    });
  }

  authenticate() {
    return new Promise((resolve, reject) => {
      const timestamp = Date.now();
      const nonce = `ws_auth_${timestamp}`;
      const bodyHash = crypto.createHash('sha256').update('').digest('hex');
      
      const canonical = `WS\n/ws/v1/stream\n${timestamp}\n${nonce}\n${bodyHash}`;
      const signature = crypto
        .createHmac('sha256', this.secret)
        .update(canonical)
        .digest('hex');
      
      const authMsg = {
        op: 'auth',
        key: this.apiKey,
        ts: timestamp,
        nonce: nonce,
        sig: signature
      };
      
      this.ws.send(JSON.stringify(authMsg));
      
      // Wait for auth response
      const handler = (msg) => {
        if (msg.op === 'auth') {
          if (msg.data?.status === 'authenticated') {
            this.authenticated = true;
            console.log('Authenticated successfully');
            resolve();
          } else if (msg.error) {
            reject(new Error(msg.error));
          }
        }
      };
      
      this.ws.once('message', (data) => {
        handler(JSON.parse(data.toString()));
      });
    });
  }

  subscribe(channel) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    this.ws.send(JSON.stringify({
      op: 'subscribe',
      ch: channel
    }));
  }

  unsubscribe(channel) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    this.ws.send(JSON.stringify({
      op: 'unsubscribe',
      ch: channel
    }));
  }

  async estimate(from, to, amount, options = {}) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    const timestamp = Date.now();
    const nonce = `estimate_${timestamp}`;
    
    const data = {
      from,
      to,
      amount,
      ...options
    };
    
    const dataStr = JSON.stringify(data);
    const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
    const canonical = `WS\n/ws/v1/estimate\n${timestamp}\n${nonce}\n${bodyHash}`;
    const signature = crypto
      .createHmac('sha256', this.secret)
      .update(canonical)
      .digest('hex');
    
    this.ws.send(JSON.stringify({
      op: 'estimate',
      ts: timestamp,
      nonce: nonce,
      sig: signature,
      data: data
    }));
  }

  async swap(from, to, amount, options = {}) {
    if (!this.authenticated) {
      throw new Error('Not authenticated');
    }
    
    const timestamp = Date.now();
    const nonce = `swap_${timestamp}`;
    
    const data = {
      from,
      to,
      amount,
      ...options
    };
    
    const dataStr = JSON.stringify(data);
    const bodyHash = crypto.createHash('sha256').update(dataStr).digest('hex');
    const canonical = `WS\n/ws/v1/swap\n${timestamp}\n${nonce}\n${bodyHash}`;
    const signature = crypto
      .createHmac('sha256', this.secret)
      .update(canonical)
      .digest('hex');
    
    this.ws.send(JSON.stringify({
      op: 'swap',
      ts: timestamp,
      nonce: nonce,
      sig: signature,
      data: data
    }));
  }

  handleMessage(msg) {
    console.log('Received message:', msg);
    
    if (msg.error) {
      console.error('Error:', msg.error);
      return;
    }
    
    // Handle different message types
    if (msg.ch) {
      // Channel update
      console.log(`Update from ${msg.ch}:`, msg.data);
    } else if (msg.op) {
      // Operation response
      console.log(`Response for ${msg.op}:`, msg.data);
    }
  }

  close() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// Usage example
(async () => {
  const client = new TradingWebSocketClient('your_api_key', 'your_secret');
  
  try {
    await client.connect();
    
    // Subscribe to balances
    client.subscribe('balances');
    
    // Subscribe to specific order updates
    client.subscribe('orders:ord_123456');
    
    // Get estimate
    await client.estimate('ETH', 'USDT', '1.5');
    
    // Execute swap
    await client.swap('ETH', 'USDT', '1.5', {
      slippage_bps: 30,
      clientOrderId: 'my_order_123'
    });
    
    // Keep connection open
    setTimeout(() => {
      client.close();
    }, 60000); // Close after 1 minute
    
  } catch (error) {
    console.error('Error:', error);
  }
})();
```

---

## Найкращі практики

### Підключення

1. **Reconnection Logic** - Реалізуйте автоматичне перепідключення
2. **Heartbeat/Ping** - Використовуйте WebSocket ping/pong для перевірки з'єднання
3. **Timeout** - Встановіть timeout для операцій
4. **Error Handling** - Обробляйте всі типи помилок

### Безпека

1. **Secure Connection** - Завжди використовуйте WSS (WebSocket Secure)
2. **Nonce Uniqueness** - Використовуйте унікальні nonce для кожного повідомлення
3. **Timestamp Validation** - Синхронізуйте час з сервером
4. **Secret Protection** - Ніколи не передавайте secret через з'єднання

### Продуктивність

1. **Batching** - Об'єднуйте підписки коли можливо
2. **Selective Subscriptions** - Підписуйтесь тільки на потрібні канали
3. **Message Queuing** - Використовуйте чергу для відправки повідомлень
4. **Connection Pooling** - Повторно використовуйте з'єднання

---

## Обмеження

- Максимум **100 одночасних WebSocket з'єднань** на API ключ
- Максимум **50 підписок** на одне з'єднання
- Максимум **100 повідомлень/секунду** на одне з'єднання
- Автоматичне відключення після **5 хвилин неактивності**
- Максимальний розмір повідомлення: **1 MB**

---

## Підтримка

Для отримання допомоги:
- Email: support@the-one.io
- Telegram: [@TheOneLiveSupportBot](https://t.me/TheOneLiveSupportBot)
- Документація: [https://partner-api.the-one.io/swagger/index.html](https://partner-api.the-one.io/swagger/index.html)
