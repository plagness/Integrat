/**
 * Integrat SDK — клиент для платформы обмена данными.
 *
 * @example
 * const { Integrat } = require('@integrat/sdk');
 * const client = new Integrat('itg_your_token');
 * const data = await client.query('channel-mcp', 'messages.fetch', { limit: 10 });
 */

const DEFAULT_BASE_URL = 'https://integrat.plag.space';

class IntegratError extends Error {
  constructor(status, body) {
    const msg = body?.error || body?.message || `HTTP ${status}`;
    super(msg);
    this.name = 'IntegratError';
    this.status = status;
    this.code = body?.code || IntegratError._codeFromStatus(status);
    this.body = body;
  }

  /** @private */
  static _codeFromStatus(status) {
    switch (status) {
      case 401: return 'unauthorized';
      case 403: return 'forbidden';
      case 404: return 'not_found';
      case 409: return 'conflict';
      case 502: case 503: case 504: return 'provider_unavailable';
      default: return 'unknown';
    }
  }

  /** Проверка: ошибка авторизации. */
  get isUnauthorized() { return this.status === 401; }
  /** Проверка: ресурс не найден. */
  get isNotFound() { return this.status === 404; }
  /** Проверка: доступ запрещён. */
  get isForbidden() { return this.status === 403; }
  /** Проверка: провайдер недоступен. */
  get isProviderError() { return this.status >= 502 && this.status <= 504; }
}

class Integrat {
  /**
   * @param {string} token — API-токен (itg_...)
   * @param {object} [opts]
   * @param {string} [opts.baseURL] — базовый URL API
   * @param {number} [opts.timeout] — таймаут запросов в мс (по умолчанию 30000)
   */
  constructor(token, opts = {}) {
    if (!token) throw new Error('Integrat: token is required');
    this.token = token;
    this.baseURL = (opts.baseURL || DEFAULT_BASE_URL).replace(/\/+$/, '');
    this.timeout = opts.timeout || 30000;
  }

  // ── Query ───────────────────────────────────────────────────────────

  /**
   * Запрос данных через прокси.
   *
   * @param {string} plugin — slug плагина
   * @param {string} endpoint — slug эндпоинта
   * @param {object} [params] — параметры запроса
   * @param {number} [chatId] — ID чата (0 = dev-режим)
   * @returns {Promise<{data: any, cached: boolean, stale: boolean, ttl: number}>}
   */
  async query(plugin, endpoint, params = {}, chatId = 0) {
    const body = { plugin, endpoint, params };
    if (chatId) body.chat_id = chatId;

    const resp = await this._fetch('/v1/query', {
      method: 'POST',
      body: JSON.stringify(body),
    });

    const json = await resp.json();
    if (!resp.ok) throw new IntegratError(resp.status, json);

    return {
      data: json.data,
      cached: resp.headers.get('x-integrat-cached') === 'true',
      stale: resp.headers.get('x-integrat-stale') === 'true',
      ttl: parseInt(resp.headers.get('x-integrat-ttl') || '0', 10),
    };
  }

  // ── Plugin CRUD ─────────────────────────────────────────────────────

  /**
   * Список плагинов текущего пользователя.
   * @returns {Promise<Array>}
   */
  async listPlugins() {
    return this._json('/v1/plugins');
  }

  /**
   * Создать плагин.
   * @param {object} params — { name, slug, base_url, description?, github_url?, version? }
   * @returns {Promise<object>}
   */
  async createPlugin(params) {
    return this._json('/v1/plugins', { method: 'POST', body: JSON.stringify(params) });
  }

  /**
   * Получить плагин по ID.
   * @param {number} id
   * @returns {Promise<object>}
   */
  async getPlugin(id) {
    return this._json(`/v1/plugins/${id}`);
  }

  /**
   * Обновить плагин.
   * @param {number} id
   * @param {object} params — { name?, description?, github_url?, base_url?, version? }
   * @returns {Promise<object>}
   */
  async updatePlugin(id, params) {
    return this._json(`/v1/plugins/${id}`, { method: 'PUT', body: JSON.stringify(params) });
  }

