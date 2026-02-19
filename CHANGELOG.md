# Changelog

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
