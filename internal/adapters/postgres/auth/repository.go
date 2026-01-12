package auth

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	oauthStateTTL = 10 * time.Minute
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) CreateUserWithPassword(ctx context.Context, email, displayName, passwordHash string) (auth.User, error) {
	var user auth.User
	var ledgerID string

	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO users (email, display_name)
			VALUES ($1, $2)
			RETURNING id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
		`, email, displayName)
		if err := scanUser(row, &user); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO auth_secrets (user_id, password_hash)
			VALUES ($1, $2)
		`, user.ID, passwordHash)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO auth_identities (user_id, provider, provider_user_id, email, email_verified, last_login_at)
			VALUES ($1, 'password', $2, $3, false, now())
		`, user.ID, user.ID, user.Email)
		if err != nil {
			return err
		}

		row = tx.QueryRow(ctx, `
			INSERT INTO ledgers (owner_user_id, name, currency_code)
			VALUES ($1, 'Pessoal', 'BRL')
			RETURNING id
		`, user.ID)
		if err := row.Scan(&ledgerID); err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO ledger_members (ledger_id, user_id, role)
			VALUES ($1, $2, 'owner')
		`, ledgerID, user.ID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO accounts (ledger_id, name, type, nature, is_active)
			VALUES
				($1, 'Conta Corrente', 'current', 'asset', true),
				($1, 'Wallet/Pessoal', 'wallet', 'asset', true)
		`, ledgerID)
		return err
	})

	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return auth.User{}, auth.ErrDuplicateEmail
		}
		return auth.User{}, err
	}

	return user, nil
}

func (r *Repository) CreateUserWithIdentity(ctx context.Context, params auth.CreateIdentityParams) (auth.User, auth.AuthIdentity, error) {
	var user auth.User
	var identity auth.AuthIdentity
	var ledgerID string

	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO users (email, display_name, avatar_url, email_verified_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
		`, params.Email, params.DisplayName, params.AvatarURL, params.EmailVerifiedAt)
		if err := scanUser(row, &user); err != nil {
			return err
		}

		row = tx.QueryRow(ctx, `
			INSERT INTO auth_identities (user_id, provider, provider_user_id, email, display_name, avatar_url, email_verified, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, now())
			RETURNING id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at
		`, user.ID, params.Provider, params.ProviderUserID, params.Email, params.DisplayName, params.AvatarURL, params.EmailVerified)
		if err := scanIdentity(row, &identity); err != nil {
			return err
		}

		row = tx.QueryRow(ctx, `
			INSERT INTO ledgers (owner_user_id, name, currency_code)
			VALUES ($1, 'Pessoal', 'BRL')
			RETURNING id
		`, user.ID)
		if err := row.Scan(&ledgerID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO ledger_members (ledger_id, user_id, role)
			VALUES ($1, $2, 'owner')
		`, ledgerID, user.ID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO accounts (ledger_id, name, type, nature, is_active)
			VALUES
				($1, 'Conta Corrente', 'current', 'asset', true),
				($1, 'Wallet/Pessoal', 'wallet', 'asset', true)
		`, ledgerID)
		return err
	})

	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return auth.User{}, auth.AuthIdentity{}, auth.ErrDuplicateEmail
		}
		return auth.User{}, auth.AuthIdentity{}, err
	}

	return user, identity, nil
}

func (r *Repository) CreateAuthIdentity(ctx context.Context, params auth.CreateIdentityParams) (auth.AuthIdentity, error) {
	var identity auth.AuthIdentity
	row := r.pool.QueryRow(ctx, `
		INSERT INTO auth_identities (user_id, provider, provider_user_id, email, display_name, avatar_url, email_verified, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		RETURNING id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at
	`, params.UserID, params.Provider, params.ProviderUserID, params.Email, params.DisplayName, params.AvatarURL, params.EmailVerified)
	if err := scanIdentity(row, &identity); err != nil {
		return auth.AuthIdentity{}, err
	}
	return identity, nil
}

