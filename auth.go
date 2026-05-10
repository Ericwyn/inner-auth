package main

import (
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
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
	if username != a.config.User || password != a.config.Password {
		return fmt.Errorf("invalid username or password")
	}

	if a.IsTOTPRequired() {
		if totpCode == "" {
			return fmt.Errorf("TOTP code is required")
		}
		valid, _ := totp.ValidateCustom(totpCode, a.config.TOTPToken, time.Now(), totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    6,
			Algorithm: 0, // SHA1
		})
		if !valid {
			return fmt.Errorf("invalid TOTP code")
		}
	}

	return nil
}
