// Пакет integrat — Go SDK для платформы Integrat.
// Позволяет запрашивать и публиковать данные плагинов через единый API.
package integrat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const DefaultBaseURL = "https://integrat.plag.space"

// Client — клиент для работы с Integrat API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New создаёт клиент с API-токеном.
func New(token string) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewWithURL создаёт клиент с кастомным URL (для self-hosted или dev).
func NewWithURL(token, baseURL string) *Client {
	c := New(token)
	c.BaseURL = baseURL
	return c
}

// ── Типы ────────────────────────────────────────────────────────────────

// QueryRequest — параметры запроса данных.
type QueryRequest struct {
	Plugin   string         `json:"plugin"`
	Endpoint string         `json:"endpoint"`
	ChatID   int64          `json:"chat_id,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
}

// QueryResponse — ответ с данными.
type QueryResponse struct {
	Data   json.RawMessage `json:"data"`
	Cached bool            `json:"cached"`
	Stale  bool            `json:"stale"`
	TTL    int             `json:"ttl"`
}

// UnmarshalData десериализует данные ответа в указанную структуру.
func (r *QueryResponse) UnmarshalData(v any) error {
	if r.Data == nil {
		return fmt.Errorf("integrat: response data is nil")
	}
	return json.Unmarshal(r.Data, v)
}

// ErrorResponse — ответ с ошибкой от API (legacy, используйте errors.Is с APIError).
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorResponse) String() string {
	if e.Message != "" {
		return fmt.Sprintf("integrat: %d %s", e.Code, e.Message)
	}
	return fmt.Sprintf("integrat: %d %s", e.Code, e.Error)
}

// Plugin — информация о плагине.
type Plugin struct {
	ID           int64           `json:"id"`
	Slug         string          `json:"slug"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Version      string          `json:"version"`
	BaseURL      string          `json:"base_url"`
	OwnerID      int64           `json:"owner_id"`
	Status       string          `json:"status"`
	ConfigFields json.RawMessage `json:"config_fields"`
}

// Endpoint — информация об эндпоинте плагина.
type Endpoint struct {
	ID             int64           `json:"id"`
	PluginID       int64           `json:"plugin_id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	Description    string          `json:"description,omitempty"`
	ParamsSchema   json.RawMessage `json:"params_schema,omitempty"`
	ResponseSchema json.RawMessage `json:"response_schema,omitempty"`
	AccessTier     string          `json:"access_tier"`
	DataType       string          `json:"data_type"`
	CacheTTL       int             `json:"cache_ttl"`
	ProxyPath      string          `json:"proxy_path,omitempty"`
	ProxyMethod    string          `json:"proxy_method,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

// CreatePluginParams — параметры создания плагина.
type CreatePluginParams struct {
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	BaseURL      string          `json:"base_url"`
	Description  string          `json:"description,omitempty"`
	GithubURL    string          `json:"github_url,omitempty"`
	Version      string          `json:"version,omitempty"`
	ConfigFields json.RawMessage `json:"config_fields,omitempty"`
}

// UpdatePluginParams — параметры обновления плагина (nil = не менять).
type UpdatePluginParams struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	GithubURL   *string `json:"github_url,omitempty"`
	BaseURL     *string `json:"base_url,omitempty"`
	Version     *string `json:"version,omitempty"`
}

// CreateEndpointParams — параметры создания эндпоинта.
type CreateEndpointParams struct {
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	Description  string          `json:"description,omitempty"`
	AccessTier   string          `json:"access_tier,omitempty"`
	DataType     string          `json:"data_type,omitempty"`
	CacheTTL     int             `json:"cache_ttl,omitempty"`
	ProxyPath    string          `json:"proxy_path,omitempty"`
	ProxyMethod  string          `json:"proxy_method,omitempty"`
	ParamsSchema json.RawMessage `json:"params_schema,omitempty"`
}

