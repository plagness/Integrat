# Integrat

Платформа обмена данными между Telegram-разработчиками.

Integrat — SDK и спецификация для подключения ваших сервисов к единому маркетплейсу данных.
Вы поднимаете свой сервер с данными, описываете его через `integrat.yaml`, а платформа
берёт на себя авторизацию, кеширование, контроль доступа и биллинг через Telegram Stars.

## Быстрый старт

### 1. Опишите плагин

Создайте `integrat.yaml` рядом с вашим проектом:

```yaml
plugin:
  slug: my-plugin
  name: Мой плагин
  description: Данные из моего сервиса
  version: 1.0.0
  homepage: https://github.com/you/my-plugin

provider:
  base_url: https://my-server.example.com
  health_path: /health
  auth:
    type: bearer
    env: MY_API_TOKEN

endpoints:
  - slug: users-online
    name: Онлайн пользователи
    description: Текущее количество онлайн пользователей
    path: /tools/users-online
    method: POST
    access: open
    cache_ttl: 60
    data_type: basic

  - slug: analytics
    name: Аналитика
    description: Детальная аналитика канала
    path: /tools/analytics
    method: POST
    access: gated
    cache_ttl: 300
    data_type: medium

config_fields:
  - slug: channel_id
    label: ID канала
    type: string
    required: true
    placeholder: "@mychannel"
    help: Юзернейм или числовой ID канала для мониторинга
```

### 2. Установите SDK

**Go:**
```bash
go get github.com/plagness/Integrat/sdk/go
```

**Node.js:**
```bash
npm install @integrat/sdk
```

### 3. Используйте SDK для запросов

**Go:**
```go
client := integrat.New("itg_your_api_token")

resp, err := client.Query("channel-mcp", "messages.fetch", map[string]any{
    "channel": "durov",
    "limit":   10,
})
```

**Node.js:**
```js
import { Integrat } from '@integrat/sdk';

const client = new Integrat('itg_your_api_token');

const resp = await client.query('channel-mcp', 'messages.fetch', {
  channel: 'durov',
  limit: 10,
});
```

## Спецификация integrat.yaml

### plugin

| Поле | Тип | Обязательное | Описание |
|------|-----|:---:|----------|
| `slug` | string | да | Уникальный идентификатор (a-z, 0-9, дефис) |
| `name` | string | да | Отображаемое имя |
| `description` | string | да | Краткое описание |
| `version` | string | да | Версия плагина (semver или calver) |
| `homepage` | string | нет | Ссылка на репозиторий или сайт |
| `icon` | string | нет | URL иконки (64x64, PNG/SVG) |

### provider

| Поле | Тип | Обязательное | Описание |
|------|-----|:---:|----------|
| `base_url` | string | да | Базовый URL вашего сервера |
| `health_path` | string | нет | Путь для health-check (по умолчанию `/health`) |
| `auth.type` | string | нет | Тип авторизации: `bearer`, `header`, `none` |
| `auth.env` | string | нет | Имя переменной окружения с токеном |
| `auth.header` | string | нет | Имя заголовка (для type=header) |

### endpoints[]

| Поле | Тип | Обязательное | Описание |
|------|-----|:---:|----------|
| `slug` | string | да | Идентификатор эндпоинта |
| `name` | string | да | Отображаемое имя |
| `description` | string | нет | Описание |
| `path` | string | да | Путь на сервере провайдера |
| `method` | string | нет | HTTP метод (по умолчанию `POST`) |
| `access` | string | да | Уровень доступа: `open`, `gated`, `private` |
| `cache_ttl` | int | нет | Время кеширования в секундах (0 = без кеша) |
| `data_type` | string | нет | Тип данных: `basic`, `medium`, `complex` |
| `params_schema` | object | нет | JSON Schema параметров запроса |

### config_fields[]

Поля конфигурации, которые пользователь заполняет в Mini App при подключении плагина к чату.

| Поле | Тип | Обязательное | Описание |
|------|-----|:---:|----------|
| `slug` | string | да | Идентификатор поля |
| `label` | string | да | Подпись в интерфейсе |
| `type` | string | да | Тип: `string`, `number`, `boolean`, `select` |
| `required` | bool | нет | Обязательное поле (по умолчанию `false`) |
| `default` | any | нет | Значение по умолчанию |
| `placeholder` | string | нет | Подсказка в поле ввода |
| `help` | string | нет | Пояснительный текст |
| `options` | array | нет | Варианты для type=select: `[{value, label}]` |

## Уровни доступа

| Уровень | Описание |
|---------|----------|
| **open** | Доступен всем участникам чата, где подключён плагин |
| **gated** | Требует одобрения владельца плагина |
| **private** | Доступен только владельцу и явно указанным пользователям |

## Типы данных

| Тип | Описание | Пример |
|-----|----------|--------|
| **basic** | Простые значения, счётчики | Онлайн, статус |
| **medium** | Структурированные данные | Списки, таблицы |
| **complex** | Тяжёлые данные, аналитика | Графики, эмбеддинги |

## API

Базовый URL: `https://integrat.plag.space` (production)

### Аутентификация

Получите API-токен в Mini App (Telegram → @IntegratBot).
Передавайте его в заголовке:

```
Authorization: Bearer itg_xxxxxxxxxxxx
```

### Основные эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Health check |
| GET | `/version` | Версия API |
| POST | `/v1/auth/token` | Получить/создать токен |
| GET | `/v1/plugins` | Список плагинов |
| POST | `/v1/plugins` | Зарегистрировать плагин |
| GET | `/v1/plugins/:id` | Информация о плагине |
| PUT | `/v1/plugins/:id` | Обновить плагин |
| DELETE | `/v1/plugins/:id` | Удалить плагин |
| GET | `/v1/plugins/:id/endpoints` | Эндпоинты плагина |
| POST | `/v1/plugins/:id/endpoints` | Добавить эндпоинт |
| GET | `/v1/plugins/:id/config` | Конфигурация плагина |
| PUT | `/v1/plugins/:id/config` | Сохранить конфигурацию |
| POST | `/v1/query` | Запрос данных через прокси |

### POST /v1/query

Основной метод для получения данных. Платформа проксирует запрос к провайдеру
с проверкой доступа, кешированием и аудитом.

```json
{
  "plugin": "channel-mcp",
  "endpoint": "messages.fetch",
  "chat_id": 123456789,
  "params": {
    "channel": "durov",
    "limit": 10
  }
}
```

**Заголовки ответа:**

| Заголовок | Описание |
|-----------|----------|
| `X-Integrat-Cached` | `true` если ответ из кеша |
| `X-Integrat-TTL` | Оставшееся время жизни кеша (сек) |
| `X-Integrat-Stale` | `true` если данные устарели (провайдер offline) |

## Лицензия

MIT
