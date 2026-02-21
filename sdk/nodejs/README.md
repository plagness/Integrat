# @integrat/sdk

Node.js клиент для платформы обмена данными [Integrat](https://github.com/plagness/Integrat).

Включает TypeScript определения.

## Установка

```bash
npm install @integrat/sdk
```

## Quick Start — получение данных

```javascript
const { Integrat } = require('@integrat/sdk');

const client = new Integrat('itg_your_token');
const { data } = await client.query('demo', 'echo', { text: 'hello' });
console.log(data); // { text: "hello" }
```

## Quick Start — публикация данных

```javascript
// 1. Создать плагин
const plugin = await client.createPlugin({
  name: 'My Weather API',
  slug: 'my-weather',
  base_url: 'https://api.example.com',
});

// 2. Добавить эндпоинт
await client.createEndpoint(plugin.id, {
  name: 'Current Weather',
  slug: 'current',
  access_tier: 'open',
  data_type: 'basic',
  cache_ttl: 300,
  proxy_path: '/v1/weather/current',
  proxy_method: 'GET',
});
```

## Все методы

| Метод | Описание |
|-------|----------|
| `query(plugin, endpoint, params?, chatId?)` | Запрос данных |
| `listPlugins()` | Мои плагины |
| `createPlugin(params)` | Создать плагин |
| `getPlugin(id)` | Получить плагин по ID |
| `updatePlugin(id, params)` | Обновить плагин |
| `deletePlugin(id)` | Удалить плагин |
| `listEndpoints(pluginId)` | Эндпоинты плагина |
| `createEndpoint(pluginId, params)` | Создать эндпоинт |
| `updateEndpoint(pluginId, endpointId, params)` | Обновить эндпоинт |
| `deleteEndpoint(pluginId, endpointId)` | Удалить эндпоинт |
| `searchMarketplace(params?)` | Поиск в маркетплейсе |
| `getPluginBySlug(slug)` | Детали плагина по slug |
| `health()` | Проверка доступности API |

## Обработка ошибок

```javascript
const { Integrat, IntegratError } = require('@integrat/sdk');

try {
  await client.query('nonexistent', 'endpoint');
} catch (err) {
  if (err instanceof IntegratError) {
    if (err.isNotFound)      console.log('Плагин не найден');
    if (err.isUnauthorized)  console.log('Неверный токен');
    if (err.isForbidden)     console.log('Нет доступа');
    if (err.isProviderError) console.log('Провайдер недоступен');

    // Детали ошибки
    console.log(err.status); // 404
    console.log(err.code);   // "not_found"
    console.log(err.body);   // { error: "plugin not found", code: "not_found" }
  }
}
```

### Свойства IntegratError

| Свойство | Тип | Описание |
|----------|-----|----------|
| `status` | number | HTTP статус |
| `code` | string | Код ошибки (`unauthorized`, `not_found`, `forbidden`, `conflict`, `provider_unavailable`) |
| `body` | any | Тело ответа API |
| `isUnauthorized` | boolean | 401 |
| `isNotFound` | boolean | 404 |
| `isForbidden` | boolean | 403 |
| `isProviderError` | boolean | 502-504 |

## Кеш-флаги

```javascript
const result = await client.query('channel-mcp', 'messages.fetch', { limit: 10 });

console.log(result.cached); // true — ответ из кеша
console.log(result.stale);  // true — данные устарели (провайдер offline)
console.log(result.ttl);    // 120 — время жизни кеша (сек)
```

## Кастомный URL

```javascript
const client = new Integrat('itg_token', {
  baseURL: 'http://localhost:30086',
  timeout: 10000,
});
```

## TypeScript

Типы включены в пакет. Импортируйте интерфейсы:

```typescript
import { Integrat, IntegratError, Plugin, Endpoint, QueryResult } from '@integrat/sdk';
```

## Документация

- [Integrat README](https://github.com/plagness/Integrat)
- [Спецификация integrat.yaml](https://github.com/plagness/Integrat/tree/main/spec)
- [Пример плагина](https://github.com/plagness/Integrat/tree/main/examples/channel-mcp)
