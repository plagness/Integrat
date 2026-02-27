// Пакет validator — валидация integrat.yaml спецификаций.
// Проверяет структуру, обязательные поля, форматы slug, enum-значения,
// уникальность идентификаторов и корректность params_schema.
package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ── Типы integrat.yaml ──────────────────────────────────────────────────

// Spec — корневая структура integrat.yaml.
type Spec struct {
	Plugin       PluginDef        `yaml:"plugin"`
	Provider     ProviderDef      `yaml:"provider"`
	Endpoints    []EndpointDef    `yaml:"endpoints"`
	ConfigFields []ConfigFieldDef `yaml:"config_fields"`
}

// PluginDef — секция plugin.
type PluginDef struct {
	Slug        string `yaml:"slug"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Homepage    string `yaml:"homepage"`
	Icon        string `yaml:"icon"`
}

// ProviderDef — секция provider.
type ProviderDef struct {
	BaseURL    string   `yaml:"base_url"`
	HealthPath string   `yaml:"health_path"`
	ProxyMode  string   `yaml:"proxy_mode"`
	Auth       *AuthDef `yaml:"auth"`
}

// AuthDef — настройки аутентификации провайдера.
type AuthDef struct {
	Type   string `yaml:"type"`
	Env    string `yaml:"env"`
	Header string `yaml:"header"`
}

// EndpointDef — определение одного эндпоинта.
type EndpointDef struct {
	Slug         string    `yaml:"slug"`
	Name         string    `yaml:"name"`
	Description  string    `yaml:"description"`
	Path         string    `yaml:"path"`
	Method       string    `yaml:"method"`
	Access       string    `yaml:"access"`
	CacheTTL     *int      `yaml:"cache_ttl"`
	DataType     string    `yaml:"data_type"`
	ParamsSchema yaml.Node `yaml:"params_schema"`
}

// ConfigFieldDef — определение поля конфигурации.
type ConfigFieldDef struct {
	Slug        string            `yaml:"slug"`
	Label       string            `yaml:"label"`
	Type        string            `yaml:"type"`
	Required    bool              `yaml:"required"`
	Default     any               `yaml:"default"`
	Placeholder string            `yaml:"placeholder"`
	Help        string            `yaml:"help"`
	Options     []ConfigOptionDef `yaml:"options"`
}

// ConfigOptionDef — вариант для type: select.
type ConfigOptionDef struct {
	Value string `yaml:"value"`
	Label string `yaml:"label"`
}

// ── Результат валидации ─────────────────────────────────────────────────

// Result — результат валидации спецификации.
type Result struct {
	Errors   []string
	Warnings []string
}

// OK возвращает true если нет ошибок.
func (r *Result) OK() bool { return len(r.Errors) == 0 }

func (r *Result) addError(format string, args ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Result) addWarning(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// ── Константы и паттерны ────────────────────────────────────────────────

// slugRe — формат slug из JSON Schema: ^[a-z0-9][a-z0-9._-]*$
var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
}

var validAccess = map[string]bool{
	"open": true, "gated": true, "private": true,
}

var validDataTypes = map[string]bool{
	"basic": true, "medium": true, "complex": true,
}

var validAuthTypes = map[string]bool{
	"bearer": true, "header": true, "none": true,
}

var validConfigFieldTypes = map[string]bool{
	"string": true, "number": true, "boolean": true, "select": true,
}

// ── API ─────────────────────────────────────────────────────────────────

// Parse парсит YAML-данные в Spec.
func Parse(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("невалидный YAML: %w", err)
	}
	return &spec, nil
}

// Validate выполняет полную валидацию спецификации.
func Validate(spec *Spec) *Result {
	r := &Result{}
	validatePlugin(spec, r)
	validateProvider(spec, r)
	validateEndpoints(spec, r)
	validateConfigFields(spec, r)
	return r
}

// ValidateBytes парсит YAML и выполняет валидацию.
func ValidateBytes(data []byte) (*Spec, *Result) {
	spec, err := Parse(data)
	if err != nil {
		return nil, &Result{Errors: []string{err.Error()}}
	}
	return spec, Validate(spec)
}

// ── Валидация секций ────────────────────────────────────────────────────

func validatePlugin(spec *Spec, r *Result) {
	p := spec.Plugin

	if p.Slug == "" {
		r.addError("plugin.slug: обязательное поле")
	} else if !slugRe.MatchString(p.Slug) {
		r.addError("plugin.slug: невалидный формат %q (ожидается ^[a-z0-9][a-z0-9._-]*$)", p.Slug)
	}

	if p.Name == "" {
		r.addError("plugin.name: обязательное поле")
	}
	if p.Description == "" {
		r.addError("plugin.description: обязательное поле")
	}
	if p.Version == "" {
		r.addError("plugin.version: обязательное поле")
	}
}

func validateProvider(spec *Spec, r *Result) {
	prov := spec.Provider

	if prov.BaseURL == "" {
		r.addError("provider.base_url: обязательное поле")
	}

	if prov.Auth != nil {
		if prov.Auth.Type != "" && !validAuthTypes[prov.Auth.Type] {
			r.addError("provider.auth.type: недопустимое значение %q (допустимо: bearer, header, none)", prov.Auth.Type)
		}
		if prov.Auth.Type == "header" && prov.Auth.Header == "" {
			r.addWarning("provider.auth: type=header, но header не указан")
		}
		if (prov.Auth.Type == "bearer" || prov.Auth.Type == "header") && prov.Auth.Env == "" {
			r.addWarning("provider.auth: type=%s, но env не указан", prov.Auth.Type)
		}
	}
}

func validateEndpoints(spec *Spec, r *Result) {
	if len(spec.Endpoints) == 0 {
		r.addError("endpoints: минимум 1 эндпоинт обязателен")
		return
	}

	slugs := make(map[string]int)
	for i, ep := range spec.Endpoints {
		prefix := fmt.Sprintf("endpoints[%d]", i)

		if ep.Slug == "" {
			r.addError("%s.slug: обязательное поле", prefix)
		} else {
			if !slugRe.MatchString(ep.Slug) {
				r.addError("%s.slug: невалидный формат %q", prefix, ep.Slug)
			}
			if prev, ok := slugs[ep.Slug]; ok {
				r.addError("%s.slug: дубликат %q (первое появление: endpoints[%d])", prefix, ep.Slug, prev)
			}
			slugs[ep.Slug] = i
		}

		if ep.Name == "" {
			r.addError("%s.name: обязательное поле", prefix)
		}
		if ep.Path == "" {
			r.addError("%s.path: обязательное поле", prefix)
		} else if !strings.HasPrefix(ep.Path, "/") {
			r.addError("%s.path: должен начинаться с / (получено %q)", prefix, ep.Path)
		}

		if ep.Access == "" {
			r.addError("%s.access: обязательное поле", prefix)
		} else if !validAccess[ep.Access] {
			r.addError("%s.access: недопустимое значение %q (допустимо: open, gated, private)", prefix, ep.Access)
		}

		if ep.Method != "" && !validMethods[ep.Method] {
			r.addError("%s.method: недопустимое значение %q (допустимо: GET, POST, PUT, DELETE)", prefix, ep.Method)
		}

		if ep.DataType != "" && !validDataTypes[ep.DataType] {
			r.addError("%s.data_type: недопустимое значение %q (допустимо: basic, medium, complex)", prefix, ep.DataType)
		}

		if ep.CacheTTL != nil && *ep.CacheTTL < 0 {
			r.addError("%s.cache_ttl: должен быть >= 0 (получено %d)", prefix, *ep.CacheTTL)
		}

		validateParamsSchema(ep, prefix, r)
	}
}

func validateParamsSchema(ep EndpointDef, prefix string, r *Result) {
	if ep.ParamsSchema.Kind == 0 {
		return // нет params_schema — OK
	}

	// Конвертируем yaml.Node → map для проверки
	var schema map[string]any
	raw, err := yaml.Marshal(&ep.ParamsSchema)
	if err != nil {
		r.addError("%s.params_schema: ошибка сериализации: %v", prefix, err)
		return
	}
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		r.addError("%s.params_schema: невалидная структура: %v", prefix, err)
		return
	}

	// params_schema.type должен быть "object"
	schemaType, ok := schema["type"]
	if !ok {
		r.addError("%s.params_schema: отсутствует поле type", prefix)
	} else if schemaType != "object" {
		r.addError("%s.params_schema.type: ожидается \"object\", получено %q", prefix, schemaType)
	}

	// Проверяем что сериализуется в JSON
	if _, err := json.Marshal(schema); err != nil {
		r.addError("%s.params_schema: не сериализуется в JSON: %v", prefix, err)
	}
}

func validateConfigFields(spec *Spec, r *Result) {
	if len(spec.ConfigFields) == 0 {
		return // config_fields необязательна
	}

	slugs := make(map[string]int)
	for i, cf := range spec.ConfigFields {
		prefix := fmt.Sprintf("config_fields[%d]", i)

		if cf.Slug == "" {
			r.addError("%s.slug: обязательное поле", prefix)
		} else {
			if prev, ok := slugs[cf.Slug]; ok {
				r.addError("%s.slug: дубликат %q (первое появление: config_fields[%d])", prefix, cf.Slug, prev)
			}
			slugs[cf.Slug] = i
		}

		if cf.Label == "" {
			r.addError("%s.label: обязательное поле", prefix)
		}

		if cf.Type == "" {
			r.addError("%s.type: обязательное поле", prefix)
		} else if !validConfigFieldTypes[cf.Type] {
			r.addError("%s.type: недопустимое значение %q (допустимо: string, number, boolean, select)", prefix, cf.Type)
		}

		if cf.Type == "select" && len(cf.Options) == 0 {
			r.addWarning("%s: type=select, но options не указаны", prefix)
		}

		for j, opt := range cf.Options {
			optPrefix := fmt.Sprintf("%s.options[%d]", prefix, j)
			if opt.Value == "" {
				r.addError("%s.value: обязательное поле", optPrefix)
			}
			if opt.Label == "" {
				r.addError("%s.label: обязательное поле", optPrefix)
			}
		}
	}
}
