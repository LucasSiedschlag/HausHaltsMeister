package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type Metrics struct {
	mu       sync.Mutex
	started  time.Time
	byCode   map[int]int64
	byMethod map[string]int64
}

type metricsResponse struct {
	StartedAt string           `json:"started_at"`
	UptimeSec int64            `json:"uptime_sec"`
	ByStatus  map[int]int64    `json:"by_status"`
	ByMethod  map[string]int64 `json:"by_method"`
}

func NewMetrics() *Metrics {
	return &Metrics{
		started:  time.Now().UTC(),
		byCode:   make(map[int]int64),
		byMethod: make(map[string]int64),
	}
}

func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}
			method := c.Request().Method

			m.mu.Lock()
			m.byCode[status]++
			m.byMethod[method]++
			m.mu.Unlock()
			return nil
		}
	}
}

func (m *Metrics) Handler(c echo.Context) error {
	m.mu.Lock()
	byCode := make(map[int]int64, len(m.byCode))
	for k, v := range m.byCode {
		byCode[k] = v
	}
	byMethod := make(map[string]int64, len(m.byMethod))
	for k, v := range m.byMethod {
		byMethod[k] = v
	}
	m.mu.Unlock()

	uptime := time.Since(m.started)
	return c.JSON(http.StatusOK, metricsResponse{
		StartedAt: m.started.Format(time.RFC3339),
		UptimeSec: int64(uptime.Seconds()),
		ByStatus:  byCode,
		ByMethod:  byMethod,
	})
}