func (r *Repository) UpdateAuthIdentityLogin(ctx context.Context, userID, provider string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_identities
		SET last_login_at = now(), updated_at = now()
		WHERE user_id = $1 AND provider = $2
	`, userID, provider)
	return err
}

func (r *Repository) UpdateAuthIdentityProfile(ctx context.Context, userID, provider, displayName, avatarURL string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_identities
		SET display_name = CASE WHEN $3 <> '' THEN $3 ELSE display_name END,
		    avatar_url = CASE WHEN $4 <> '' THEN $4 ELSE avatar_url END,
		    updated_at = now()
		WHERE user_id = $1 AND provider = $2
	`, userID, provider, displayName, avatarURL)
	return err
}

func (r *Repository) GetLatestAuthIdentityForUser(ctx context.Context, userID string) (auth.AuthIdentity, error) {
	var identity auth.AuthIdentity
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at
		FROM auth_identities
		WHERE user_id = $1
		ORDER BY last_login_at DESC NULLS LAST, created_at DESC
		LIMIT 1
	`, userID)
	if err := scanIdentity(row, &identity); err != nil {
		return auth.AuthIdentity{}, err
	}
	return identity, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	var user auth.User
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
	if err := scanUser(row, &user); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (auth.User, error) {
	var user auth.User
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	if err := scanUser(row, &user); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (r *Repository) MarkUserEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users
		SET email_verified_at = now(), updated_at = now()
		WHERE id = $1 AND email_verified_at IS NULL
	`, userID)
	return err
}

func (r *Repository) GetAuthSecretHash(ctx context.Context, userID string) (string, error) {
	var hash string
	row := r.pool.QueryRow(ctx, `
		SELECT password_hash
		FROM auth_secrets
		WHERE user_id = $1
	`, userID)
	if err := row.Scan(&hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", auth.ErrNotFound
		}
		return "", err
	}
	return hash, nil
}

func (r *Repository) CreateAuthSession(ctx context.Context, params auth.CreateSessionParams) (auth.AuthSession, error) {
	var session auth.AuthSession
	row := r.pool.QueryRow(ctx, `
		INSERT INTO auth_sessions (user_id, refresh_token_hash, is_persistent, expires_at, user_agent, ip, device_name, rotated_from_session_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, expires_at, revoked_at, is_persistent
	`, params.UserID, params.RefreshTokenHash, params.IsPersistent, params.ExpiresAt, params.UserAgent, params.IP, params.DeviceName, params.RotatedFromSessionID)
	if err := scanSession(row, &session); err != nil {
		return auth.AuthSession{}, err
	}
	return session, nil
}

func (r *Repository) GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (auth.AuthSession, error) {
	var session auth.AuthSession
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked_at, is_persistent
		FROM auth_sessions
		WHERE refresh_token_hash = $1
	`, refreshTokenHash)
	if err := scanSession(row, &session); err != nil {
		return auth.AuthSession{}, err
	}
	return session, nil
}

func (r *Repository) RotateAuthSession(ctx context.Context, params auth.RotateSessionParams) (auth.AuthSession, auth.User, error) {
	var session auth.AuthSession
	var user auth.User

	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			SELECT id, user_id, expires_at, revoked_at, is_persistent
			FROM auth_sessions
			WHERE refresh_token_hash = $1
			FOR UPDATE
		`, params.OldRefreshTokenHash)
		if err := scanSession(row, &session); err != nil {
			return err
		}
		if session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt) {
			return auth.ErrRefreshRevoked
		}

		row = tx.QueryRow(ctx, `
			SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
			FROM users
			WHERE id = $1
		`, session.UserID)
		if err := scanUser(row, &user); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			UPDATE auth_sessions
			SET revoked_at = now(), updated_at = now()
			WHERE id = $1
		`, session.ID)
		if err != nil {
			return err
		}

		row = tx.QueryRow(ctx, `
			INSERT INTO auth_sessions (user_id, refresh_token_hash, is_persistent, expires_at, user_agent, ip, rotated_from_session_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, user_id, expires_at, revoked_at, is_persistent
		`, session.UserID, params.NewRefreshTokenHash, params.IsPersistent, params.ExpiresAt, params.UserAgent, params.IP, session.ID)
		return scanSession(row, &session)
	})

	if err != nil {
		return auth.AuthSession{}, auth.User{}, err
	}

	return session, user, nil
}

func (r *Repository) RevokeAuthSession(ctx context.Context, refreshTokenHash string) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = now(), updated_at = now()
		WHERE refresh_token_hash = $1 AND revoked_at IS NULL
	`, refreshTokenHash)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return auth.ErrRefreshRevoked
	}
	return nil
}

