# REST API Documentation

## Загальна інформація

TheOne Trading API надає RESTful інтерфейс для виконання торгових операцій з підтримкою HMAC-SHA256 автентифікації.

**Base URL**: `https://partner-api.the-one.io`

**API Version**: `v1`

## Автентифікація

Всі приватні ендпоінти вимагають HMAC-SHA256 автентифікації з наступними заголовками:

| Заголовок | Тип | Опис |
|-----------|-----|------|
| `X-API-KEY` | string | Ваш API ключ |
| `X-API-SECRET` | string | Ваш секретний ключ (використовується для підпису) |
| `X-API-TIMESTAMP` | string | Часова мітка запиту в мілісекундах |
| `X-API-NONCE` | string | Унікальний рядок (UUID або timestamp-based) |
| `X-API-SIGN` | string | HMAC-SHA256 підпис |

### Формат канонічного рядка

Підпис генерується з канонічного рядка в такому форматі:

```
<HTTP_METHOD>\n<PATH_WITH_QUERY>\n<TIMESTAMP_MS>\n<NONCE>\n<BODY_SHA256_HEX>
```

**Приклад канонічного рядка**:
```
POST\n/api/v1/estimate\n1732526400000\nnonce_123\nsha256_hash_of_body
```

### Алгоритм генерації підпису

1. Створіть SHA256 хеш тіла запиту (для GET запитів - порожній рядок)
2. Сформуйте канонічний рядок
3. Обчисліть HMAC-SHA256 підпис використовуючи секретний ключ
4. Конвертуйте результат в hex формат

### Приклад коду (Go)

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

func generateSignature(secret, method, path string, timestamp int64, nonce, bodyHash string) string {
    canonical := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
        strings.ToUpper(method),
        path,
        timestamp,
        nonce,
        bodyHash,
    )
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(canonical))
    return hex.EncodeToString(mac.Sum(nil))
}

func hashBody(body []byte) string {
    hash := sha256.Sum256(body)
    return hex.EncodeToString(hash[:])
}
```

### Приклад коду (Python)

```python
import hashlib
import hmac
import time
import json