  /**
   * Удалить плагин.
   * @param {number} id
   */
  async deletePlugin(id) {
    const resp = await this._fetch(`/v1/plugins/${id}`, { method: 'DELETE' });
    if (!resp.ok && resp.status !== 204) {
      const body = await resp.json().catch(() => ({}));
      throw new IntegratError(resp.status, body);
    }
  }

  // ── Endpoint CRUD ───────────────────────────────────────────────────

  /**
   * Список эндпоинтов плагина.
   * @param {number} pluginId
   * @returns {Promise<Array>}
   */
  async listEndpoints(pluginId) {
    return this._json(`/v1/plugins/${pluginId}/endpoints`);
  }

  /**
   * Создать эндпоинт.
   * @param {number} pluginId
   * @param {object} params — { name, slug, description?, access_tier?, data_type?, cache_ttl? }
   * @returns {Promise<object>}
   */
  async createEndpoint(pluginId, params) {
    return this._json(`/v1/plugins/${pluginId}/endpoints`, {
      method: 'POST', body: JSON.stringify(params),
    });
  }

  /**
   * Обновить эндпоинт.
   * @param {number} pluginId
   * @param {number} endpointId
   * @param {object} params
   * @returns {Promise<object>}
   */
  async updateEndpoint(pluginId, endpointId, params) {
    return this._json(`/v1/plugins/${pluginId}/endpoints/${endpointId}`, {
      method: 'PUT', body: JSON.stringify(params),
    });
  }

  /**
   * Удалить эндпоинт.
   * @param {number} pluginId
   * @param {number} endpointId
   */
  async deleteEndpoint(pluginId, endpointId) {
    const resp = await this._fetch(`/v1/plugins/${pluginId}/endpoints/${endpointId}`, { method: 'DELETE' });
    if (!resp.ok && resp.status !== 204) {
      const body = await resp.json().catch(() => ({}));
      throw new IntegratError(resp.status, body);
    }
  }

  // ── Marketplace ─────────────────────────────────────────────────────

  /**
   * Поиск плагинов в маркетплейсе.
   * @param {object} [params] — { query?, category?, sort?, page?, limit? }
   * @returns {Promise<{plugins: Array, total: number, page: number, pages: number}>}
   */
  async searchMarketplace(params = {}) {
    const qs = new URLSearchParams();
    if (params.query) qs.set('q', params.query);
    if (params.category) qs.set('category', params.category);
    if (params.sort) qs.set('sort', params.sort);
    if (params.page) qs.set('page', String(params.page));
    if (params.limit) qs.set('limit', String(params.limit));
    const q = qs.toString();
    return this._json(`/v1/marketplace${q ? '?' + q : ''}`);
  }

  /**
   * Детали плагина по slug из маркетплейса.
   * @param {string} slug
   * @returns {Promise<{plugin: object, endpoints: Array}>}
   */
  async getPluginBySlug(slug) {
    return this._json(`/v1/marketplace/${slug}`);
  }

  // ── Health ──────────────────────────────────────────────────────────

  /**
   * Health check.
   * @returns {Promise<boolean>}
   */
  async health() {
    const resp = await this._fetch('/health');
    return resp.ok;
  }

  // ── Внутренние методы ───────────────────────────────────────────────

  /** @private Выполняет запрос и возвращает parsed JSON. */
  async _json(path, opts = {}) {
    const resp = await this._fetch(path, opts);
    const json = await resp.json();
    if (!resp.ok) throw new IntegratError(resp.status, json);
    return json;
  }

  /** @private */
  async _fetch(path, opts = {}) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeout);

    try {
      return await fetch(this.baseURL + path, {
        ...opts,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${this.token}`,
          ...opts.headers,
        },
      });
    } finally {
      clearTimeout(timer);
    }
  }
}

module.exports = { Integrat, IntegratError };
