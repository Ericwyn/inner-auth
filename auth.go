package main

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	config *AuthConfig
}

func NewAuthenticator(config *AuthConfig) *Authenticator {
	return &Authenticator{config: config}
}

func (a *Authenticator) IsTOTPRequired() bool {
	return a.config.TOTPToken != ""
}

func (a *Authenticator) Authenticate(username, password, totpCode string) error {
	if username != a.config.User || !a.validPassword(password) {
		return fmt.Errorf("invalid username or password")
	}

	if a.IsTOTPRequired() {
		if totpCode == "" {
			return fmt.Errorf("TOTP code is required")
		}
		valid, err := totp.ValidateCustom(totpCode, a.config.TOTPToken, time.Now(), totp.ValidateOpts{
			Period:    30,
			Skew:      2,
			Digits:    6,
			Algorithm: 0, // SHA1
		})
		if !valid {
			return fmt.Errorf("invalid TOTP code: %w", err)
		}
	}

	return nil
}

func (a *Authenticator) validPassword(password string) bool {
	if a.config.PasswordHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(a.config.PasswordHash), []byte(password)) == nil
	}

	return subtle.ConstantTimeCompare([]byte(password), []byte(a.config.Password)) == 1
}
