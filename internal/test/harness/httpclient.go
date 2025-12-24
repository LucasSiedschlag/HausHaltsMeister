package harness

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type HTTPClient struct {
	Handler http.Handler
}

func NewHTTPClient(e *echo.Echo) *HTTPClient {
	return &HTTPClient{Handler: e}
}

func (c *HTTPClient) Request(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c.Handler.ServeHTTP(rec, req)
	return rec
}
