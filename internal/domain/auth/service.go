package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
)

type Repository interface {
	CreateUserWithPassword(ctx context.Context, email, displayName, passwordHash string) (User, error)
	CreateUserWithIdentity(ctx context.Context, params CreateIdentityParams) (User, AuthIdentity, error)
	CreateAuthIdentity(ctx context.Context, params CreateIdentityParams) (AuthIdentity, error)
	UpdateAuthIdentityProfile(ctx context.Context, userID, provider, displayName, avatarURL string) error
	GetLatestAuthIdentityForUser(ctx context.Context, userID string) (AuthIdentity, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, userID string) (User, error)
	GetAuthSecretHash(ctx context.Context, userID string) (string, error)
	CreateAuthSession(ctx context.Context, params CreateSessionParams) (AuthSession, error)
	GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (AuthSession, error)
	RotateAuthSession(ctx context.Context, params RotateSessionParams) (AuthSession, User, error)
	RevokeAuthSession(ctx context.Context, refreshTokenHash string) error
	CreateOAuthState(ctx context.Context, provider, state, codeVerifier, redirectURI string) (OAuthState, error)
	GetOAuthState(ctx context.Context, state string) (OAuthState, error)
	MarkOAuthStateUsed(ctx context.Context, stateID string) error
	GetAuthIdentityByProvider(ctx context.Context, provider, providerUserID string) (AuthIdentity, error)
	UpdateAuthIdentityLogin(ctx context.Context, userID, provider string) error
	MarkUserEmailVerified(ctx context.Context, userID string) error
}

type CreateSessionParams struct {
	UserID               string
	RefreshTokenHash     string
	ExpiresAt            time.Time
	UserAgent            string
	IP                   string
	DeviceName           string
	RotatedFromSessionID *string
	IsPersistent         bool
}

type RotateSessionParams struct {
	OldRefreshTokenHash string
	NewRefreshTokenHash string
	ExpiresAt           time.Time
	UserAgent           string
	IP                  string
	IsPersistent        bool
}

type CreateIdentityParams struct {
	UserID          string
	Provider        string
	ProviderUserID  string
	Email           string
	EmailVerified   bool
	EmailVerifiedAt *time.Time
	DisplayName     string
	AvatarURL       string
}

type Service struct {
	repo        Repository
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	refreshSessionTTL time.Duration
	now         func() time.Time
	providers   map[string]OAuthProvider
}

type ServiceConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	RefreshSessionTTL time.Duration
	Providers  map[string]OAuthProvider
}

func NewService(repo Repository, cfg ServiceConfig) *Service {
	return &Service{
		repo:       repo,
		jwtSecret: []byte(cfg.JWTSecret),
		accessTTL: cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		refreshSessionTTL: cfg.RefreshSessionTTL,
		now:        time.Now().UTC,
		providers:  cfg.Providers,
	}
}

func (s *Service) SignUp(ctx context.Context, email, password, displayName, userAgent, ip string) (AuthResult, error) {
	email = normalizeEmail(email)
	details := map[string]string{}
	if !isValidEmail(email) {
		details["email"] = "invalid"
	}
	if !ValidatePassword(password) {
		details["password"] = "min_length"
	}
	if len(details) > 0 {
		return AuthResult{}, NewError("VALIDATION_ERROR", "Validacao falhou", details)
	}

	hash, err := HashPassword(password)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := s.repo.CreateUserWithPassword(ctx, email, displayName, hash)
	if err != nil {
		return AuthResult{}, err
	}

	return s.createSession(ctx, user, userAgent, ip, nil, true)
}

func (s *Service) Login(ctx context.Context, email, password, userAgent, ip string, remember bool) (AuthResult, error) {
	email = normalizeEmail(email)
	if !isValidEmail(email) || password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if !user.IsActive {
		return AuthResult{}, ErrUserInactive
	}

	hash, err := s.repo.GetAuthSecretHash(ctx, user.ID)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if !ComparePassword(hash, password) {
		return AuthResult{}, ErrInvalidCredentials
	}

	if err := s.repo.UpdateAuthIdentityLogin(ctx, user.ID, "password"); err != nil {
		return AuthResult{}, err
	}

	return s.createSession(ctx, user, userAgent, ip, nil, remember)
}