def generate_signature(secret, method, path, timestamp, nonce, body_hash):
    canonical = f"{method.upper()}\n{path}\n{timestamp}\n{nonce}\n{body_hash}"
    signature = hmac.new(
        secret.encode('utf-8'),
        canonical.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    return signature

def hash_body(body):
    if isinstance(body, dict):
        body = json.dumps(body, separators=(',', ':')).encode('utf-8')
    return hashlib.sha256(body).hexdigest()
```

### Приклад коду (JavaScript)

```javascript
const crypto = require('crypto');

function generateSignature(secret, method, path, timestamp, nonce, bodyHash) {
    const canonical = `${method.toUpperCase()}\n${path}\n${timestamp}\n${nonce}\n${bodyHash}`;
    return crypto
        .createHmac('sha256', secret)
        .update(canonical)
        .digest('hex');
}

function hashBody(body) {
    const bodyStr = typeof body === 'object' ? JSON.stringify(body) : body;
    return crypto
        .createHash('sha256')
        .update(bodyStr)
        .digest('hex');
}
```

## Обмеження швидкості (Rate Limiting)

| Тип ендпоінту | Ліміт |
|---------------|-------|
| Публічні | 100 запитів/хвилину |
| Приватні | 1000 запитів/хвилину |

## Формат помилок

```json
{
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "request_id": "uuid-request-id"
}
```

### Коди помилок

| Код | HTTP Status | Опис |
|-----|-------------|------|
| `UNAUTHORIZED` | 401 | Невалідна автентифікація |
| `INVALID_REQUEST` | 400 | Невалідний запит |
| `INTERNAL_ERROR` | 500 | Внутрішня помилка сервера |
| `NOT_FOUND` | 404 | Ресурс не знайдено |
| `RATE_LIMIT` | 429 | Перевищено ліміт запитів |
| `INVALID_SIGNATURE` | 401 | Невалідний підпис |
| `EXPIRED_REQUEST` | 401 | Застарілий запит (timestamp) |
| `INVALID_NONCE` | 401 | Nonce вже використаний або застарілий |
| `MISSING_API_KEY` | 401 | Відсутній API ключ |
| `DUPLICATE_REQUEST` | 409 | Дублікат запиту |
| `SWAP_ERROR` | 400 | Помилка виконання swap |

---

## Публічні ендпоінти

### Перевірка здоров'я сервісу

**GET** `/healthz`

Повертає статус здоров'я сервісу.

**Відповідь**:
```json
{
  "status": "ok",
  "time": 1732526400000
}
```

### Час сервера

**GET** `/api/v1/time`

Повертає поточний час сервера в мілісекундах (для синхронізації timestamp).

**Відповідь**:
```json
{
  "serverTime": 1732526400000
}
```

### Інформація про версію

**GET** `/version`

Повертає інформацію про версію API.

**Відповідь**:
```json
{
  "version": "1.0.0",
  "commit": "abc123def",
  "buildTime": "2024-01-15T10:30:00Z",
  "goVersion": "go1.22.0"
}
```

---

## Приватні ендпоінти

### 1. Оцінка свопу (Estimate)

**POST** `/api/v1/estimate`

Отримати оцінку ціни для операції обміну (swap).

**Заголовки**:
- `Content-Type: application/json`
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Тіло запиту**:
```json
{
  "from": "ETH",
  "to": "USDT",
  "amount": "1.5",
  "network": "ETH",
  "account": "0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
  "filter": ["binance", "gate"]
}
```

**Параметри**:

| Поле | Тип | Обов'язкове | Опис |
|------|-----|-------------|------|
| `from` | string | Так | Актив для обміну |
| `to` | string | Так | Актив для отримання |
| `amount` | string | Так | Кількість для обміну |
| `network` | string | Ні | Мережа (ETH, BSC тощо) |
| `account` | string | Ні | Адреса гаманця |
| `filter` | []string | Ні | Фільтр бірж/провайдерів |

**Відповідь** (200 OK):
```json
{
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
```

**Приклад cURL**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
METHOD="POST"
PATH="/api/v1/estimate"
BODY='{"from":"ETH","to":"USDT","amount":"1.5"}'
BODY_HASH=$(echo -n "$BODY" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X POST "https://partner-api.the-one.io/api/v1/estimate" \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE" \
  -d "$BODY"
```

---

### 2. Виконання свопу (Swap)

**POST** `/api/v1/swap`

Виконати операцію обміну (swap). Ендпоінт є ідемпотентним.

**Заголовки**:
- `Content-Type: application/json`
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`
- `Idempotency-Key: {unique_key}` - для запобігання дублікатів

**Тіло запиту**:
```json
{
  "from": "ETH",
  "to": "USDT",
  "amount": "1.5",
  "slippage_bps": 30,
  "clientOrderId": "client_ord_123",
  "filter": ["binance", "gate"]
}
```

**Параметри**:

| Поле | Тип | Обов'язкове | Опис |
|------|-----|-------------|------|
| `from` | string | Так | Актив для обміну |
| `to` | string | Так | Актив для отримання |
| `amount` | string | Так | Кількість для обміну |
| `slippage_bps` | integer | Ні | Допустиме відхилення в базисних пунктах (0-10000) |
| `clientOrderId` | string | Ні | Ваш ID замовлення для відстеження |
| `filter` | []string | Ні | Фільтр бірж/провайдерів |

**Відповідь** (200 OK):
```json
{
  "orderId": "ord_123456",
  "status": "PENDING"
}
```

**Можливі статуси**:
- `PENDING` - Замовлення створено, очікує виконання
- `PROCESSING` - Замовлення виконується
- `FILLED` - Замовлення виконано
- `FAILED` - Замовлення не виконано
- `CANCELLED` - Замовлення скасовано

**Приклад cURL**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
IDEMPOTENCY_KEY="swap_$(uuidgen)"
METHOD="POST"
PATH="/api/v1/swap"
BODY='{"from":"ETH","to":"USDT","amount":"1.5","slippage_bps":30}'
BODY_HASH=$(echo -n "$BODY" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X POST "https://partner-api.the-one.io/api/v1/swap" \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d "$BODY"
```

---

### 3. Статус замовлення

**GET** `/api/v1/orders/{id}/status`

Отримати статус замовлення за його ID або clientOrderId.

**Параметри шляху**:
- `id` - Order ID або Client Order ID

**Query параметри**:
- `clientOrderId` (optional) - Client Order ID як альтернатива до path параметра

**Заголовки**:
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Відповідь** (200 OK):
```json
{
  "orderId": "ord_123456",
  "clientOrderId": "client_ord_123",
  "status": "FILLED",
  "filledOut": "3245.67",
  "txHash": "0xabc123def456789abc123def456789abc123def456789abc123def456789abc123de",
  "updatedAt": 1732526402000
}
```

**Приклад cURL**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
ORDER_ID="ord_123456"
METHOD="GET"
PATH="/api/v1/orders/$ORDER_ID/status"
BODY_HASH=$(echo -n "" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X GET "https://partner-api.the-one.io/api/v1/orders/$ORDER_ID/status" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE"
```

---

### 4. Баланси

**GET** `/api/v1/balances`

Отримати баланси користувача для всіх активів.

**Заголовки**:
- `X-API-KEY: {your_api_key}`
- `X-API-TIMESTAMP: {timestamp_ms}`
- `X-API-NONCE: {unique_nonce}`
- `X-API-SIGN: {hmac_signature}`

**Відповідь** (200 OK):
```json
{
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
    },
    {
      "asset": "BTC",
      "total": "0.1",
      "locked": "0"
    }
  ]
}
```

**Поля балансу**:

| Поле | Тип | Опис |
|------|-----|------|
| `asset` | string | Символ активу |
| `total` | string | Загальний баланс |
| `locked` | string | Заблокований баланс (в активних замовленнях) |

**Приклад cURL**:
```bash
#!/bin/bash

API_KEY="your_api_key"
SECRET="your_secret"
TIMESTAMP=$(date +%s)000
NONCE="nonce_$(date +%s)"
METHOD="GET"
PATH="/api/v1/balances"
BODY_HASH=$(echo -n "" | sha256sum | cut -d' ' -f1)

CANONICAL="$METHOD\n$PATH\n$TIMESTAMP\n$NONCE\n$BODY_HASH"
SIGNATURE=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

curl -X GET "https://partner-api.the-one.io/api/v1/balances" \
  -H "X-API-KEY: $API_KEY" \
  -H "X-API-TIMESTAMP: $TIMESTAMP" \
  -H "X-API-NONCE: $NONCE" \
  -H "X-API-SIGN: $SIGNATURE"
```

---

## Управління API ключами

### Список API ключів

**GET** `/broker/v1/keys`

Отримати список всіх API ключів користувача.

**Query параметри**:
- `limit` (optional, default: 20, max: 100) - Кількість записів
- `offset` (optional, default: 0) - Зміщення для пагінації

**Відповідь** (200 OK):
```json
{
  "api_keys": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "key": "api_key_abc123",
      "name": "Production Key",
      "description": "Main production API key",
      "is_active": true,
      "permissions": ["read", "trade"],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "last_used_at": "2024-01-20T14:25:00Z",
      "expires_at": "2025-01-15T10:30:00Z",
      "metadata": {
        "ip_whitelist": ["192.168.1.1", "10.0.0.1"]
      }
    }
  ],
  "pagination": {
    "count": 1,
    "limit": 20,
    "offset": 0
  }
}
```

### Створити API ключ

**POST** `/broker/v1/keys`

Створити новий API ключ.

**Тіло запиту**:
```json
{
  "name": "Production Key",
  "description": "Main production API key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "expires_at": "2025-01-15T10:30:00Z",
  "metadata": {
    "ip_whitelist": ["192.168.1.1"]
  }
}
```

**Відповідь** (201 Created):
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "key": "api_key_abc123",
  "secret": "api_secret_xyz789",
  "name": "Production Key",
  "description": "Main production API key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2025-01-15T10:30:00Z"
}
```

