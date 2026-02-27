package validator

import (
	"strings"
	"testing"
)

// ── Helpers ─────────────────────────────────────────────────────────────

func mustParse(t *testing.T, yaml string) *Spec {
	t.Helper()
	spec, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return spec
}

func hasError(r *Result, substr string) bool {
	for _, e := range r.Errors {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

func hasWarning(r *Result, substr string) bool {
	for _, e := range r.Warnings {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

// ── Минимальный валидный YAML ───────────────────────────────────────────

const validMinimal = `
plugin:
  slug: test-plugin
  name: Test Plugin
  description: Тестовый плагин
  version: "1.0.0"

provider:
  base_url: http://localhost:8080

endpoints:
  - slug: test.endpoint
    name: Test Endpoint
    path: /v1/test
    method: POST
    access: open
    data_type: basic
`

// ── Parse ───────────────────────────────────────────────────────────────

func TestParse_ValidYAML(t *testing.T) {
	spec, err := Parse([]byte(validMinimal))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Plugin.Slug != "test-plugin" {
		t.Errorf("plugin.slug = %q, want %q", spec.Plugin.Slug, "test-plugin")
	}
	if len(spec.Endpoints) != 1 {
		t.Errorf("endpoints count = %d, want 1", len(spec.Endpoints))
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte("{{invalid yaml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_EmptyInput(t *testing.T) {
	spec, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := Validate(spec)
	if r.OK() {
		t.Error("expected errors for empty spec")
	}
}

// ── ValidateBytes ───────────────────────────────────────────────────────

func TestValidateBytes_Valid(t *testing.T) {
	spec, r := ValidateBytes([]byte(validMinimal))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("unexpected errors: %v", r.Errors)
	}
}

func TestValidateBytes_InvalidYAML(t *testing.T) {
	_, r := ValidateBytes([]byte("{{bad"))
	if r.OK() {
		t.Error("expected errors for invalid YAML")
	}
}

// ── Plugin ──────────────────────────────────────────────────────────────

func TestValidatePlugin_MissingSlug(t *testing.T) {
	spec := mustParse(t, `
plugin:
  name: X
  description: Y
  version: "1.0"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "plugin.slug") {
		t.Errorf("expected plugin.slug error, got: %v", r.Errors)
	}
}

func TestValidatePlugin_InvalidSlugFormat(t *testing.T) {
	tests := []struct {
		slug string
		ok   bool
	}{
		{"channel-mcp", true},
		{"bcs-mcp", true},
		{"democracy", true},
		{"llm-mcp", true},
		{"metrics-api", true},
		{"a", true},           // одиночный символ
		{"test.plugin", true}, // точка допустима
		{"test_plug", true},   // подчёркивание допустимо
		{"3d-plugin", true},   // цифра в начале
		{"", false},           // пустой
		{"Test-Plugin", false},   // заглавные
		{"-bad-start", false},    // дефис в начале
		{"плагин", false},        // кириллица
		{"bad slug", false},      // пробел
		{"bad/slug", false},      // слеш
	}
	for _, tt := range tests {
		ok := slugRe.MatchString(tt.slug)
		if ok != tt.ok {
			t.Errorf("slugRe.Match(%q) = %v, want %v", tt.slug, ok, tt.ok)
		}
	}
}

func TestValidatePlugin_MissingName(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  description: Y
  version: "1.0"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "plugin.name") {
		t.Errorf("expected plugin.name error, got: %v", r.Errors)
	}
}

func TestValidatePlugin_MissingDescription(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  version: "1.0"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "plugin.description") {
		t.Errorf("expected plugin.description error, got: %v", r.Errors)
	}
}

func TestValidatePlugin_MissingVersion(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: Desc
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "plugin.version") {
		t.Errorf("expected plugin.version error, got: %v", r.Errors)
	}
}

// ── Provider ────────────────────────────────────────────────────────────

func TestValidateProvider_MissingBaseURL(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider: {}
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "provider.base_url") {
		t.Errorf("expected provider.base_url error, got: %v", r.Errors)
	}
}

func TestValidateProvider_InvalidAuthType(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider:
  base_url: http://x
  auth:
    type: oauth2
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "provider.auth.type") {
		t.Errorf("expected auth.type error, got: %v", r.Errors)
	}
}

func TestValidateProvider_AuthHeaderWarning(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider:
  base_url: http://x
  auth:
    type: header
    env: TOKEN
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !hasWarning(r, "header не указан") {
		t.Errorf("expected auth header warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateProvider_ValidAuth(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider:
  base_url: http://x
  auth:
    type: bearer
    env: MY_TOKEN
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if r.OK() != true {
		t.Errorf("unexpected errors: %v", r.Errors)
	}
}

// ── Endpoints ───────────────────────────────────────────────────────────

func TestValidateEndpoints_Empty(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider:
  base_url: http://x
endpoints: []
`)
	r := Validate(spec)
	if !hasError(r, "минимум 1 эндпоинт") {
		t.Errorf("expected endpoints empty error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_MissingRequired(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: Test
  description: D
  version: "1.0"
provider:
  base_url: http://x
endpoints:
  - description: "без обязательных полей"
`)
	r := Validate(spec)
	if !hasError(r, "endpoints[0].slug") {
		t.Errorf("expected slug error, got: %v", r.Errors)
	}
	if !hasError(r, "endpoints[0].name") {
		t.Errorf("expected name error, got: %v", r.Errors)
	}
	if !hasError(r, "endpoints[0].path") {
		t.Errorf("expected path error, got: %v", r.Errors)
	}
	if !hasError(r, "endpoints[0].access") {
		t.Errorf("expected access error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_InvalidMethod(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    method: PATCH
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "method") {
		t.Errorf("expected method error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_InvalidAccess(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: public
`)
	r := Validate(spec)
	if !hasError(r, "access") {
		t.Errorf("expected access error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_InvalidDataType(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
    data_type: huge
`)
	r := Validate(spec)
	if !hasError(r, "data_type") {
		t.Errorf("expected data_type error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_DuplicateSlug(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: same.slug
    name: A
    path: /a
    access: open
  - slug: same.slug
    name: B
    path: /b
    access: gated
`)
	r := Validate(spec)
	if !hasError(r, "дубликат") {
		t.Errorf("expected duplicate slug error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_PathWithoutSlash(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: v1/test
    access: open
`)
	r := Validate(spec)
	if !hasError(r, "начинаться с /") {
		t.Errorf("expected path error, got: %v", r.Errors)
	}
}

func TestValidateEndpoints_NegativeCacheTTL(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
    cache_ttl: -1
`)
	r := Validate(spec)
	if !hasError(r, "cache_ttl") {
		t.Errorf("expected cache_ttl error, got: %v", r.Errors)
	}
}

// ── params_schema ───────────────────────────────────────────────────────

func TestValidateParamsSchema_Valid(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
    params_schema:
      type: object
      properties:
        query:
          type: string
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("unexpected errors: %v", r.Errors)
	}
}

func TestValidateParamsSchema_MissingType(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
    params_schema:
      properties:
        query:
          type: string
`)
	r := Validate(spec)
	if !hasError(r, "params_schema") && !hasError(r, "type") {
		t.Errorf("expected params_schema type error, got: %v", r.Errors)
	}
}

func TestValidateParamsSchema_WrongType(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
    params_schema:
      type: array
`)
	r := Validate(spec)
	if !hasError(r, "\"object\"") {
		t.Errorf("expected params_schema type=object error, got: %v", r.Errors)
	}
}

// ── config_fields ───────────────────────────────────────────────────────

func TestValidateConfigFields_Valid(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - slug: api_key
    label: API Key
    type: string
    required: true
  - slug: mode
    label: Mode
    type: select
    options:
      - value: fast
        label: Fast
      - value: slow
        label: Slow
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("unexpected errors: %v", r.Errors)
	}
}

func TestValidateConfigFields_MissingRequired(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - help: Подсказка без slug/label/type
`)
	r := Validate(spec)
	if !hasError(r, "config_fields[0].slug") {
		t.Errorf("expected slug error, got: %v", r.Errors)
	}
	if !hasError(r, "config_fields[0].label") {
		t.Errorf("expected label error, got: %v", r.Errors)
	}
	if !hasError(r, "config_fields[0].type") {
		t.Errorf("expected type error, got: %v", r.Errors)
	}
}

func TestValidateConfigFields_InvalidType(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - slug: x
    label: X
    type: textarea
`)
	r := Validate(spec)
	if !hasError(r, "config_fields[0].type") {
		t.Errorf("expected type error, got: %v", r.Errors)
	}
}

func TestValidateConfigFields_DuplicateSlug(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - slug: dup
    label: First
    type: string
  - slug: dup
    label: Second
    type: number
`)
	r := Validate(spec)
	if !hasError(r, "дубликат") {
		t.Errorf("expected duplicate config slug error, got: %v", r.Errors)
	}
}

func TestValidateConfigFields_SelectWithoutOptions(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - slug: mode
    label: Mode
    type: select
`)
	r := Validate(spec)
	if !hasWarning(r, "options не указаны") {
		t.Errorf("expected select warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateConfigFields_OptionMissingValue(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
config_fields:
  - slug: mode
    label: Mode
    type: select
    options:
      - label: Only Label
`)
	r := Validate(spec)
	if !hasError(r, "options[0].value") {
		t.Errorf("expected option value error, got: %v", r.Errors)
	}
}

// ── Полные интеграционные тесты с реальными YAML ────────────────────────

const channelMCPYaml = `
plugin:
  slug: channel-mcp
  name: Channel Analytics
  description: Аналитика Telegram-каналов
  version: 2026.02.9
  homepage: https://github.com/plagness/Channel-MCP

provider:
  base_url: ${CHANNEL_MCP_URL}
  health_path: /health
  auth:
    type: bearer
    env: MCP_HTTP_TOKEN

endpoints:
  - slug: channels.list
    name: Список каналов
    path: /tools/channels.list
    method: POST
    access: open
    cache_ttl: 300
    data_type: basic

  - slug: messages.fetch
    name: Сообщения
    path: /tools/messages.fetch
    method: POST
    access: open
    cache_ttl: 120
    data_type: medium
    params_schema:
      type: object
      properties:
        channel:
          type: string
        limit:
          type: integer
          minimum: 1
          maximum: 500

  - slug: tags.top
    name: Топ тегов
    path: /tools/tags.top
    method: POST
    access: open
    cache_ttl: 600
    data_type: basic

  - slug: messages.search
    name: Семантический поиск
    path: /tools/messages.search
    method: POST
    access: gated
    cache_ttl: 60
    data_type: complex
    params_schema:
      type: object
      required:
        - query
      properties:
        query:
          type: string

config_fields:
  - slug: channels_json
    label: Каналы для мониторинга
    type: string
    required: true
  - slug: backfill_days
    label: Глубина загрузки
    type: number
    required: false
    default: 7
`

func TestValidate_ChannelMCP_Full(t *testing.T) {
	spec, r := ValidateBytes([]byte(channelMCPYaml))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("channel-mcp validation failed: %v", r.Errors)
	}
	if spec.Plugin.Slug != "channel-mcp" {
		t.Errorf("slug = %q", spec.Plugin.Slug)
	}
	if len(spec.Endpoints) != 4 {
		t.Errorf("endpoints = %d, want 4", len(spec.Endpoints))
	}
	if len(spec.ConfigFields) != 2 {
		t.Errorf("config_fields = %d, want 2", len(spec.ConfigFields))
	}
}

const democracyYaml = `
plugin:
  slug: democracy
  name: Democracy
  description: Модуль управления сообществами
  version: "2026.02.1"
  icon: "🏛"

provider:
  base_url: http://democracycore:8087
  health_path: /health
  auth:
    type: header
    env: X_INIT_DATA
    header: X-Init-Data

endpoints:
  - slug: governance.regime
    name: Режим правления
    path: /v1/governance/{chat_id}
    method: GET
    access: gated
    cache_ttl: 60
    data_type: basic
    params_schema:
      type: object
      properties:
        chat_id:
          type: integer
      required: [chat_id]

  - slug: proposals.list
    name: Список предложений
    path: /v1/proposals
    method: GET
    access: gated
    cache_ttl: 30
    data_type: medium

  - slug: citizens.list
    name: Граждане чата
    path: /v1/citizens/{chat_id}
    method: GET
    access: gated
    cache_ttl: 60
    data_type: medium

config_fields:
  - slug: default_regime
    label: Режим по умолчанию
    type: select
    options:
      - value: democracy
        label: Прямая демократия
      - value: autocracy
        label: Автократия
`

func TestValidate_Democracy_Full(t *testing.T) {
	spec, r := ValidateBytes([]byte(democracyYaml))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("democracy validation failed: %v", r.Errors)
	}
	if spec.Plugin.Slug != "democracy" {
		t.Errorf("slug = %q", spec.Plugin.Slug)
	}
	if len(spec.Endpoints) != 3 {
		t.Errorf("endpoints = %d, want 3", len(spec.Endpoints))
	}
}

// ── Множественные ошибки ────────────────────────────────────────────────

func TestValidate_MultipleErrors(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: ""
provider: {}
endpoints: []
`)
	r := Validate(spec)
	if r.OK() {
		t.Fatal("expected errors for broken spec")
	}
	if len(r.Errors) < 4 {
		t.Errorf("expected >= 4 errors (slug, name, desc, version, base_url, endpoints), got %d: %v", len(r.Errors), r.Errors)
	}
}

func TestValidate_NoConfigFields_OK(t *testing.T) {
	_, r := ValidateBytes([]byte(validMinimal))
	if !r.OK() {
		t.Errorf("minimal valid yaml should pass: %v", r.Errors)
	}
}

// ── Граничные случаи ────────────────────────────────────────────────────

func TestValidate_ZeroCacheTTL_OK(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: a
    name: A
    path: /a
    access: private
    cache_ttl: 0
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("cache_ttl=0 should be valid: %v", r.Errors)
	}
}

func TestValidate_EnvVarInBaseURL_OK(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: ${MY_SERVICE_URL}
endpoints:
  - slug: a
    name: A
    path: /a
    access: open
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("env var in base_url should be valid: %v", r.Errors)
	}
}

func TestValidate_AllAccessTypes(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: ep.open
    name: Open
    path: /open
    access: open
  - slug: ep.gated
    name: Gated
    path: /gated
    access: gated
  - slug: ep.private
    name: Private
    path: /private
    access: private
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("all access types should be valid: %v", r.Errors)
	}
}

func TestValidate_AllMethods(t *testing.T) {
	spec := mustParse(t, `
plugin:
  slug: test
  name: T
  description: D
  version: "1"
provider:
  base_url: http://x
endpoints:
  - slug: ep.get
    name: Get
    path: /get
    method: GET
    access: open
  - slug: ep.post
    name: Post
    path: /post
    method: POST
    access: open
  - slug: ep.put
    name: Put
    path: /put
    method: PUT
    access: open
  - slug: ep.delete
    name: Delete
    path: /delete
    method: DELETE
    access: open
`)
	r := Validate(spec)
	if !r.OK() {
		t.Errorf("all methods should be valid: %v", r.Errors)
	}
}