// UpdateEndpointParams — параметры обновления эндпоинта (nil = не менять).
type UpdateEndpointParams struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AccessTier  *string `json:"access_tier,omitempty"`
	DataType    *string `json:"data_type,omitempty"`
	CacheTTL    *int    `json:"cache_ttl,omitempty"`
	ProxyPath   *string `json:"proxy_path,omitempty"`
	ProxyMethod *string `json:"proxy_method,omitempty"`
}

// MarketplaceSearchParams — параметры поиска в маркетплейсе.
type MarketplaceSearchParams struct {
	Query    string `json:"q,omitempty"`
	Category string `json:"category,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// MarketplaceResult — результат поиска в маркетплейсе.
type MarketplaceResult struct {
	Plugins []Plugin `json:"plugins"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Pages   int      `json:"pages"`
}

// PluginDetail — детальная информация о плагине из маркетплейса.
type PluginDetail struct {
	Plugin    Plugin     `json:"plugin"`
	Endpoints []Endpoint `json:"endpoints"`
}

// ── Внутренний HTTP ─────────────────────────────────────────────────────

// doRequest выполняет HTTP-запрос и возвращает тело ответа.
// При статусе >= 400 возвращает *APIError.
func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return nil, 0, fmt.Errorf("integrat: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("integrat: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("integrat: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, newAPIError(resp.StatusCode, respBody)
	}

	return respBody, resp.StatusCode, nil
}

// doJSON маршалит body в JSON и вызывает doRequest.
func (c *Client) doJSON(method, path string, body any) ([]byte, int, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("integrat: marshal: %w", err)
	}
	return c.doRequest(method, path, bytes.NewReader(data))
}

// ── Query ───────────────────────────────────────────────────────────────

// Query выполняет запрос данных через прокси (dev-режим, без привязки к чату).
func (c *Client) Query(plugin, endpoint string, params map[string]any) (*QueryResponse, error) {
	return c.QueryInChat(plugin, endpoint, 0, params)
}

