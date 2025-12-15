# WebSocket Client for Python

WebSocket client для Broker Trading API на Python с поддержкой асинхронной обработки, аутентификации и реального времени.

## Установка

```bash
pip install -r requirements.txt
```

## Зависимости

- `websockets` - Асинхронная WebSocket библиотека для Python
- `requests` - Для HTTP запросов (уже установлено для REST клиента)

## Использование

### Простое подключение

```python
import asyncio
from broker_ws_client import BrokerWSClient

async def main():
    # API ключи
    api_key = 'your_api_key'
    secret_key = 'your_secret_key'
    ws_url = 'wss://api.example.com/ws'
    
    # Создание и подключение клиента
    async with BrokerWSClient(api_key, secret_key, ws_url) as client:
        # Подписка на баланс
        await client.subscribe('balances', lambda msg: print(f"Balances: {msg['data']}"))
        
        # Ожидание
        await asyncio.sleep(60)  # Слушать 1 минуту

if __name__ == "__main__":
    asyncio.run(main())
```

### Подписка на каналы

#### Баланс пользователя

```python
async def handle_balances(message):
    print(f"Received balances: {message['data']}")

await client.subscribe('balances', handle_balances)
```

#### Статус ордера

```python
order_id = 'ord_12345678'
channel = f'orders:{order_id}'

async def handle_order_update(message):
    print(f"Order {order_id} status: {message['data']}")

await client.subscribe(channel, handle_order_update)
```

### Обработка ошибок

```python
async def handle_balances(message):
    if 'error' in message and message['error']:
        print(f"Error: {message['error']}")
        return
    
    # Обработка данных
    print(f"Data: {message['data']}")

try:
    async with BrokerWSClient(api_key, secret_key, ws_url) as client:
        await client.subscribe('balances', handle_balances)
        await asyncio.sleep(60)
except Exception as e:
    print(f"Connection failed: {e}")
```

### Ручное управление соединением

```python
client = BrokerWSClient(api_key, secret_key, ws_url)

try:
    await client.connect()
    await client.subscribe('balances', handle_balances)
    
    # Ваш код...
    
finally:
    await client.close()
```

## API Reference

### BrokerWSClient

```python
class BrokerWSClient:
    def __init__(self, api_key: str, secret_key: str, ws_url: str)
```

#### Методы

##### \_\_init\_\_

```python
def __init__(self, api_key: str, secret_key: str, ws_url: str)
```

Создает новый WebSocket клиент.

**Параметры:**
- `api_key` - API ключ
- `secret_key` - Секретный ключ
- `ws_url` - URL WebSocket сервера

##### connect()

```python
async def connect()
```

Устанавливает соединение и проходит аутентификацию.

**Возвращает:** `None`  
**Исключения:** `Exception` при ошибке подключения

##### subscribe()

```python
async def subscribe(channel: str, handler: Callable[[Dict[str, Any]], None])
```

Подписывается на канал.

**Параметры:**
- `channel` - Название канала
- `handler` - Функция обработчик сообщений (может быть sync или async)

**Возвращает:** `None`  
**Исключения:** `Exception` если не аутентифицирован

##### unsubscribe()

```python
async def unsubscribe(channel: str)
```

Отменяет подписку на канал.

**Параметры:**
- `channel` - Название канала

**Возвращает:** `None`

##### close()

```python
async def close()
```

Закрывает WebSocket соединение.

##### is_connected()

```python
def is_connected() -> bool
```

Проверяет состояние соединения.

**Возвращает:** `bool` - True если подключен и аутентифицирован

#### Context Manager

Клиент поддерживает async context manager:

```python
async with BrokerWSClient(api_key, secret_key, ws_url) as client:
    # Автоматическое подключение
    await client.subscribe('balances', handler)
    # Автоматическое отключение при выходе
```

### Обработчики сообщений

Функции-обработчики могут быть синхронными или асинхронными:

```python
# Синхронный обработчик
def sync_handler(message):
    print("Sync handler:", message['data'])

# Асинхронный обработчик  
async def async_handler(message):
    await some_async_operation()
    print("Async handler:", message['data'])

await client.subscribe('balances', sync_handler)
await client.subscribe('orders:123', async_handler)
```

### Структура сообщений

```python
message = {
    'op': 'string',          # Операция (auth, subscribe, unsubscribe)
    'ch': 'string',          # Канал  
    'key': 'string',         # API ключ (только для аутентификации)
    'ts': 1640995200000,     # Timestamp в миллисекундах
    'nonce': 'string',       # Уникальный nonce
    'sig': 'string',         # HMAC подпись
    'data': {},              # Данные сообщения
    'error': 'string'        # Описание ошибки
}
```

## Доступные каналы

### balances
Получение обновлений баланса пользователя.

