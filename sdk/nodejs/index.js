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
    this.body = body;
  }
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

  /**
   * Запрос данных через прокси.
   *
   * @param {string} plugin — slug плагина
   * @param {string} endpoint — slug эндпоинта
   * @param {object} [params] — параметры запроса
   * @param {number} [chatId] — ID чата (контекст доступа)
   * @returns {Promise<{data: any, cached: boolean, stale: boolean, ttl: number}>}
   */
  async query(plugin, endpoint, params = {}, chatId = 0) {
    const body = {
      plugin,
      endpoint,
      params,
    };
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

  /**
   * Список плагинов.
   * @returns {Promise<Array>}
   */
  async listPlugins() {
    const resp = await this._fetch('/v1/plugins');
    const json = await resp.json();
    if (!resp.ok) throw new IntegratError(resp.status, json);
    return json;
  }

  /**
   * Health check.
   * @returns {Promise<boolean>}
   */
  async health() {
    const resp = await this._fetch('/health');
    return resp.ok;
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
