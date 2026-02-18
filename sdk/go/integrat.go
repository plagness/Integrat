// Пакет integrat — Go SDK для платформы Integrat.
// Позволяет запрашивать данные плагинов через единый API.
package integrat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// ErrorResponse — ответ с ошибкой от API.
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

// Query выполняет запрос данных через прокси.
func (c *Client) Query(plugin, endpoint string, params map[string]any) (*QueryResponse, error) {
	return c.QueryInChat(plugin, endpoint, 0, params)
}

// QueryInChat выполняет запрос данных в контексте конкретного чата.
func (c *Client) QueryInChat(plugin, endpoint string, chatID int64, params map[string]any) (*QueryResponse, error) {
	req := QueryRequest{
		Plugin:   plugin,
		Endpoint: endpoint,
		ChatID:   chatID,
		Params:   params,
	}

	body, err := json.Marshal(req)
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
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			errResp.Code = httpResp.StatusCode
			return nil, fmt.Errorf("%s", errResp.String())
		}
		return nil, fmt.Errorf("integrat: HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	var result QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("integrat: unmarshal response: %w", err)
	}

	result.Cached = httpResp.Header.Get("X-Integrat-Cached") == "true"
	result.Stale = httpResp.Header.Get("X-Integrat-Stale") == "true"

	return &result, nil
}

// Plugin — информация о плагине.
type Plugin struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	BaseURL     string          `json:"base_url"`
	OwnerID     int64           `json:"owner_id"`
	Status      string          `json:"status"`
	ConfigFields json.RawMessage `json:"config_fields"`
}

// ListPlugins возвращает список плагинов.
func (c *Client) ListPlugins() ([]Plugin, error) {
	httpReq, err := http.NewRequest("GET", c.BaseURL+"/v1/plugins", nil)
	if err != nil {
		return nil, fmt.Errorf("integrat: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("integrat: http request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("integrat: HTTP %d: %s", httpResp.StatusCode, string(body))
	}

	var plugins []Plugin
	if err := json.NewDecoder(httpResp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("integrat: decode response: %w", err)
	}
	return plugins, nil
}

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