// QueryInChat выполняет запрос данных в контексте конкретного чата.
func (c *Client) QueryInChat(plugin, endpoint string, chatID int64, params map[string]any) (*QueryResponse, error) {
	qr := QueryRequest{
		Plugin:   plugin,
		Endpoint: endpoint,
		ChatID:   chatID,
		Params:   params,
	}

	body, err := json.Marshal(qr)
	if err != nil {
		return nil, fmt.Errorf("integrat: marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/v1/query", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("integrat: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("integrat: http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("integrat: read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		return nil, newAPIError(httpResp.StatusCode, respBody)
	}

	var result QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal response: %w", err)
	}

	// Кеш-заголовки
	result.Cached = httpResp.Header.Get("X-Integrat-Cached") == "true"
	result.Stale = httpResp.Header.Get("X-Integrat-Stale") == "true"

	return &result, nil
}

// ── Plugin CRUD ─────────────────────────────────────────────────────────

// ListPlugins возвращает плагины текущего пользователя.
func (c *Client) ListPlugins() ([]Plugin, error) {
	respBody, _, err := c.doRequest("GET", "/v1/plugins", nil)
	if err != nil {
		return nil, err
	}
	var plugins []Plugin
	if err := json.Unmarshal(respBody, &plugins); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return plugins, nil
}

// CreatePlugin создаёт новый плагин.
func (c *Client) CreatePlugin(params CreatePluginParams) (*Plugin, error) {
	respBody, _, err := c.doJSON("POST", "/v1/plugins", params)
	if err != nil {
		return nil, err
	}
	var plugin Plugin
	if err := json.Unmarshal(respBody, &plugin); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &plugin, nil
}

// GetPlugin возвращает плагин по ID.
func (c *Client) GetPlugin(id int64) (*Plugin, error) {
	respBody, _, err := c.doRequest("GET", fmt.Sprintf("/v1/plugins/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var plugin Plugin
	if err := json.Unmarshal(respBody, &plugin); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &plugin, nil
}

// UpdatePlugin обновляет плагин.
func (c *Client) UpdatePlugin(id int64, params UpdatePluginParams) (*Plugin, error) {
	respBody, _, err := c.doJSON("PUT", fmt.Sprintf("/v1/plugins/%d", id), params)
	if err != nil {
		return nil, err
	}
	var plugin Plugin
	if err := json.Unmarshal(respBody, &plugin); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &plugin, nil
}

// DeletePlugin удаляет плагин.
func (c *Client) DeletePlugin(id int64) error {
	_, _, err := c.doRequest("DELETE", fmt.Sprintf("/v1/plugins/%d", id), nil)
	return err
}

// ── Endpoint CRUD ───────────────────────────────────────────────────────

// ListEndpoints возвращает эндпоинты плагина.
func (c *Client) ListEndpoints(pluginID int64) ([]Endpoint, error) {
	respBody, _, err := c.doRequest("GET", fmt.Sprintf("/v1/plugins/%d/endpoints", pluginID), nil)
	if err != nil {
		return nil, err
	}
	var endpoints []Endpoint
	if err := json.Unmarshal(respBody, &endpoints); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return endpoints, nil
}

// CreateEndpoint создаёт эндпоинт для плагина.
func (c *Client) CreateEndpoint(pluginID int64, params CreateEndpointParams) (*Endpoint, error) {
	respBody, _, err := c.doJSON("POST", fmt.Sprintf("/v1/plugins/%d/endpoints", pluginID), params)
	if err != nil {
		return nil, err
	}
	var ep Endpoint
	if err := json.Unmarshal(respBody, &ep); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &ep, nil
}

// UpdateEndpoint обновляет эндпоинт.
func (c *Client) UpdateEndpoint(pluginID, endpointID int64, params UpdateEndpointParams) (*Endpoint, error) {
	respBody, _, err := c.doJSON("PUT", fmt.Sprintf("/v1/plugins/%d/endpoints/%d", pluginID, endpointID), params)
	if err != nil {
		return nil, err
	}
	var ep Endpoint
	if err := json.Unmarshal(respBody, &ep); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &ep, nil
}

// DeleteEndpoint удаляет эндпоинт.
func (c *Client) DeleteEndpoint(pluginID, endpointID int64) error {
	_, _, err := c.doRequest("DELETE", fmt.Sprintf("/v1/plugins/%d/endpoints/%d", pluginID, endpointID), nil)
	return err
}

// ── Marketplace ─────────────────────────────────────────────────────────

// SearchMarketplace ищет плагины в маркетплейсе.
func (c *Client) SearchMarketplace(params MarketplaceSearchParams) (*MarketplaceResult, error) {
	v := url.Values{}
	if params.Query != "" {
		v.Set("q", params.Query)
	}
	if params.Category != "" {
		v.Set("category", params.Category)
	}
	if params.Sort != "" {
		v.Set("sort", params.Sort)
	}
	if params.Page > 0 {
		v.Set("page", strconv.Itoa(params.Page))
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}

	path := "/v1/marketplace"
	if qs := v.Encode(); qs != "" {
		path += "?" + qs
	}

	respBody, _, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var result MarketplaceResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &result, nil
}

// GetPluginBySlug возвращает полную информацию о плагине из маркетплейса.
func (c *Client) GetPluginBySlug(slug string) (*PluginDetail, error) {
	respBody, _, err := c.doRequest("GET", "/v1/marketplace/"+slug, nil)
	if err != nil {
		return nil, err
	}
	var detail PluginDetail
	if err := json.Unmarshal(respBody, &detail); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal: %w", err)
	}
	return &detail, nil
}

// ── Health ───────────────────────────────────────────────────────────────

// Health проверяет доступность API.
func (c *Client) Health() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("integrat: health check: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("integrat: health check returned %d", resp.StatusCode)
	}
	return nil
}
