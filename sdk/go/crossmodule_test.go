package integrat_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/plagness/Integrat/sdk/go/internal/validator"
)

// testdataRoot возвращает путь к корню NeuronSwarm (3 уровня вверх от sdk/go/).
func testdataRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

// readModuleYAML читает integrat.yaml из модуля NeuronSwarm.
func readModuleYAML(t *testing.T, module string) []byte {
	t.Helper()
	path := filepath.Join(testdataRoot(t), module, "integrat.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return data
}

// ── Валидация integrat.yaml всех модулей ─────────────────────────────────

func TestCrossModule_ChannelMCP(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "channel-mcp"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("channel-mcp: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "channel-mcp" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "channel-mcp")
	}
	if len(spec.Endpoints) != 4 {
		t.Errorf("endpoints = %d, want 4", len(spec.Endpoints))
	}
	if len(spec.ConfigFields) != 2 {
		t.Errorf("config_fields = %d, want 2", len(spec.ConfigFields))
	}
}

func TestCrossModule_LLMMCP(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "llm-mcp"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("llm-mcp: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "llm-mcp" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "llm-mcp")
	}
	if len(spec.Endpoints) < 3 {
		t.Errorf("endpoints = %d, want >= 3", len(spec.Endpoints))
	}
}

func TestCrossModule_TelegramMCP(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "telegram-mcp"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("telegram-mcp: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "telegram-chat-data" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "telegram-chat-data")
	}
	if len(spec.Endpoints) < 5 {
		t.Errorf("endpoints = %d, want >= 5", len(spec.Endpoints))
	}
}

func TestCrossModule_BCSMCP(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "bcs-mcp"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("bcs-mcp: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "bcs-mcp" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "bcs-mcp")
	}
	if len(spec.Endpoints) < 10 {
		t.Errorf("endpoints = %d, want >= 10", len(spec.Endpoints))
	}
}

func TestCrossModule_ArenaLLM(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "arena-llm"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("arena-llm: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "arena-llm" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "arena-llm")
	}
	if len(spec.Endpoints) < 5 {
		t.Errorf("endpoints = %d, want >= 5", len(spec.Endpoints))
	}
}

func TestCrossModule_Democracy(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "Democracy"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("democracy: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "democracy" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "democracy")
	}
	if len(spec.Endpoints) < 3 {
		t.Errorf("endpoints = %d, want >= 3", len(spec.Endpoints))
	}
}

func TestCrossModule_Metrics(t *testing.T) {
	spec, r := validator.ValidateBytes(readModuleYAML(t, "metrics"))
	if spec == nil {
		t.Fatal("spec is nil")
	}
	if !r.OK() {
		t.Errorf("metrics: validation errors: %v", r.Errors)
	}
	if spec.Plugin.Slug != "metrics-api" {
		t.Errorf("slug = %q, want %q", spec.Plugin.Slug, "metrics-api")
	}
	if len(spec.Endpoints) < 3 {
		t.Errorf("endpoints = %d, want >= 3", len(spec.Endpoints))
	}
}

// ── Общие проверки ────────────────────────────────────────────────────────

// TestCrossModule_AllSlugsUnique проверяет уникальность slug среди всех модулей.
func TestCrossModule_AllSlugsUnique(t *testing.T) {
	modules := []string{
		"channel-mcp", "llm-mcp", "telegram-mcp", "bcs-mcp",
		"arena-llm", "Democracy", "metrics",
	}

	seen := make(map[string]string) // slug → module
	for _, mod := range modules {
		spec, r := validator.ValidateBytes(readModuleYAML(t, mod))
		if spec == nil {
			t.Errorf("%s: parse failed", mod)
			continue
		}
		if !r.OK() {
			t.Errorf("%s: validation errors: %v", mod, r.Errors)
		}

		slug := spec.Plugin.Slug
		if prev, ok := seen[slug]; ok {
			t.Errorf("duplicate plugin slug %q: %s and %s", slug, prev, mod)
		}
		seen[slug] = mod
	}

	if len(seen) != len(modules) {
		t.Errorf("expected %d unique slugs, got %d", len(modules), len(seen))
	}
}

// TestCrossModule_AllEndpointsHaveAccess проверяет что все эндпоинты имеют access level.
func TestCrossModule_AllEndpointsHaveAccess(t *testing.T) {
	modules := []string{
		"channel-mcp", "llm-mcp", "telegram-mcp", "bcs-mcp",
		"arena-llm", "Democracy", "metrics",
	}

	validAccess := map[string]bool{"open": true, "gated": true, "private": true}
	total := 0

	for _, mod := range modules {
		spec, _ := validator.ValidateBytes(readModuleYAML(t, mod))
		if spec == nil {
			continue
		}
		for _, ep := range spec.Endpoints {
			total++
			if !validAccess[ep.Access] {
				t.Errorf("%s/%s: invalid access %q", mod, ep.Slug, ep.Access)
			}
		}
	}

	if total < 30 {
		t.Errorf("expected >= 30 total endpoints across all modules, got %d", total)
	}
}