**⚠️ Важливо**: Секретний ключ (`secret`) повертається тільки один раз при створенні!

### Отримати API ключ

**GET** `/broker/v1/keys/{uuid}`

Отримати інформацію про конкретний API ключ.

**Відповідь** (200 OK):
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "key": "api_key_abc123",
  "name": "Production Key",
  "is_active": true,
  "permissions": ["read", "trade"],
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Видалити API ключ

**DELETE** `/broker/v1/keys/{uuid}`

Видалити API ключ.

**Відповідь** (204 No Content)

---

## Найкращі практики

### Безпека

1. **Ніколи не розголошуйте секретний ключ** - зберігайте його в безпечному місці
2. **Використовуйте HTTPS** - завжди використовуйте захищене з'єднання
3. **Ротація ключів** - регулярно оновлюйте API ключі
4. **IP Whitelist** - обмежте доступ до API з конкретних IP адрес
5. **Мінімальні права** - надавайте тільки необхідні дозволи (permissions)

### Обробка помилок

1. **Перевіряйте код помилки** - використовуйте `code` для програмної обробки
2. **Логуйте request_id** - використовуйте для відстеження проблем
3. **Повтори запитів** - використовуйте exponential backoff для повторних спроб
4. **Обробка 429** - реалізуйте логіку очікування при rate limiting

### Продуктивність

1. **Кешування** - кешуйте час сервера та інші статичні дані
2. **Пакетні запити** - об'єднуйте запити коли можливо
3. **WebSocket** - використовуйте для real-time оновлень замість polling
4. **Idempotency Keys** - використовуйте для критичних операцій

### Тестування

1. **Тестове середовище** - використовуйте тестові ключі для розробки
2. **Логування** - логуйте всі запити та відповіді
3. **Моніторинг** - відстежуйте rate limits та помилки
4. **Алерти** - налаштуйте сповіщення для критичних помилок

---

## Приклади клієнтів

Готові клієнти доступні для різних мов програмування:

- **Go**: `examples/client/client-go/`
- **Python**: `examples/client/client-py/`
- **JavaScript**: `examples/client/client-js/`

Кожен клієнт включає повну реалізацію автентифікації та всіх ендпоінтів.

---

## Підтримка

Для отримання допомоги:
- Email: support@the-one.io
- Telegram: [@TheOneLiveSupportBot](https://t.me/TheOneLiveSupportBot)
- Документація: [https://partner-api.the-one.io/swagger/index.html](https://partner-api.the-one.io/swagger/index.html)
