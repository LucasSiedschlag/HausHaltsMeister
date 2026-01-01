package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AccessTokenClaims struct {
	Sub string `json:"sub"`
	Sid string `json:"sid"`
	Jti string `json:"jti"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func GenerateAccessToken(secret []byte, userID, sessionID string, ttl time.Duration) (string, int64, error) {
	issuedAt := time.Now().UTC()
	jti, err := randomToken(16)
	if err != nil {
		return "", 0, err
	}
	claims := AccessTokenClaims{
		Sub: userID,
		Sid: sessionID,
		Jti: jti,
		Iat: issuedAt.Unix(),
		Exp: issuedAt.Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", 0, err
	}
	header := []byte(`{"alg":"HS256","typ":"JWT"}`)
	headerEnc := base64.RawURLEncoding.EncodeToString(header)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := headerEnc + "." + payloadEnc
	sig := signHS256(secret, signingInput)
	token := signingInput + "." + sig
	return token, claims.Exp - claims.Iat, nil
}

func ParseAccessToken(secret []byte, token string) (AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, errors.New("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	if !verifyHS256(secret, signingInput, parts[2]) {
		return AccessTokenClaims{}, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AccessTokenClaims{}, err
	}
	var claims AccessTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return AccessTokenClaims{}, err
	}
	if claims.Exp == 0 || time.Now().UTC().Unix() > claims.Exp {
		return AccessTokenClaims{}, errors.New("token expired")
	}
	if claims.Sub == "" {
		return AccessTokenClaims{}, errors.New("missing subject")
	}
	return claims, nil
}

func GenerateRefreshToken() (string, string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	hash := hashToken(token)
	return token, hash, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h[:])
}

func signHS256(secret []byte, message string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func verifyHS256(secret []byte, message, signature string) bool {
	expected := signHS256(secret, message)
	return hmac.Equal([]byte(expected), []byte(signature))
}
