# Broker Trading API - Python Client Example

Этот пример демонстрирует, как использовать Broker Trading API с помощью Python клиента.

## Установка

```bash
cd examples/client-py
pip install -r requirements.txt
```

Или с использованием виртуального окружения (рекомендуется):

```bash
cd examples/client-py
python -m venv venv
source venv/bin/activate  # На Windows: venv\Scripts\activate
pip install -r requirements.txt
```

## Описание

Клиент реализует все основные методы API:
- **GET /api/v1/balances** - получение балансов пользователя
- **POST /api/v1/estimate** - получение оценки свопа
- **POST /api/v1/swap** - выполнение свопа (с поддержкой идемпотентности)
- **GET /api/v1/orders/{id}/status** - получение статуса ордера

## Аутентификация

API использует HMAC-SHA256 аутентификацию со следующими заголовками:
- `X-API-KEY` - API ключ
- `X-API-TIMESTAMP` - временная метка в миллисекундах
- `X-API-NONCE` - уникальное значение nonce
- `X-API-SIGN` - HMAC-SHA256 подпись

### Формат подписи

Подпись создается из канонической строки:
```
<HTTP_METHOD>\n<PATH_WITH_QUERY>\n<TIMESTAMP_MS>\n<NONCE>\n<BODY_SHA256_HEX>
```

## Использование

### Быстрый старт

```python
from broker_client import BrokerClient, APIError

# Настройка клиента
api_key = "ak_NCkCBuPwz-76ZuIufZesX3CU0AEtZ_S3yAUDwJtBTsk"
secret_key = "tMLS4vFaKLApc8WL6qvn-3Gu11agkJe31ijrftHUMFYOjnXlUIkkc1sUzHqNSWjt"
base_url = "http://localhost:8080"

client = BrokerClient(api_key, secret_key, base_url)
```

### Получение балансов

```python
try:
    balances = client.get_balances()
    print(f"Найдено {len(balances)} активов:")
    for balance in balances:
        print(f"  {balance.asset}: {balance.total} (заблокировано: {balance.locked})")
except APIError as e:
    print(f"Ошибка API: {e}")
except Exception as e:
    print(f"Сетевая ошибка: {e}")
```

### Оценка свопа

```python
try:
    estimate = client.estimate_swap(
        from_asset="ETH",
        to_asset="USDT", 
        amount="1.0",
        network="ETH",  # опционально
        account="0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234"  # опционально
    )
    
    print("Оценка получена:")
    print(f"  Цена: {estimate['price']}")
    print(f"  Ожидаемый выход: {estimate['expectedOut']}")
    print(f"  Истекает в: {estimate['expiresAt']}")
    
    # Информация о маршруте
    for i, step in enumerate(estimate['route']):
        print(f"  Шаг {i+1}: {step['exchange']} - {step['fromAsset']} -> {step['toAsset']}")
        
except APIError as e:
    print(f"Ошибка API: {e}")
```

### Выполнение свопа

```python
import time

try:
    # Уникальный ключ идемпотентности
    idempotency_key = f"swap_{int(time.time() * 1000)}_{hash('unique_string') % 1000000}"
    
    swap_response = client.swap(
        from_asset="ETH",
        to_asset="USDT",
        amount="1.0",
        account="0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
        slippage_bps=30,
        idempotency_key=idempotency_key,
        client_order_id="my_order_123"  # опционально
    )
    
    print("Своп создан:")
    print(f"  Order ID: {swap_response['orderId']}")
    print(f"  Status: {swap_response['status']}")
    
except APIError as e:
    print(f"Ошибка API: {e}")
```

### Проверка статуса ордера

```python
try:
    order_status = client.get_order_status(
        order_id=swap_response['orderId'],
        client_order_id="my_order_123"  # опционально
    )
    
    print("Статус ордера:")
    print(f"  Order ID: {order_status['orderId']}")
    print(f"  Status: {order_status['status']}")
    print(f"  Filled Out: {order_status.get('filledOut', 'N/A')}")
    print(f"  TX Hash: {order_status.get('txHash', 'N/A')}")
    
except APIError as e:
    print(f"Ошибка API: {e}")
```