func (r *Repository) ListAuthSessionsByUser(ctx context.Context, userID string) ([]auth.AuthSessionDetails, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, expires_at, revoked_at, is_persistent, COALESCE(user_agent, ''), COALESCE(ip, ''), COALESCE(device_name, ''), created_at, updated_at, rotated_from_session_id
		FROM auth_sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []auth.AuthSessionDetails
	for rows.Next() {
		var session auth.AuthSessionDetails
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.ExpiresAt,
			&session.RevokedAt,
			&session.IsPersistent,
			&session.UserAgent,
			&session.IP,
			&session.DeviceName,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.RotatedFromSession,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *Repository) RevokeAuthSessionByID(ctx context.Context, userID, sessionID string) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, sessionID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return auth.ErrNotFound
	}
	return nil
}

func (r *Repository) RevokeAllAuthSessionsForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = now(), updated_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

func (r *Repository) CreateOAuthState(ctx context.Context, provider, state, codeVerifier, redirectURI string) (auth.OAuthState, error) {
	var saved auth.OAuthState
	row := r.pool.QueryRow(ctx, `
		INSERT INTO oauth_states (provider, state, code_verifier, redirect_uri, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, provider, state, code_verifier, COALESCE(redirect_uri, ''), expires_at, used_at
	`, provider, state, codeVerifier, redirectURI, time.Now().UTC().Add(oauthStateTTL))
	if err := scanOAuthState(row, &saved); err != nil {
		return auth.OAuthState{}, err
	}
	return saved, nil
}

func (r *Repository) GetOAuthState(ctx context.Context, state string) (auth.OAuthState, error) {
	var saved auth.OAuthState
	row := r.pool.QueryRow(ctx, `
		SELECT id, provider, state, code_verifier, COALESCE(redirect_uri, ''), expires_at, used_at
		FROM oauth_states
		WHERE state = $1
	`, state)
	if err := scanOAuthState(row, &saved); err != nil {
		return auth.OAuthState{}, err
	}
	return saved, nil
}

func (r *Repository) MarkOAuthStateUsed(ctx context.Context, stateID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE oauth_states
		SET used_at = now(), updated_at = now()
		WHERE id = $1 AND used_at IS NULL
	`, stateID)
	return err
}

func (r *Repository) GetAuthIdentityByProvider(ctx context.Context, provider, providerUserID string) (auth.AuthIdentity, error) {
	var identity auth.AuthIdentity
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at
		FROM auth_identities
		WHERE provider = $1 AND provider_user_id = $2
	`, provider, providerUserID)
	if err := scanIdentity(row, &identity); err != nil {
		return auth.AuthIdentity{}, err
	}
	return identity, nil
}

func scanUser(row pgx.Row, user *auth.User) error {
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.EmailVerifiedAt,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrNotFound
		}
		return err
	}
	return nil
}

func scanSession(row pgx.Row, session *auth.AuthSession) error {
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.IsPersistent,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrNotFound
		}
		return err
	}
	return nil
}

func scanOAuthState(row pgx.Row, state *auth.OAuthState) error {
	if err := row.Scan(
		&state.ID,
		&state.Provider,
		&state.State,
		&state.CodeVerifier,
		&state.RedirectURI,
		&state.ExpiresAt,
		&state.UsedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrNotFound
		}
		return err
	}
	return nil
}

func scanIdentity(row pgx.Row, identity *auth.AuthIdentity) error {
	if err := row.Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&identity.Email,
		&identity.DisplayName,
		&identity.AvatarURL,
		&identity.EmailVerified,
		&identity.LastLoginAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrNotFound
		}
		return err
	}
	return nil
}
