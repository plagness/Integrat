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

export interface Endpoint {
  id: number;
  plugin_id: number;
  name: string;
  slug: string;
  description?: string;
  params_schema?: any;
  response_schema?: any;
  access_tier: string;
  data_type: string;
  cache_ttl: number;
  proxy_path?: string;
  proxy_method?: string;
  created_at: string;
}

export interface CreatePluginParams {
  name: string;
  slug: string;
  base_url: string;
  description?: string;
  github_url?: string;
  version?: string;
  config_fields?: any;
}

export interface UpdatePluginParams {
  name?: string;
  description?: string;
  github_url?: string;
  base_url?: string;
  version?: string;
}

export interface CreateEndpointParams {
  name: string;
  slug: string;
  description?: string;
  access_tier?: string;
  data_type?: string;
  cache_ttl?: number;
  proxy_path?: string;
  proxy_method?: string;
  params_schema?: any;
}

export interface UpdateEndpointParams {
  name?: string;
  description?: string;
  access_tier?: string;
  data_type?: string;
  cache_ttl?: number;
  proxy_path?: string;
  proxy_method?: string;
}

export interface MarketplaceSearchParams {
  query?: string;
  category?: string;
  sort?: string;
  page?: number;
  limit?: number;
}

export interface MarketplaceResult {
  plugins: Plugin[];
  total: number;
  page: number;
  pages: number;
}

export interface PluginDetail {
  plugin: Plugin;
  endpoints: Endpoint[];
}

export interface IntegratOptions {
  /** Базовый URL API (по умолчанию https://integrat.plag.space) */
  baseURL?: string;
  /** Таймаут запросов в мс (по умолчанию 30000) */
  timeout?: number;
}

export class IntegratError extends Error {
  status: number;
  code: string;
  body: any;
  constructor(status: number, body: any);

  /** Проверка: ошибка авторизации (401). */
  get isUnauthorized(): boolean;
  /** Проверка: ресурс не найден (404). */
  get isNotFound(): boolean;
  /** Проверка: доступ запрещён (403). */
  get isForbidden(): boolean;
  /** Проверка: провайдер недоступен (502-504). */
  get isProviderError(): boolean;
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

  /** Список плагинов текущего пользователя */
  listPlugins(): Promise<Plugin[]>;

  /** Создать плагин */
  createPlugin(params: CreatePluginParams): Promise<Plugin>;

  /** Получить плагин по ID */
  getPlugin(id: number): Promise<Plugin>;

  /** Обновить плагин */
  updatePlugin(id: number, params: UpdatePluginParams): Promise<Plugin>;

  /** Удалить плагин */
  deletePlugin(id: number): Promise<void>;

  /** Список эндпоинтов плагина */
  listEndpoints(pluginId: number): Promise<Endpoint[]>;

  /** Создать эндпоинт */
  createEndpoint(pluginId: number, params: CreateEndpointParams): Promise<Endpoint>;

  /** Обновить эндпоинт */
  updateEndpoint(pluginId: number, endpointId: number, params: UpdateEndpointParams): Promise<Endpoint>;

  /** Удалить эндпоинт */
  deleteEndpoint(pluginId: number, endpointId: number): Promise<void>;

  /** Поиск плагинов в маркетплейсе */
  searchMarketplace(params?: MarketplaceSearchParams): Promise<MarketplaceResult>;

  /** Детали плагина по slug из маркетплейса */
  getPluginBySlug(slug: string): Promise<PluginDetail>;

  /** Health check */
  health(): Promise<boolean>;
}