func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (AuthResult, error) {
	if refreshToken == "" {
		return AuthResult{}, ErrRefreshRevoked
	}

	sessionMeta, err := s.repo.GetAuthSessionByRefreshTokenHash(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrRefreshRevoked) {
			return AuthResult{}, ErrRefreshRevoked
		}
		return AuthResult{}, err
	}
	if sessionMeta.RevokedAt != nil || s.now().After(sessionMeta.ExpiresAt) {
		return AuthResult{}, ErrRefreshRevoked
	}

	newRefreshToken, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	refreshTTL := s.refreshSessionTTL
	if sessionMeta.IsPersistent {
		refreshTTL = s.refreshTTL
	}

	session, user, err := s.repo.RotateAuthSession(ctx, RotateSessionParams{
		OldRefreshTokenHash: hashToken(refreshToken),
		NewRefreshTokenHash: refreshHash,
		ExpiresAt:           s.now().Add(refreshTTL),
		UserAgent:           userAgent,
		IP:                  ip,
		IsPersistent:        sessionMeta.IsPersistent,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrRefreshRevoked) {
			return AuthResult{}, ErrRefreshRevoked
		}
		return AuthResult{}, err
	}
	if !user.IsActive {
		return AuthResult{}, ErrUserInactive
	}

	identity, err := s.repo.GetLatestAuthIdentityForUser(ctx, user.ID)
	if err == nil {
		if identity.DisplayName != "" {
			user.DisplayName = identity.DisplayName
		}
		if identity.AvatarURL != "" {
			user.AvatarURL = identity.AvatarURL
		}
	}

	accessToken, expiresIn, err := GenerateAccessToken(s.jwtSecret, user.ID, session.ID, s.accessTTL)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Tokens: AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    expiresIn,
			SessionID:    session.ID,
		},
		Session: session,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrRefreshRevoked
	}
	if err := s.repo.RevokeAuthSession(ctx, hashToken(refreshToken)); err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrRefreshRevoked) {
			return ErrRefreshRevoked
		}
		return err
	}
	return nil
}

func (s *Service) Me(ctx context.Context, userID string) (User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}
	if !user.IsActive {
		return User{}, ErrUserInactive
	}
	return user, nil
}

func (s *Service) StartOAuth(ctx context.Context, provider, redirectURI string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", ErrNotImplemented
	}

	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		return "", err
	}
	stateToken, err := randomToken(20)
	if err != nil {
		return "", err
	}

	_, err = s.repo.CreateOAuthState(ctx, provider, stateToken, verifier, redirectURI)
	if err != nil {
		return "", err
	}

	authURL, err := p.AuthorizeURL(stateToken, challenge, redirectURI)
	if err != nil {
		return "", ErrNotImplemented
	}

	return authURL, nil
}

func (s *Service) HandleOAuthCallback(ctx context.Context, provider, code, state, userAgent, ip string) (AuthResult, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return AuthResult{}, "", ErrNotImplemented
	}

	saved, err := s.repo.GetOAuthState(ctx, state)
	if err != nil {
		return AuthResult{}, "", ErrOAuthStateInvalid
	}
	if saved.Provider != provider || saved.UsedAt != nil || s.now().After(saved.ExpiresAt) {
		return AuthResult{}, "", ErrOAuthStateInvalid
	}
	if err := s.repo.MarkOAuthStateUsed(ctx, saved.ID); err != nil {
		return AuthResult{}, "", ErrOAuthStateInvalid
	}

	profile, err := p.Exchange(ctx, code, saved.CodeVerifier, saved.RedirectURI)
	if err != nil {
		if _, ok := err.(*Error); ok {
			return AuthResult{}, "", err
		}
		return AuthResult{}, "", ErrOAuthProvider
	}

	result, err := s.handleOAuthProfile(ctx, provider, profile, userAgent, ip)
	if err != nil {
		return AuthResult{}, "", err
	}
	return result, saved.RedirectURI, nil
}

