package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateTestECKey creates a P-256 EC private key and returns PEM-encoded bytes.
func generateTestECKey(t *testing.T) ([]byte, *ecdsa.PrivateKey) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating EC key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshalling EC key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
	return pemBytes, priv
}

// generateTestRSAKey creates an RSA private key and returns PEM-encoded bytes.
func generateTestRSAKey(t *testing.T) []byte {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshalling RSA key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
}

func TestBuildJWTFromPEM(t *testing.T) {
	pemData, ecKey := generateTestECKey(t)

	teamID := "TEAM123"
	clientID := "com.example.client"
	keyID := "KEY456"
	expiry := 1 * time.Hour

	tokenStr, err := BuildJWTFromPEM(teamID, clientID, keyID, pemData, expiry)
	if err != nil {
		t.Fatalf("BuildJWTFromPEM() error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("BuildJWTFromPEM() returned empty token")
	}

	// Parse the token back to verify claims and header
	parsed, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &ecKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("parsing JWT: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("parsed JWT is not valid")
	}

	// Verify header
	if alg := parsed.Method.Alg(); alg != "ES256" {
		t.Errorf("header alg = %q, want %q", alg, "ES256")
	}
	kid, ok := parsed.Header["kid"]
	if !ok {
		t.Fatal("header missing kid")
	}
	if kid != keyID {
		t.Errorf("header kid = %q, want %q", kid, keyID)
	}

	// Verify claims
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("failed to cast claims to RegisteredClaims")
	}
	if claims.Issuer != teamID {
		t.Errorf("iss = %q, want %q", claims.Issuer, teamID)
	}
	if claims.Subject != clientID {
		t.Errorf("sub = %q, want %q", claims.Subject, clientID)
	}

	// Audience should contain exactly "https://appleid.apple.com"
	aud := claims.Audience
	if len(aud) != 1 || aud[0] != "https://appleid.apple.com" {
		t.Errorf("aud = %v, want [\"https://appleid.apple.com\"]", aud)
	}

	// ExpiresAt should be after IssuedAt
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("iat or exp is nil")
	}
	if !claims.ExpiresAt.Time.After(claims.IssuedAt.Time) {
		t.Errorf("exp (%v) is not after iat (%v)", claims.ExpiresAt.Time, claims.IssuedAt.Time)
	}

	// The difference between exp and iat should be approximately the expiry duration
	diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if diff < expiry-time.Second || diff > expiry+time.Second {
		t.Errorf("exp - iat = %v, want approximately %v", diff, expiry)
	}

	// Verify the token is properly signed by trying to verify with the public key
	// (already done implicitly by jwt.Parse above, but let's be explicit)
	_, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &ecKey.PublicKey, nil
	})
	if err != nil {
		t.Errorf("token signature verification failed: %v", err)
	}
}

func TestBuildJWTFromPEM_InvalidKey(t *testing.T) {
	invalidPEM := []byte("this is not a valid PEM block")

	_, err := BuildJWTFromPEM("team", "client", "key", invalidPEM, time.Hour)
	if err == nil {
		t.Fatal("BuildJWTFromPEM() with invalid PEM should return error")
	}
	t.Logf("got expected error: %v", err)
}

func TestBuildJWTFromPEM_WrongKeyType(t *testing.T) {
	rsaPEM := generateTestRSAKey(t)

	_, err := BuildJWTFromPEM("team", "client", "key", rsaPEM, time.Hour)
	if err == nil {
		t.Fatal("BuildJWTFromPEM() with RSA key should return error")
	}
	t.Logf("got expected error: %v", err)
}
