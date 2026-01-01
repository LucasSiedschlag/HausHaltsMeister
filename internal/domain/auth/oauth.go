package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OAuthProfile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	DisplayName    string
	AvatarURL      string
}

type OAuthProvider interface {
	Name() string
	AuthorizeURL(state, codeChallenge, redirectURI string) (string, error)
	Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (OAuthProfile, error)
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GoogleProvider struct {
	cfg    OAuthProviderConfig
	client *http.Client
}

type GitHubProvider struct {
	cfg    OAuthProviderConfig
	client *http.Client
}

func NewGoogleProvider(cfg OAuthProviderConfig) *GoogleProvider {
	return &GoogleProvider{cfg: cfg, client: defaultOAuthClient()}
}

func NewGitHubProvider(cfg OAuthProviderConfig) *GitHubProvider {
	return &GitHubProvider{cfg: cfg, client: defaultOAuthClient()}
}

func (g *GoogleProvider) Name() string { return "google" }

func (g *GoogleProvider) AuthorizeURL(state, codeChallenge, redirectURI string) (string, error) {
	if g.cfg.ClientID == "" || g.cfg.RedirectURL == "" {
		return "", errors.New("oauth provider not configured")
	}
	q := url.Values{}
	q.Set("client_id", g.cfg.ClientID)
	q.Set("redirect_uri", g.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(), nil
}

func (g *GoogleProvider) Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (OAuthProfile, error) {
	form := url.Values{}
	form.Set("client_id", g.cfg.ClientID)
	form.Set("client_secret", g.cfg.ClientSecret)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", g.cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthProfile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(req)
	if err != nil {
		return OAuthProfile{}, classifyOAuthError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return OAuthProfile{}, ErrOAuthProvider
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return OAuthProfile{}, ErrOAuthProvider
	}
	if tokenResp.AccessToken == "" {
		return OAuthProfile{}, ErrOAuthProvider
	}

	profileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return OAuthProfile{}, err
	}
	profileReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	profileResp, err := g.client.Do(profileReq)
	if err != nil {
		return OAuthProfile{}, classifyOAuthError(err)
	}
	defer profileResp.Body.Close()

	if profileResp.StatusCode < 200 || profileResp.StatusCode >= 300 {
		return OAuthProfile{}, ErrOAuthProvider
	}

	var userinfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&userinfo); err != nil {
		return OAuthProfile{}, ErrOAuthProvider
	}

	return OAuthProfile{
		ProviderUserID: userinfo.Sub,
		Email:          userinfo.Email,
		EmailVerified:  userinfo.EmailVerified,
		DisplayName:    userinfo.Name,
		AvatarURL:      userinfo.Picture,
	}, nil
}

func (g *GitHubProvider) Name() string { return "github" }

func (g *GitHubProvider) AuthorizeURL(state, codeChallenge, redirectURI string) (string, error) {
	if g.cfg.ClientID == "" || g.cfg.RedirectURL == "" {
		return "", errors.New("oauth provider not configured")
	}
	q := url.Values{}
	q.Set("client_id", g.cfg.ClientID)
	q.Set("redirect_uri", g.cfg.RedirectURL)
	q.Set("scope", "read:user user:email")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	return "https://github.com/login/oauth/authorize?" + q.Encode(), nil
}

func (g *GitHubProvider) Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (OAuthProfile, error) {
	form := url.Values{}
	form.Set("client_id", g.cfg.ClientID)
	form.Set("client_secret", g.cfg.ClientSecret)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("redirect_uri", g.cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthProfile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return OAuthProfile{}, classifyOAuthError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return OAuthProfile{}, ErrOAuthProvider
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return OAuthProfile{}, ErrOAuthProvider
	}
	if tokenResp.AccessToken == "" {
		return OAuthProfile{}, ErrOAuthProvider
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return OAuthProfile{}, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github+json")

	userResp, err := g.client.Do(userReq)
	if err != nil {
		return OAuthProfile{}, classifyOAuthError(err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode < 200 || userResp.StatusCode >= 300 {
		return OAuthProfile{}, ErrOAuthProvider
	}

	var ghUser struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&ghUser); err != nil {
		return OAuthProfile{}, ErrOAuthProvider
	}

	email := strings.TrimSpace(ghUser.Email)
	emailVerified := false
	if email == "" {
		email, emailVerified, err = g.fetchPrimaryEmail(ctx, tokenResp.AccessToken)
		if err != nil {
			return OAuthProfile{}, err
		}
	}

	return OAuthProfile{
		ProviderUserID: strconv.FormatInt(ghUser.ID, 10),
		Email:          email,
		EmailVerified:  emailVerified,
		DisplayName:    ghUser.Name,
		AvatarURL:      ghUser.AvatarURL,
	}, nil
}

func (g *GitHubProvider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", false, classifyOAuthError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = body
		return "", false, ErrOAuthProvider
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}

	for _, item := range emails {
		if item.Primary {
			return strings.TrimSpace(item.Email), item.Verified, nil
		}
	}
	for _, item := range emails {
		if item.Verified {
			return strings.TrimSpace(item.Email), item.Verified, nil
		}
	}
	return "", false, nil
}

func defaultOAuthClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func classifyOAuthError(err error) error {
	if err == nil {
		return ErrOAuthProvider
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrGatewayTimeout
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrGatewayTimeout
	}
	return ErrServiceUnavailable
}

func GeneratePKCE() (string, string, error) {
	verifier, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	challenge := codeChallenge(verifier)
	return verifier, challenge, nil
}

func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
