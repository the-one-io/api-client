# WebSocket Client for JavaScript/Node.js

WebSocket client для Broker Trading API на JavaScript/Node.js с поддержкой аутентификации и реального времени.

## Установка

```bash
npm install
```

## Зависимости

- `ws` - WebSocket библиотека для Node.js
- `node-fetch` - Для HTTP запросов (уже установлено для REST клиента)

## Использование

### Простое подключение

```javascript
import { BrokerWSClient } from './ws-client.js';

async function main() {
    // API ключи
    const apiKey = 'your_api_key';
    const secretKey = 'your_secret_key';
    const wsURL = 'wss://api.example.com/ws';
    
    // Создание клиента
    const client = new BrokerWSClient(apiKey, secretKey, wsURL);
    
    // Подключение
    await client.connect();
    
    // Подписка на баланс
    await client.subscribe('balances', (message) => {
        console.log('Balances update:', message.data);
    });
    
    // Graceful shutdown
    process.on('SIGINT', () => {
        client.close();
        process.exit(0);
    });
}

main().catch(console.error);
```

### Подписка на каналы

#### Баланс пользователя

```javascript
await client.subscribe('balances', (message) => {
    console.log('Received balances:', message.data);
});
```

#### Статус ордера

```javascript
const orderID = 'ord_12345678';
const channel = `orders:${orderID}`;

await client.subscribe(channel, (message) => {
    console.log(`Order ${orderID} status:`, message.data);
});
```

### Обработка событий

```javascript
// Аутентификация
client.on('authenticated', () => {
    console.log('✅ Authenticated successfully');
});

// Отключение
client.on('disconnected', ({ code, reason }) => {
    console.log(`❌ Disconnected: ${code} ${reason}`);
});

// Ошибки
client.on('error', (error) => {
    console.error('❌ WebSocket error:', error.message);
});

// Подписки
client.on('subscribed', (channel) => {
    console.log(`✅ Subscribed to: ${channel}`);
});

client.on('unsubscribed', (channel) => {
    console.log(`✅ Unsubscribed from: ${channel}`);
});
```

### Обработка ошибок

```javascript
try {
    await client.connect();
    await client.subscribe('balances', handleBalances);
} catch (error) {
    console.error('Connection failed:', error);
}

function handleBalances(message) {
    if (message.error) {
        console.error('Message error:', message.error);
        return;
    }
    
    // Обработка данных
    console.log('Data:', message.data);
}
```

## API Reference

### BrokerWSClient

```javascript
class BrokerWSClient extends EventEmitter {
    constructor(apiKey, secretKey, wsURL)
}
```

#### Методы

##### constructor

```javascript
new BrokerWSClient(apiKey, secretKey, wsURL)
```

Создает новый WebSocket клиент.

**Параметры:**
- `apiKey` - API ключ
- `secretKey` - Секретный ключ 
- `wsURL` - URL WebSocket сервера

##### connect()

```javascript
async connect()
```

Устанавливает соединение и проходит аутентификацию.

**Возвращает:** `Promise<void>`

##### subscribe()

```javascript
async subscribe(channel, handler)
```

Подписывается на канал.

**Параметры:**
- `channel` - Название канала
- `handler` - Функция обработчик сообщений

**Возвращает:** `Promise<void>`

##### unsubscribe()

```javascript
async unsubscribe(channel)
```

Отменяет подписку на канал.

**Параметры:**
- `channel` - Название канала

**Возвращает:** `Promise<void>`

##### close()

```javascript
close()
```

Закрывает WebSocket соединение.

##### isConnected()

```javascript
isConnected()
```

Проверяет состояние соединения.

**Возвращает:** `boolean`

#### События

- `authenticated` - Когда клиент прошел аутентификацию
- `disconnected` - Когда соединение разорвано
- `error` - При возникновении ошибки 
- `subscribed` - При успешной подписке на канал
- `unsubscribed` - При отписке от канала

### Структура сообщений

```javascript
const message = {
    op: 'string',        // Операция (auth, subscribe, unsubscribe) 
    ch: 'string',        // Канал
    key: 'string',       // API ключ (только для аутентификации)
    ts: 1640995200000,   // Timestamp в миллисекундах
    nonce: 'string',     // Уникальный nonce
    sig: 'string',       // HMAC подпись
    data: {},            // Данные сообщения
    error: 'string'      // Описание ошибки
};
```

## Доступные каналы

### balances
Получение обновлений баланса пользователя.

**Пример данных:**
```javascript
{
  ch: 'balances',
  data: [
    {
      asset: 'USDT',
      total: '1000.00',
      locked: '0'
    }
  ]
}
```

### orders:{orderId}
Получение обновлений статуса конкретного ордера.

**Пример данных:**
```javascript
{
  ch: 'orders:ord_12345678',
  data: {
    status: 'FILLED',
    txHash: '0xabc123...',
    filledOut: '100.00',
    updatedAt: 1640995200000
  }
}
```

## Запуск примера

```bash
# Запуск WebSocket примера
node ws-example.js

# Или с ES модулями
node --experimental-modules ws-example.js
```

## Особенности

### Автоматическое переподключение

Клиент автоматически переподключается при разрыве соединения с экспоненциальной задержкой:

```javascript
// Настройки переподключения (по умолчанию)
client.maxReconnectAttempts = 5;
client.reconnectDelay = 1000; // 1 секунда
```

### Асинхронная обработка

Все методы возвращают промисы:

```javascript
try {
    await client.connect();
    await client.subscribe('balances', handler);
} catch (error) {
    console.error('Error:', error);
}
```

### Event-driven архитектура

Клиент наследует от EventEmitter для удобной обработки событий:

```javascript
client.on('error', (error) => {
    console.error('WebSocket error:', error);
});
```

## Безопасность

- Все сообщения подписываются с использованием HMAC-SHA256
- API ключи должны быть сохранены в безопасном месте
- Рекомендуется использовать переменные окружения для ключей

```javascript
const apiKey = process.env.BROKER_API_KEY;
const secretKey = process.env.BROKER_SECRET_KEY;
```

## Отладка

Включите дебаг логи для диагностики:

```javascript
// Установка уровня логирования
client.logger.setLevel('debug');

// Или прослушивание всех событий
client.on('response', (message) => {
    console.log('Server response:', message);
});
```

## Интеграция с существующими проектами

### Express.js

```javascript
import express from 'express';
import { BrokerWSClient } from './ws-client.js';

const app = express();
const wsClient = new BrokerWSClient(apiKey, secretKey, wsURL);

app.get('/status', (req, res) => {
    res.json({
        websocket_connected: wsClient.isConnected()
    });
});

// При старте сервера
wsClient.connect().then(() => {
    console.log('WebSocket client ready');
});
```

### Как middleware

```javascript
export function createWSMiddleware(apiKey, secretKey, wsURL) {
    const client = new BrokerWSClient(apiKey, secretKey, wsURL);
    
    return {
        async init() {
            await client.connect();
        },
        
        subscribe: client.subscribe.bind(client),
        unsubscribe: client.unsubscribe.bind(client),
        
        close() {
            client.close();
        }
    };
}
