package handlers

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/middleware"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/config"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Service        AuthService
	Config         config.Config
	RateLimiter    *middleware.RateLimiter
	RefreshLimiter *middleware.RateLimiter
}

type authRequest = dto.AuthRequest
type authResponse = dto.AuthResponse
type userResponse = dto.UserResponse

type AuthService interface {
	SignUp(ctx context.Context, email, password, displayName, userAgent, ip string) (auth.AuthResult, error)
	Login(ctx context.Context, email, password, userAgent, ip string, remember bool) (auth.AuthResult, error)
	Refresh(ctx context.Context, refreshToken, userAgent, ip string) (auth.AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	ListSessions(ctx context.Context, userID string) ([]auth.AuthSessionDetails, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	Me(ctx context.Context, userID string) (auth.User, error)
	StartOAuth(ctx context.Context, provider, redirectURI string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code, state, userAgent, ip string) (auth.AuthResult, string, error)
	ParseAccessToken(token string) (auth.AccessTokenClaims, error)
}

func (h *AuthHandler) Register(g *echo.Group) {
	g.POST("/signup", h.SignUp)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout)
	g.GET("/me", h.Me, middleware.RequireAuth(h.Service))
	g.GET("/sessions", h.ListSessions, middleware.RequireAuth(h.Service))
	g.DELETE("/sessions/:sessionId", h.RevokeSession, middleware.RequireAuth(h.Service))
	g.POST("/logout-all", h.LogoutAll, middleware.RequireAuth(h.Service))
	g.GET("/oauth/:provider/start", h.OAuthStart)
	g.GET("/oauth/:provider/callback", h.OAuthCallback)
}

func (h *AuthHandler) SignUp(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	result, err := h.Service.SignUp(c.Request().Context(), req.Email, req.Password, req.DisplayName, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		if authErr, ok := err.(*auth.Error); ok && authErr.Code() == "VALIDATION_ERROR" {
			key := strings.ToLower(strings.TrimSpace(req.Email))
			allowed, retryAfter := h.allow(c, "signup:"+key)
			if !allowed {
				return rateLimitError(c, retryAfter)
			}
		}
		return httpx.WriteAuthError(c, err)
	}
	if h.RateLimiter != nil {
		key := strings.ToLower(strings.TrimSpace(req.Email))
		h.RateLimiter.Reset(c.RealIP() + ":signup:" + key)
	}

	h.setRefreshCookie(c, result.Tokens.RefreshToken, result.Session.ExpiresAt)
	return c.JSON(http.StatusCreated, buildAuthResponse(result))
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	result, err := h.Service.Login(c.Request().Context(), req.Email, req.Password, c.Request().UserAgent(), c.RealIP(), req.Remember)
	if err != nil {
		if authErr, ok := err.(*auth.Error); ok && authErr.Code() == "AUTH_INVALID_CREDENTIALS" {
			key := strings.ToLower(strings.TrimSpace(req.Email))
			allowed, retryAfter := h.allow(c, "login:"+key)
			if !allowed {
				return rateLimitError(c, retryAfter)
			}
		}
		return httpx.WriteAuthError(c, err)
	}
	if h.RateLimiter != nil {
		key := strings.ToLower(strings.TrimSpace(req.Email))
		h.RateLimiter.Reset(c.RealIP() + ":login:" + key)
	}

	h.setRefreshCookie(c, result.Tokens.RefreshToken, result.Session.ExpiresAt)
	return c.JSON(http.StatusOK, buildAuthResponse(result))
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	refreshToken := h.getRefreshToken(c)
	if refreshToken == "" {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_REFRESH_REVOKED", "Refresh token revogado", nil)
	}

	result, err := h.Service.Refresh(c.Request().Context(), refreshToken, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		return httpx.WriteAuthError(c, err)
	}

	h.setRefreshCookie(c, result.Tokens.RefreshToken, result.Session.ExpiresAt)
	return c.JSON(http.StatusOK, buildAuthResponse(result))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	refreshToken := h.getRefreshToken(c)
	if err := h.Service.Logout(c.Request().Context(), refreshToken); err != nil {
		return httpx.WriteAuthError(c, err)
	}
	h.clearRefreshCookie(c)
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) Me(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	return c.JSON(http.StatusOK, userResponseFrom(user))
}

func (h *AuthHandler) ListSessions(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	currentSessionID := ""
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		claims, err := h.Service.ParseAccessToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err == nil {
			currentSessionID = claims.Sid
		}
	}

	sessions, err := h.Service.ListSessions(c.Request().Context(), user.ID)
	if err != nil {
		return httpx.WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Erro interno", nil)
	}

	resp := make([]dto.AuthSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		resp = append(resp, dto.AuthSessionResponse{
			ID:         session.ID,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			UserAgent:  session.UserAgent,
			IP:         session.IP,
			DeviceName: session.DeviceName,
			IsCurrent:  session.ID == currentSessionID,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RevokeSession(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	sessionID, err := httpx.RequireUUIDParam(c, "sessionId")
	if err != nil {
		return err
	}

	if err := h.Service.RevokeSession(c.Request().Context(), user.ID, sessionID); err != nil {
		return httpx.WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Erro interno", nil)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	if err := h.Service.LogoutAll(c.Request().Context(), user.ID); err != nil {
		return httpx.WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Erro interno", nil)
	}
	h.clearRefreshCookie(c)
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) OAuthStart(c echo.Context) error {
	provider := c.Param("provider")
	redirectURI := c.QueryParam("redirect_uri")

	if redirectURI != "" && !isAllowedRedirect(redirectURI, h.Config.OAuthRedirectAllowlist) {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Redirect URI invalida", nil)
	}

	url, err := h.Service.StartOAuth(c.Request().Context(), provider, redirectURI)
	if err != nil {
		return httpx.WriteAuthError(c, err)
	}
	return c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", nil)
	}

	result, redirectURI, err := h.Service.HandleOAuthCallback(c.Request().Context(), provider, code, state, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		return httpx.WriteAuthError(c, err)
	}

	h.setRefreshCookie(c, result.Tokens.RefreshToken, result.Session.ExpiresAt)
	if redirectURI != "" {
		return c.Redirect(http.StatusFound, redirectURI)
	}
	return c.JSON(http.StatusOK, buildAuthResponse(result))
}

func (h *AuthHandler) allow(c echo.Context, action string) (bool, time.Duration) {
	limiter := h.RateLimiter
	if action == "refresh" && h.RefreshLimiter != nil {
		limiter = h.RefreshLimiter
	}
	if limiter == nil {
		return true, 0
	}
	key := c.RealIP() + ":" + action
	allowed, retryAfter := limiter.Allow(key)
	return allowed, retryAfter
}

func rateLimitError(c echo.Context, retryAfter time.Duration) error {
	seconds := int(retryAfter.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return httpx.WriteError(c, http.StatusTooManyRequests, "AUTH_RATE_LIMITED", "Muitas tentativas", map[string]string{"retry_after": strconv.Itoa(seconds)})
}

func buildAuthResponse(result auth.AuthResult) authResponse {
	return authResponse{
		AccessToken: result.Tokens.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.Tokens.ExpiresIn,
		User:        userResponseFrom(result.User),
	}
}

func userResponseFrom(user auth.User) userResponse {
	return userResponse{
		ID:              user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		EmailVerifiedAt: user.EmailVerifiedAt,
	}
}

func (h *AuthHandler) getRefreshToken(c echo.Context) string {
	cookie, err := c.Cookie(h.Config.RefreshCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func (h *AuthHandler) setRefreshCookie(c echo.Context, value string, expiresAt time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = h.Config.RefreshCookieName
	cookie.Value = value
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = h.Config.RefreshCookieSecure
	cookie.SameSite = parseSameSite(h.Config.RefreshCookieSameSite)
	cookie.Expires = expiresAt
	if h.Config.RefreshCookieDomain != "" {
		cookie.Domain = h.Config.RefreshCookieDomain
	}
	c.SetCookie(cookie)
}

func (h *AuthHandler) clearRefreshCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = h.Config.RefreshCookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = h.Config.RefreshCookieSecure
	cookie.SameSite = parseSameSite(h.Config.RefreshCookieSameSite)
	cookie.Expires = time.Now().UTC().Add(-time.Hour)
	if h.Config.RefreshCookieDomain != "" {
		cookie.Domain = h.Config.RefreshCookieDomain
	}
	c.SetCookie(cookie)
}

func parseSameSite(value string) http.SameSite {
	switch value {
	case "None", "none":
		return http.SameSiteNoneMode
	case "Strict", "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func isAllowedRedirect(raw string, allowlist []string) bool {
	if len(allowlist) == 0 {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	origin := parsed.Scheme + "://" + parsed.Host
	for _, allowed := range allowlist {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if allowed == raw || allowed == origin {
			return true
		}
		allowedParsed, err := url.Parse(allowed)
		if err == nil && allowedParsed.Scheme != "" && allowedParsed.Host != "" {
			if allowedParsed.Scheme+"://"+allowedParsed.Host == origin {
				return true
			}
		}
	}
	return false
}
