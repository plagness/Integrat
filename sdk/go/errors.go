// Типизированные ошибки Integrat SDK.
// Позволяют проверять тип ошибки через errors.Is.
package integrat

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Sentinel-ошибки для проверки через errors.Is.
var (
	ErrUnauthorized = errors.New("integrat: unauthorized")
	ErrForbidden    = errors.New("integrat: access denied")
	ErrNotFound     = errors.New("integrat: not found")
	ErrConflict     = errors.New("integrat: conflict")
	ErrProvider     = errors.New("integrat: provider unavailable")
)

// APIError — структурная ошибка API с HTTP-статусом, кодом и сообщением.
type APIError struct {
	StatusCode int    // HTTP статус
	Code       string // Код ошибки из API (например "not_found", "limit_exceeded")
	Message    string // Текст ошибки
	Err        error  // Базовая sentinel-ошибка для errors.Is
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("integrat: %d [%s] %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("integrat: %d %s", e.StatusCode, e.Message)
}

// Unwrap позволяет errors.Is работать с sentinel-ошибками.
func (e *APIError) Unwrap() error {
	return e.Err
}

// newAPIError создаёт APIError из HTTP-ответа.
func newAPIError(status int, body []byte) *APIError {
	ae := &APIError{StatusCode: status}

	// Пытаемся распарсить JSON-ответ API
	var errResp struct {
		Error   string `json:"error"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &errResp) == nil {
		if errResp.Error != "" {
			ae.Message = errResp.Error
		} else if errResp.Message != "" {
			ae.Message = errResp.Message
		}
		ae.Code = errResp.Code
	}
	if ae.Message == "" {
		ae.Message = string(body)
	}

	// Привязываем sentinel-ошибку по статусу
	switch {
	case status == 401:
		ae.Err = ErrUnauthorized
	case status == 403:
		ae.Err = ErrForbidden
	case status == 404:
		ae.Err = ErrNotFound
	case status == 409:
		ae.Err = ErrConflict
	case status >= 502 && status <= 504:
		ae.Err = ErrProvider
	}

	return ae
}
