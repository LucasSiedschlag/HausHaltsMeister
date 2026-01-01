package auth

import "golang.org/x/crypto/bcrypt"

const minPasswordLength = 8

func ValidatePassword(password string) bool {
	return len(password) >= minPasswordLength
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
