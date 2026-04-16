package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// BuildJWT creates an ES256-signed JWT assertion for Apple Ads OAuth2.
// Parameters:
//   - teamID: iss claim
//   - clientID: sub claim
//   - keyID: kid header
//   - privateKeyPath: path to P-256 PEM file
//   - expiry: token lifetime (max 180 days)
func BuildJWT(teamID, clientID, keyID, privateKeyPath string, expiry time.Duration) (string, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("reading private key: %w", err)
	}
	return BuildJWTFromPEM(teamID, clientID, keyID, keyData, expiry)
}

func BuildJWTFromPEM(teamID, clientID, keyID string, pemData []byte, expiry time.Duration) (string, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try EC key format
		key, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("parsing private key: %w", err)
		}
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an ECDSA key")
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    teamID,
		Subject:   clientID,
		Audience:  jwt.ClaimStrings{"https://appleid.apple.com"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	return token.SignedString(ecKey)
}