## Запуск примера

```bash
python example.py
```

Или:

```bash
python3 example.py
```

## Структура проекта

- `example.py` - основной файл с примером использования
- `broker_client.py` - класс клиента API с вспомогательными классами
- `requirements.txt` - зависимости Python
- `README.md` - данная документация

## Классы и исключения

### BrokerClient

Основной класс клиента со следующими методами:

- `get_balances()` → `List[Balance]`
- `estimate_swap(from_asset, to_asset, amount, network=None, account=None)` → `Dict`
- `swap(from_asset, to_asset, amount, account, slippage_bps, idempotency_key, client_order_id=None)` → `Dict`
- `get_order_status(order_id, client_order_id=None)` → `Dict`

### Balance

Класс для представления баланса актива:

```python
class Balance:
    def __init__(self, asset: str, total: str, locked: str):
        self.asset = asset      # Название актива
        self.total = total      # Общий баланс
        self.locked = locked    # Заблокированный баланс
```

### APIError

Исключение для ошибок API:

```python
class APIError(Exception):
    def __init__(self, code: str, message: str, request_id: str = ""):
        self.code = code           # Код ошибки
        self.message = message     # Сообщение об ошибке
        self.request_id = request_id  # ID запроса
```

## Обработка ошибок

Клиент автоматически обрабатывает ошибки API и HTTP:

```python
try:
    result = client.get_balances()
    print("Успех:", result)
except APIError as e:
    # Ошибка от API сервера
    print(f"API Error [{e.code}]: {e.message}")
    if e.request_id:
        print(f"Request ID: {e.request_id}")
except requests.RequestException as e:
    # Сетевая ошибка
    print(f"Network Error: {e}")
except Exception as e:
    # Другие ошибки
    print(f"Unexpected Error: {e}")
```

## Возможные статусы ордеров

- `PENDING` - ордер в обработке
- `FILLED` - ордер исполнен
- `PARTIAL` - ордер частично исполнен
- `CANCELED` - ордер отменен
- `FAILED` - ордер провален

## Особенности Python реализации

### Типизация
Код использует type hints для лучшей читаемости и IDE поддержки.

### Зависимости
- `requests` - для HTTP запросов
- Встроенные модули: `hashlib`, `hmac`, `json`, `time`, `random`, `base64`

### Совместимость
- Python 3.7+ (для поддержки type hints)
- Все основные операционные системы

## Безопасность

- Никогда не храните API ключи в коде
- Используйте переменные окружения для продакшена
- Следите за уникальностью nonce values
- Используйте HTTPS в продакшене

## Переменные окружения (рекомендуется)

Создайте файл `.env` или используйте переменные окружения:

```bash
export BROKER_API_KEY="your_api_key_here"
export BROKER_SECRET_KEY="your_secret_key_here"
export BROKER_BASE_URL="https://partner-api-dev.the-one.io"
```

И используйте в коде:

```python
import os

api_key = os.getenv('BROKER_API_KEY')
secret_key = os.getenv('BROKER_SECRET_KEY')
base_url = os.getenv('BROKER_BASE_URL')

if not all([api_key, secret_key, base_url]):
    raise ValueError("Отсутствуют необходимые переменные окружения")

client = BrokerClient(api_key, secret_key, base_url)
```

## Логирование

Для продакшена рекомендуется настроить логирование:

```python
import logging

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# В клиенте можно добавить логирование
try:
    result = client.get_balances()
    logger.info(f"Получены балансы: {len(result)} активов")
except APIError as e:
    logger.error(f"Ошибка API: {e}")
```

## Асинхронная версия (aiohttp)

Для асинхронного использования можно создать версию с `aiohttp`:

```python
import aiohttp
import asyncio

class AsyncBrokerClient:
    async def get_balances(self):
        async with aiohttp.ClientSession() as session:
            # Реализация с aiohttp
            pass
```

## Тестирование

Для тестирования можно использовать мокирование:

```python
import unittest
from unittest.mock import patch, MagicMock

class TestBrokerClient(unittest.TestCase):
    @patch('broker_client.requests.Session')
    def test_get_balances(self, mock_session):
        # Тестирование с мокированием
        pass
