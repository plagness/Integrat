# Changelog

## [2026.02.2] - 2026-02-21

- **Типизированные ошибки:**
  - Go: sentinel-ошибки (`ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrProvider`) + `APIError` с `errors.Is` поддержкой.
  - Node.js: `IntegratError.code`, геттеры `isNotFound`, `isUnauthorized`, `isForbidden`, `isProviderError`.
- **Publisher API — 10 новых методов** (идентичны в Go и Node.js):
  - `CreatePlugin`, `GetPlugin`, `UpdatePlugin`, `DeletePlugin`.
  - `ListEndpoints`, `CreateEndpoint`, `UpdateEndpoint`, `DeleteEndpoint`.
  - `SearchMarketplace`, `GetPluginBySlug`.
- **Go SDK:** `doRequest`/`doJSON` — централизованный HTTP + auth + error handling; `UnmarshalData` на `QueryResponse`.
- **Node.js SDK:** `_json` — внутренний fetch-хелпер с автоматической обработкой ошибок.
- **JSON Schema** (`spec/integrat.schema.json`) — draft-07 схема для `integrat.yaml` с IDE-автодополнением.
- **README** для обоих SDK (`sdk/go/README.md`, `sdk/nodejs/README.md`) — Quick Start, все методы, обработка ошибок.
- **CI:** добавлен `publish-npm` job (публикация `@integrat/sdk` в npm при создании тега).

## [2026.02.1] - 2026-02-18

- Начальный релиз SDK для платформы Integrat:
  - Go SDK (`sdk/go/integrat.go`) — `Query`, `ListPlugins`, `Health`.
  - Node.js SDK (`sdk/nodejs/index.js`) — `Integrat` класс с `query`, `listPlugins`, `health`.
  - TypeScript определения (`sdk/nodejs/index.d.ts`).
- Спецификация `integrat.yaml` для описания плагинов:
  - `plugin`, `provider`, `endpoints[]`, `config_fields[]`.
  - Три уровня доступа: `open`, `gated`, `private`.
  - Типы данных: `basic`, `medium`, `complex`.
- Пример плагина: `examples/channel-mcp/integrat.yaml`.
- Governance-файлы: `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`.
- CI: Go build, markdown link validation.
