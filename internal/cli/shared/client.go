package shared

import (
	"context"
	"fmt"
	"sync"

	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/auth"
	"github.com/imesart/apple-ads-cli/internal/config"
)

// Package-level client singleton. A CLI process runs one command then exits,
// so a lazily-initialized shared client avoids threading a client through every
// command constructor while remaining safe for concurrent use via sync.Once.
// Tests swap these globals through SetClientForTesting.
var (
	clientOnce  = &sync.Once{}
	cachedState *clientState
	clientErr   error
)

type clientState struct {
	client     *apiPkg.Client
	tokenStore *auth.TokenStore
	cfg        *config.Profile
}

func loadConfiguredProfile() (*config.Profile, error) {
	cfg, err := config.Load(Profile())
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if OrgID() != "" {
		cfg.OrgID = OrgID()
	}

	cfg.ApplyEnv()
	return cfg, nil
}

// GetClient returns a configured API client, creating it on first call.
func GetClient() (*apiPkg.Client, error) {
	clientOnce.Do(func() {
		timeout, err := config.SelectedTimeout(apiPkg.DefaultTimeout)
		if err != nil {
			clientErr = fmt.Errorf("config validation: %w", err)
			return
		}

		cfg, err := loadConfiguredProfile()
		if err != nil {
			clientErr = err
			return
		}

		if err := cfg.Validate(); err != nil {
			clientErr = fmt.Errorf("config validation: %w", err)
			return
		}

		tokenStore := auth.NewTokenStore(
			cfg.TeamID,
			cfg.ClientID,
			cfg.KeyID,
			cfg.PrivateKeyPath,
			config.DefaultTokenCachePath(),
		)

		client := apiPkg.NewClient(
			tokenStore.GetToken,
			tokenStore.Invalidate,
			cfg.OrgID,
			Verbose(),
		)
		client.SetTimeout(timeout)
		auth.SetHTTPClientTimeout(timeout)

		cachedState = &clientState{
			client:     client,
			tokenStore: tokenStore,
			cfg:        cfg,
		}
	})

	if clientErr != nil {
		return nil, clientErr
	}
	return cachedState.client, nil
}

// SetClientForTesting overrides the cached client state for tests.
func SetClientForTesting(client *apiPkg.Client, cfg *config.Profile) func() {
	prevOnce := clientOnce
	prevState := cachedState
	prevErr := clientErr

	clientOnce = &sync.Once{}
	clientOnce.Do(func() {})
	cachedState = &clientState{
		client: client,
		cfg:    cfg,
	}
	clientErr = nil

	return func() {
		clientOnce = prevOnce
		cachedState = prevState
		clientErr = prevErr
	}
}

// GetConfig returns the loaded config.
func GetConfig() (*config.Profile, error) {
	if _, err := GetClient(); err != nil {
		return nil, err
	}
	return cachedState.cfg, nil
}

// LoadConfig returns the selected profile with flag and environment overrides
// applied, without requiring auth validation.
func LoadConfig() (*config.Profile, error) {
	if cachedState != nil && cachedState.cfg != nil {
		return cachedState.cfg, nil
	}
	return loadConfiguredProfile()
}

// GetTokenStore returns the token store for auth retry.
func GetTokenStore() *auth.TokenStore {
	if cachedState == nil {
		return nil
	}
	return cachedState.tokenStore
}

// ContextWithTimeout returns a context with the default API timeout.
func ContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout, err := config.SelectedTimeout(apiPkg.DefaultTimeout)
	if err != nil {
		timeout = apiPkg.DefaultTimeout
	}
	return context.WithTimeout(ctx, timeout)
}
