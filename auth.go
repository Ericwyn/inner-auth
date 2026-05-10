package main

import (
	"fmt"

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
		if !totp.Validate(totpCode, a.config.TOTPToken) {
			return fmt.Errorf("invalid TOTP code")
		}
	}

	return nil
}
