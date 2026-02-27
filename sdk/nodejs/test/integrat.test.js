/**
 * Cross-module тесты — проверка integrat.yaml всех модулей NeuronSwarm.
 * Использует Node.js built-in test runner (node:test, Node 18+).
 *
 * Запуск: node --test sdk/nodejs/test/integrat.test.js
 */

const { describe, it } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

// Корень NeuronSwarm (3 уровня вверх от sdk/nodejs/test/)
const ROOT = path.resolve(__dirname, '..', '..', '..', '..');

// ── Минимальный YAML-парсер для integrat.yaml ────────────────────────────
// (без внешних зависимостей — парсит достаточно для валидации)

function parseSimpleYAML(text) {
  const result = {};
  let currentSection = null;
  let currentItem = null;
  let items = [];

  for (const line of text.split('\n')) {
    // Пропускаем комментарии и пустые строки
    if (line.trim().startsWith('#') || line.trim() === '') continue;

    // Секция верхнего уровня (без отступа)
    const topMatch = line.match(/^(\w[\w_]*)\s*:/);
    if (topMatch) {
      if (currentSection === 'endpoints' || currentSection === 'config_fields') {
        result[currentSection] = items;
      }
      currentSection = topMatch[1];
      items = [];
      currentItem = null;
      if (!result[currentSection]) result[currentSection] = {};
      continue;
    }

    // Элемент списка
    const listMatch = line.match(/^\s+-\s+(\w[\w_]*)\s*:\s*(.+)/);
    if (listMatch) {
      currentItem = { [listMatch[1]]: listMatch[2].trim().replace(/^["']|["']$/g, '') };
      items.push(currentItem);
      continue;
    }

    // Поле вложенного объекта
    const fieldMatch = line.match(/^\s+(\w[\w_]*)\s*:\s*(.+)/);
    if (fieldMatch) {
      const key = fieldMatch[1];
      let value = fieldMatch[2].trim().replace(/^["']|["']$/g, '');
      if (value === 'true') value = true;
      else if (value === 'false') value = false;

      if (currentItem) {
        currentItem[key] = value;
      } else if (currentSection && typeof result[currentSection] === 'object') {
        result[currentSection][key] = value;
      }
    }
  }

  if (currentSection === 'endpoints' || currentSection === 'config_fields') {
    result[currentSection] = items;
  }

  return result;
}

// ── Валидация integrat.yaml ──────────────────────────────────────────────

function validateIntegratYAML(yamlText) {
  const errors = [];
  const spec = parseSimpleYAML(yamlText);

  // plugin секция
  if (!spec.plugin?.slug) errors.push('plugin.slug: обязательное поле');
  if (!spec.plugin?.name) errors.push('plugin.name: обязательное поле');
  if (!spec.plugin?.description) errors.push('plugin.description: обязательное поле');
  if (!spec.plugin?.version) errors.push('plugin.version: обязательное поле');

  // Формат slug
  if (spec.plugin?.slug && !/^[a-z0-9][a-z0-9._-]*$/.test(spec.plugin.slug)) {
    errors.push(`plugin.slug: невалидный формат "${spec.plugin.slug}"`);
  }

  // provider секция
  if (!spec.provider?.base_url) errors.push('provider.base_url: обязательное поле');

  // endpoints
  if (!Array.isArray(spec.endpoints) || spec.endpoints.length === 0) {
    errors.push('endpoints: минимум 1 эндпоинт обязателен');
  } else {
    const slugs = new Set();
    for (const ep of spec.endpoints) {
      if (!ep.slug) errors.push('endpoint: отсутствует slug');
      if (!ep.name) errors.push(`endpoint ${ep.slug}: отсутствует name`);
      if (!ep.path) errors.push(`endpoint ${ep.slug}: отсутствует path`);
      if (!ep.access) errors.push(`endpoint ${ep.slug}: отсутствует access`);
      if (ep.slug && slugs.has(ep.slug)) errors.push(`endpoint: дубликат slug "${ep.slug}"`);
      if (ep.slug) slugs.add(ep.slug);
    }
  }

  return { spec, errors, ok: errors.length === 0 };
}

// ── Загрузка integrat.yaml модуля ────────────────────────────────────────

function loadModuleYAML(module) {
  const filePath = path.join(ROOT, module, 'integrat.yaml');
  return fs.readFileSync(filePath, 'utf-8');
}

// ── Тесты по модулям ─────────────────────────────────────────────────────

const modules = [
  { dir: 'channel-mcp', slug: 'channel-mcp', minEndpoints: 4 },
  { dir: 'llm-mcp', slug: 'llm-mcp', minEndpoints: 3 },
  { dir: 'telegram-mcp', slug: 'telegram-chat-data', minEndpoints: 5 },
  { dir: 'bcs-mcp', slug: 'bcs-mcp', minEndpoints: 10 },
  { dir: 'arena-llm', slug: 'arena-llm', minEndpoints: 5 },
  { dir: 'Democracy', slug: 'democracy', minEndpoints: 3 },
  { dir: 'metrics', slug: 'metrics-api', minEndpoints: 3 },
];

describe('Cross-module integrat.yaml validation', () => {
  for (const mod of modules) {
    describe(mod.dir, () => {
      it('файл существует и читается', () => {
        const filePath = path.join(ROOT, mod.dir, 'integrat.yaml');
        assert.ok(fs.existsSync(filePath), `${filePath} не найден`);
      });

      it('проходит валидацию', () => {
        const yaml = loadModuleYAML(mod.dir);
        const { errors, ok } = validateIntegratYAML(yaml);
        assert.ok(ok, `Ошибки валидации: ${errors.join(', ')}`);
      });

      it(`slug = "${mod.slug}"`, () => {
        const yaml = loadModuleYAML(mod.dir);
        const { spec } = validateIntegratYAML(yaml);
        assert.equal(spec.plugin.slug, mod.slug);
      });

      it(`endpoints >= ${mod.minEndpoints}`, () => {
        const yaml = loadModuleYAML(mod.dir);
        const { spec } = validateIntegratYAML(yaml);
        assert.ok(
          Array.isArray(spec.endpoints) && spec.endpoints.length >= mod.minEndpoints,
          `endpoints = ${spec.endpoints?.length}, ожидается >= ${mod.minEndpoints}`
        );
      });
    });
  }

  it('все slug уникальны', () => {
    const slugs = new Map();
    for (const mod of modules) {
      const yaml = loadModuleYAML(mod.dir);
      const { spec } = validateIntegratYAML(yaml);
      const slug = spec.plugin?.slug;
      assert.ok(!slugs.has(slug), `дубликат slug "${slug}": ${slugs.get(slug)} и ${mod.dir}`);
      slugs.set(slug, mod.dir);
    }
    assert.equal(slugs.size, modules.length);
  });
});
