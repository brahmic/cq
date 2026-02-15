package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

func ParseAccessToken(token string) AccessTokenClaims {
	token = strings.TrimSpace(token)
	if token == "" {
		return AccessTokenClaims{}
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return AccessTokenClaims{}
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AccessTokenClaims{}
	}

	claimsMap := map[string]any{}
	if err := json.Unmarshal(payload, &claimsMap); err != nil {
		return AccessTokenClaims{}
	}

	claims := AccessTokenClaims{
		ClientID:  strings.TrimSpace(asString(claimsMap["cid"])),
		AccountID: strings.TrimSpace(asString(claimsMap["https://api.openai.com/auth"])),
		Email:     strings.TrimSpace(asString(claimsMap["email"])),
	}

	if claims.AccountID == "" {
		claims.AccountID = strings.TrimSpace(asString(claimsMap["account_id"]))
	}
	if claims.AccountID == "" {
		claims.AccountID = strings.TrimSpace(asString(claimsMap["sub"]))
	}

	if exp, ok := asInt64(claimsMap["exp"]); ok && exp > 0 {
		claims.ExpiresAt = time.Unix(exp, 0)
	}

	return claims
}