func (s *Service) handleOAuthProfile(ctx context.Context, provider string, profile OAuthProfile, userAgent, ip string) (AuthResult, error) {
	if profile.ProviderUserID == "" {
		return AuthResult{}, ErrOAuthProvider
	}

	identity, err := s.repo.GetAuthIdentityByProvider(ctx, provider, profile.ProviderUserID)
	if err == nil {
		user, err := s.repo.GetUserByID(ctx, identity.UserID)
		if err != nil {
			return AuthResult{}, err
		}
		if !user.IsActive {
			return AuthResult{}, ErrUserInactive
		}
		if err := s.repo.UpdateAuthIdentityLogin(ctx, user.ID, provider); err != nil {
			return AuthResult{}, err
		}
		if err := s.repo.UpdateAuthIdentityProfile(ctx, user.ID, provider, profile.DisplayName, profile.AvatarURL); err != nil {
			return AuthResult{}, err
		}
		if profile.DisplayName != "" {
			user.DisplayName = profile.DisplayName
		}
		if profile.AvatarURL != "" {
			user.AvatarURL = profile.AvatarURL
		}
		return s.createSession(ctx, user, userAgent, ip, nil, true)
	}

	if err != nil && !errors.Is(err, ErrNotFound) {
		return AuthResult{}, err
	}

	if profile.Email == "" {
		return AuthResult{}, ErrOAuthEmailRequired
	}

	verified := profile.EmailVerified
	userEmail := normalizeEmail(profile.Email)

	user, err := s.repo.GetUserByEmail(ctx, userEmail)
	if err == nil {
		if user.EmailVerifiedAt == nil && !verified {
			return AuthResult{}, ErrEmailNotVerified
		}
		if user.EmailVerifiedAt == nil && verified {
			if err := s.repo.MarkUserEmailVerified(ctx, user.ID); err != nil {
				return AuthResult{}, err
			}
		}

		_, err := s.repo.CreateAuthIdentity(ctx, CreateIdentityParams{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: profile.ProviderUserID,
			Email:          userEmail,
			EmailVerified:  verified,
			DisplayName:    profile.DisplayName,
			AvatarURL:      profile.AvatarURL,
		})
		if err != nil {
			return AuthResult{}, err
		}
		if profile.DisplayName != "" {
			user.DisplayName = profile.DisplayName
		}
		if profile.AvatarURL != "" {
			user.AvatarURL = profile.AvatarURL
		}

		return s.createSession(ctx, user, userAgent, ip, nil, true)
	}

	if err != nil && !errors.Is(err, ErrNotFound) {
		return AuthResult{}, err
	}

	if !verified {
		return AuthResult{}, ErrEmailNotVerified
	}

	verifiedAt := s.now()
	user, _, err = s.repo.CreateUserWithIdentity(ctx, CreateIdentityParams{
		Provider:        provider,
		ProviderUserID:  profile.ProviderUserID,
		Email:           userEmail,
		EmailVerified:   verified,
		EmailVerifiedAt: &verifiedAt,
		DisplayName:     profile.DisplayName,
		AvatarURL:       profile.AvatarURL,
	})
	if err != nil {
		return AuthResult{}, err
	}

	return s.createSession(ctx, user, userAgent, ip, nil, true)
}

func (s *Service) ParseAccessToken(token string) (AccessTokenClaims, error) {
	if token == "" {
		return AccessTokenClaims{}, errors.New("missing token")
	}
	return ParseAccessToken(s.jwtSecret, token)
}

func (s *Service) createSession(ctx context.Context, user User, userAgent, ip string, rotatedFrom *string, isPersistent bool) (AuthResult, error) {
	refreshToken, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	refreshTTL := s.refreshSessionTTL
	if isPersistent {
		refreshTTL = s.refreshTTL
	}

	session, err := s.repo.CreateAuthSession(ctx, CreateSessionParams{
		UserID:               user.ID,
		RefreshTokenHash:     refreshHash,
		ExpiresAt:            s.now().Add(refreshTTL),
		UserAgent:            userAgent,
		IP:                   ip,
		RotatedFromSessionID: rotatedFrom,
		IsPersistent:         isPersistent,
	})
	if err != nil {
		return AuthResult{}, err
	}

	accessToken, expiresIn, err := GenerateAccessToken(s.jwtSecret, user.ID, session.ID, s.accessTTL)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Tokens: AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    expiresIn,
			SessionID:    session.ID,
		},
		Session: session,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