**Пример данных:**
```python
{
  'ch': 'balances',
  'data': [
    {
      'asset': 'USDT',
      'total': '1000.00', 
      'locked': '0'
    }
  ]
}
```

### orders:{orderId}
Получение обновлений статуса конкретного ордера.

**Пример данных:**
```python
{
  'ch': 'orders:ord_12345678',
  'data': {
    'status': 'FILLED',
    'txHash': '0xabc123...',
    'filledOut': '100.00',
    'updatedAt': 1640995200000
  }
}
```

## Запуск примера

```bash
# Запуск WebSocket примера
python ws_example.py

# С виртуальным окружением
python -m venv venv
source venv/bin/activate  # или venv\Scripts\activate на Windows
pip install -r requirements.txt
python ws_example.py
```

## Особенности

### Автоматическое переподключение

Клиент автоматически переподключается при разрыве соединения с экспоненциальной задержкой:

```python
# Настройки переподключения (по умолчанию)
client.max_reconnect_attempts = 5
client.reconnect_delay = 1.0  # 1 секунда
```

При переподключении автоматически восстанавливаются все подписки.

### Логирование

Клиент использует стандартную библиотеку `logging`:

```python
import logging

# Включение debug логов
logging.basicConfig(level=logging.DEBUG)

# Или для конкретного клиента
client = BrokerWSClient(api_key, secret_key, ws_url)
client.logger.setLevel(logging.DEBUG)
```

### Асинхронная архитектура

Клиент полностью асинхронный и использует `asyncio`:

```python
# Все операции асинхронные
await client.connect()
await client.subscribe('balances', handler)
await client.unsubscribe('balances')
await client.close()

# Поддержка нескольких клиентов
clients = [
    BrokerWSClient(key1, secret1, url),
    BrokerWSClient(key2, secret2, url)
]

await asyncio.gather(*[client.connect() for client in clients])
```

## Интеграция с существующими проектами

### FastAPI

```python
from fastapi import FastAPI
from broker_ws_client import BrokerWSClient

app = FastAPI()
ws_client = None

@app.on_event("startup")
async def startup():
    global ws_client
    ws_client = BrokerWSClient(api_key, secret_key, ws_url)
    await ws_client.connect()

@app.on_event("shutdown") 
async def shutdown():
    if ws_client:
        await ws_client.close()

@app.get("/status")
async def get_status():
    return {"websocket_connected": ws_client.is_connected() if ws_client else False}
```

### Django (с django-channels)

```python
from channels.generic.websocket import AsyncWebsocketConsumer
from broker_ws_client import BrokerWSClient

class TradingConsumer(AsyncWebsocketConsumer):
    async def connect(self):
        self.ws_client = BrokerWSClient(api_key, secret_key, ws_url)
        await self.ws_client.connect()
        await self.accept()
    
    async def disconnect(self, close_code):
        if hasattr(self, 'ws_client'):
            await self.ws_client.close()
```

### Как сервис

```python
class WebSocketService:
    def __init__(self, api_key, secret_key, ws_url):
        self.client = BrokerWSClient(api_key, secret_key, ws_url)
        self.handlers = {}
    
    async def start(self):
        await self.client.connect()
    
    async def stop(self):
        await self.client.close()
    
    async def add_subscription(self, channel, callback):
        self.handlers[channel] = callback
        await self.client.subscribe(channel, callback)
    
    async def remove_subscription(self, channel):
        if channel in self.handlers:
            await self.client.unsubscribe(channel)
            del self.handlers[channel]
```

## Безопасность

- Все сообщения подписываются с использованием HMAC-SHA256
- API ключи должны быть сохранены в безопасном месте
- Рекомендуется использовать переменные окружения для ключей

```python
import os

api_key = os.getenv('BROKER_API_KEY')
secret_key = os.getenv('BROKER_SECRET_KEY')

# Или с python-dotenv
from dotenv import load_dotenv
load_dotenv()

api_key = os.getenv('BROKER_API_KEY')
```

## Отладка

### Включение debug логов

```python
import logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
```

### Мониторинг состояния

```python
async def monitor_connection():
    while True:
        print(f"Connected: {client.is_connected()}")
        await asyncio.sleep(10)

# Запуск в фоне
asyncio.create_task(monitor_connection())
```

### Обработка всех сообщений

```python
async def debug_handler(message):
    print(f"DEBUG: Received message: {message}")

# Подписаться на все каналы для отладки
await client.subscribe('balances', debug_handler)
await client.subscribe('orders:test_order', debug_handler)
```

## Требования

- Python 3.7+
- `asyncio` поддержка
- `websockets` библиотека
- Стабильное интернет соединение

## Ограничения

- Максимум 100 активных подписок на одно соединение
- Таймаут аутентификации: 10 секунд  
- Максимальный размер сообщения: 64KB
- Heartbeat интервал: 30 секунд
