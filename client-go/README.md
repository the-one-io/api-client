# Broker Trading API - Go Client Example

Этот пример демонстрирует, как использовать Broker Trading API с помощью Go клиента.

## Описание

Клиент реализует все основные методы API:
- **GET /balances** - получение балансов пользователя
- **POST /estimate** - получение оценки свопа
- **POST /swap** - выполнение свопа (с поддержкой идемпотентности)
- **GET /orders/{id}/status** - получение статуса ордера

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

### Настройка

```go
apiKey := "ak_NCkCBuPwz-76ZuIufZesX3CU0AEtZ_S3yAUDwJtBTsk"
secretKey := "tMLS4vFaKLApc8WL6qvn-3Gu11agkJe31ijrftHUMFYOjnXlUIkkc1sUzHqNSWjt"
baseURL := "http://localhost:8080"

httpClient := NewDefaultHTTPClient()
client := NewBrokerClient(apiKey, secretKey, baseURL, httpClient)
```

### Получение балансов

```go
balances, err := client.GetBalances(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Балансы: %+v\n", balances)
```

### Оценка свопа

```go
estimateReq := &EstimateRequest{
    From:        "ETH",
    To:          "USDT", 
    Amount:      "1.0",
    SlippageBps: 30,
    Network:     "ETH",
}

estimate, err := client.EstimateSwap(ctx, estimateReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Оценка: %+v\n", estimate)
```

### Выполнение свопа

```go
swapReq := &SwapRequest{
    EstimateID: estimate.ID, // Получен из предыдущей оценки
    From:       "ETH",
    To:         "USDT",
    Amount:     "1.0",
    Account:    "0x742d35Cc6634C0532925a3b8D0c79FA0Fa2d1234",
}

// Важно: используйте уникальный ключ идемпотентности для каждого свопа
idempotencyKey := fmt.Sprintf("swap_%d", time.Now().UnixNano())

swapResponse, err := client.Swap(ctx, swapReq, idempotencyKey)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Своп создан: %+v\n", swapResponse)
```

### Проверка статуса ордера

```go
orderStatus, err := client.GetOrderStatus(ctx, swapResponse.OrderID, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Статус ордера: %+v\n", orderStatus)
```

## Запуск примера

```bash
cd examples/client
go mod tidy
go run .
```

## Структура проекта

- `main.go` - основной файл с клиентом и примером использования
- `http_client.go` - HTTP клиент с поддержкой таймаутов
- `go.mod` - модуль Go

## Обработка ошибок

Клиент автоматически обрабатывает ошибки API и возвращает структурированные ошибки типа `APIError` с кодом, сообщением и ID запроса.

```go
if err != nil {
    if apiErr, ok := err.(*APIError); ok {
        fmt.Printf("API Error: %s (%s)\n", apiErr.Message, apiErr.Code)
    } else {
        fmt.Printf("Network/Other Error: %v\n", err)
    }
}
```

## Возможные статусы ордеров

- `PENDING` - ордер в обработке
- `FILLED` - ордер исполнен
- `PARTIAL` - ордер частично исполнен
- `CANCELED` - ордер отменен
- `FAILED` - ордер провален

## Безопасность

- Никогда не храните API ключи в коде
- Используйте переменные окружения для продакшена
- Следите за уникальностью nonce values
- Используйте HTTPS в продакшене
