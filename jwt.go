package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const CookieName = "inner_auth_token"

type Claims struct {
	User string `json:"user"`
	Site string `json:"site"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, user string, site string, ttlHours int) (string, error) {
	claims := Claims{
		User: user,
		Site: site,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "inner-auth",
			Subject:   user,
			Audience:  jwt.ClaimStrings{site},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(secret string, tokenString string, site string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithIssuer("inner-auth"), jwt.WithAudience(site))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Site != site {
		return nil, fmt.Errorf("invalid token site")
	}

	return claims, nil
}
