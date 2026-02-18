/**
 * Integrat SDK — TypeScript определения.
 */

export interface QueryResult {
  /** Данные от провайдера */
  data: any;
  /** true если ответ из кеша */
  cached: boolean;
  /** true если данные устарели (провайдер offline) */
  stale: boolean;
  /** Оставшееся время жизни кеша (сек) */
  ttl: number;
}

export interface Plugin {
  id: number;
  slug: string;
  name: string;
  description: string;
  version: string;
  base_url: string;
  owner_id: number;
  status: string;
  config_fields: any;
}

export interface IntegratOptions {
  /** Базовый URL API (по умолчанию https://integrat.plag.space) */
  baseURL?: string;
  /** Таймаут запросов в мс (по умолчанию 30000) */
  timeout?: number;
}

export class IntegratError extends Error {
  status: number;
  body: any;
  constructor(status: number, body: any);
}

export class Integrat {
  token: string;
  baseURL: string;
  timeout: number;

  constructor(token: string, opts?: IntegratOptions);

  /** Запрос данных через прокси */
  query(
    plugin: string,
    endpoint: string,
    params?: Record<string, any>,
    chatId?: number
  ): Promise<QueryResult>;

  /** Список плагинов */
  listPlugins(): Promise<Plugin[]>;

  /** Health check */
  health(): Promise<boolean>;
}
