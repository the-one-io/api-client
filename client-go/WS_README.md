# WebSocket Client for Go

WebSocket client для TheOne Trading API на языке Go с поддержкой аутентификации и реального времени.

## Установка

```bash
go mod tidy
```

## Зависимости

- `github.com/gorilla/websocket` - WebSocket библиотека для Go

## Использование

### Простое подключение

```go
package main

import (
    "log"
    "time"
)

func main() {
    // API ключи
    apiKey := "your_api_key"
    secretKey := "your_secret_key"
    wsURL := "wss://api.example.com/ws"
    
    // Создание клиента
    client := NewWSClient(apiKey, secretKey, wsURL)
    
    // Подключение
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Подписка на баланс
    err := client.Subscribe("balances", func(msg *WSMessage) {
        log.Printf("Balances update: %+v", msg.Data)
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Ожидание
    client.Wait()
}
```

### Подписка на каналы

#### Баланс пользователя

```go
client.Subscribe("balances", func(msg *WSMessage) {
    log.Printf("Received balances: %+v", msg.Data)
})
```

#### Статус ордера

```go
orderID := "ord_12345678"
channel := "orders:" + orderID

client.Subscribe(channel, func(msg *WSMessage) {
    log.Printf("Order %s status: %+v", orderID, msg.Data)
})
```

### Обработка ошибок

```go
client.Subscribe("balances", func(msg *WSMessage) {
    if msg.Error != "" {
        log.Printf("Error: %s", msg.Error)
        return
    }
    
    // Обработка данных
    log.Printf("Data: %+v", msg.Data)
})
```

## API Reference

### WSClient

```go
type WSClient struct {
    // Приватные поля
}
```

#### Методы соединения и подписок

##### NewWSClient

```go
func NewWSClient(apiKey, secretKey, wsURL string) *WSClient
```

Создает новый WebSocket клиент.

**Параметры:**
- `apiKey` - API ключ
- `secretKey` - Секретный ключ
- `wsURL` - URL WebSocket сервера

##### Connect

```go
func (ws *WSClient) Connect() error
```

Устанавливает соединение и проходит аутентификацию.

##### Subscribe

```go
func (ws *WSClient) Subscribe(channel string, handler MessageHandler) error
```

Подписывается на канал.

**Параметры:**
- `channel` - Название канала
- `handler` - Функция обработчик сообщений

##### Unsubscribe

```go
func (ws *WSClient) Unsubscribe(channel string) error
```

Отменяет подписку на канал.

##### Close

```go
func (ws *WSClient) Close()
```

Закрывает WebSocket соединение.

##### IsConnected

```go
func (ws *WSClient) IsConnected() bool
```

Проверяет состояние соединения.

#### REST API команды через WebSocket

##### EstimateSwap

```go
func (ws *WSClient) EstimateSwap(amountIn, assetIn, assetOut string) error
```

Оценивает обмен валют.

**Параметры:**
- `amountIn` - Количество входящей валюты
- `assetIn` - Входящая валюта
- `assetOut` - Исходящая валюта

##### DoSwap

```go
func (ws *WSClient) DoSwap(amountIn, assetIn, assetOut string) error
```

Выполняет обмен валют.

**Параметры:**
- `amountIn` - Количество входящей валюты
- `assetIn` - Входящая валюта
- `assetOut` - Исходящая валюта

##### GetOrderStatus

```go
func (ws *WSClient) GetOrderStatus(orderID string) error
```

Получает статус ордера.

**Параметры:**
- `orderID` - ID ордера

##### GetBalances

```go
func (ws *WSClient) GetBalances() error
```

Получает балансы аккаунта.

### WSMessage

```go
type WSMessage struct {
    Op        string      `json:"op,omitempty"`
    Channel   string      `json:"ch,omitempty"`
    Key       string      `json:"key,omitempty"`
    Timestamp int64       `json:"ts,omitempty"`
    Nonce     string      `json:"nonce,omitempty"`
    Signature string      `json:"sig,omitempty"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
}
```

## Доступные каналы

### balances
Получение обновлений баланса пользователя.

**Пример данных:**
```json
{
  "ch": "balances",
  "data": [
    {
      "asset": "USDT",
      "total": "1000.00",
      "locked": "0"
    }
  ]
}
```

### orders:{orderId}
Получение обновлений статуса конкретного ордера.

**Пример данных:**
```json
{
  "ch": "orders:ord_12345678",
  "data": {
    "status": "FILLED",
    "txHash": "0xabc123...",
    "filledOut": "100.00",
    "updatedAt": 1640995200000
  }
}
```

## Запуск примера

```bash
# Установка зависимостей
make deps

# Запуск WebSocket примера
make run-ws

# Или напрямую
go run ws_client.go ws_example.go

# Или сборка и запуск
go build -o ws_example ws_client.go ws_example.go
./ws_example
```

## Безопасность

- Все сообщения подписываются с использованием HMAC-SHA256
- API ключи должны быть сохранены в безопасном месте
- Рекомендуется использовать переменные окружения для ключей

## Особенности реализации

- Автоматическое переподключение с экспоненциальной задержкой
- Потокобезопасность
- Graceful shutdown с правильной очисткой ресурсов
- Обработка ошибок соединения
