# Integrat Go SDK

Go-клиент для платформы обмена данными [Integrat](https://github.com/plagness/Integrat).

## Установка

```bash
go get github.com/plagness/Integrat/sdk/go
```

## Quick Start — получение данных

```go
package main

import (
    "fmt"
    "log"

    integrat "github.com/plagness/Integrat/sdk/go"
)

func main() {
    client := integrat.New("itg_your_token")

    resp, err := client.Query("demo", "echo", map[string]any{"text": "hello"})
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(resp.Data)) // {"text":"hello"}
}
```

### Парсинг ответа в структуру

```go
var result struct {
    Echo struct {
        Text string `json:"text"`
    } `json:"echo"`
}
if err := resp.UnmarshalData(&result); err != nil {
    log.Fatal(err)
}
fmt.Println(result.Echo.Text) // hello
```

## Quick Start — публикация данных

```go
// 1. Создать плагин
plugin, err := client.CreatePlugin(integrat.CreatePluginParams{
    Name:    "My Weather API",
    Slug:    "my-weather",
    BaseURL: "https://api.example.com",
})

// 2. Добавить эндпоинт
_, err = client.CreateEndpoint(plugin.ID, integrat.CreateEndpointParams{
    Name:       "Current Weather",
    Slug:       "current",
    AccessTier: "open",
    DataType:   "basic",
    CacheTTL:   300,
    ProxyPath:  "/v1/weather/current",
    ProxyMethod: "GET",
})
```

## Все методы

| Метод | Описание |
|-------|----------|
| `Query(plugin, endpoint, params)` | Запрос данных (dev-режим) |
| `QueryInChat(plugin, endpoint, chatID, params)` | Запрос данных в контексте чата |
| `ListPlugins()` | Мои плагины |
| `CreatePlugin(params)` | Создать плагин |
| `GetPlugin(id)` | Получить плагин по ID |
| `UpdatePlugin(id, params)` | Обновить плагин |
| `DeletePlugin(id)` | Удалить плагин |
| `ListEndpoints(pluginID)` | Эндпоинты плагина |
| `CreateEndpoint(pluginID, params)` | Создать эндпоинт |
| `UpdateEndpoint(pluginID, epID, params)` | Обновить эндпоинт |
| `DeleteEndpoint(pluginID, epID)` | Удалить эндпоинт |
| `SearchMarketplace(params)` | Поиск в маркетплейсе |
| `GetPluginBySlug(slug)` | Детали плагина по slug |
| `Health()` | Проверка доступности API |

## Обработка ошибок

SDK возвращает типизированные ошибки — проверяйте через `errors.Is`:

```go
import "errors"

resp, err := client.Query("nonexistent", "endpoint", nil)
if err != nil {
    if errors.Is(err, integrat.ErrNotFound) {
        fmt.Println("Плагин не найден")
    } else if errors.Is(err, integrat.ErrUnauthorized) {
        fmt.Println("Неверный токен")
    } else if errors.Is(err, integrat.ErrProvider) {
        fmt.Println("Провайдер недоступен")
    } else {
        fmt.Println("Ошибка:", err)
    }
}

// Доступ к деталям ошибки
var apiErr *integrat.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("HTTP %d, code=%s, message=%s\n",
        apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

### Sentinel-ошибки

| Ошибка | HTTP | Описание |
|--------|------|----------|
| `ErrUnauthorized` | 401 | Неверный или отсутствующий токен |
| `ErrForbidden` | 403 | Нет доступа |
| `ErrNotFound` | 404 | Ресурс не найден |
| `ErrConflict` | 409 | Конфликт (например, лимит плагинов) |
| `ErrProvider` | 502-504 | Провайдер данных недоступен |

## Кастомный URL

```go
client := integrat.NewWithURL("itg_token", "http://localhost:30086")
```

## Документация

- [Integrat README](https://github.com/plagness/Integrat)
- [Спецификация integrat.yaml](https://github.com/plagness/Integrat/tree/main/spec)
- [Пример плагина](https://github.com/plagness/Integrat/tree/main/examples/channel-mcp)
