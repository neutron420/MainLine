package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/pkg/encryption"
)

type OAuthRefreshWorker struct {
	oauthRepo     domain.OAuthIdentityRepository
	userRepo      domain.UserRepository
	encryptionKey []byte
	clientSecrets map[string]string
}

func NewOAuthRefreshWorker(oauthRepo domain.OAuthIdentityRepository, userRepo domain.UserRepository, encryptionKey []byte) *OAuthRefreshWorker {
	return &OAuthRefreshWorker{
		oauthRepo:     oauthRepo,
		userRepo:      userRepo,
		encryptionKey: encryptionKey,
		clientSecrets: make(map[string]string),
	}
}

func (w *OAuthRefreshWorker) SetClientSecret(provider, secret string) {
	w.clientSecrets[provider] = secret
}

func (w *OAuthRefreshWorker) Name() string {
	return "oauth-token-refresh"
}

func (w *OAuthRefreshWorker) Interval() time.Duration {
	return 30 * time.Minute
}

type oauthRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var refreshEndpoints = map[string]string{
	"google": "https://oauth2.googleapis.com/token",
	"github": "https://github.com/login/oauth/access_token",
	"slack":  "https://slack.com/api/openid.connect.token",
}

func (w *OAuthRefreshWorker) Run(ctx context.Context) error {
	identities, err := w.oauthRepo.GetExpiringSoon(ctx, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("fetching expiring identities: %w", err)
	}

	for _, identity := range identities {
		if identity.RefreshTokenEncrypted == "" {
			continue
		}

		refreshURL, ok := refreshEndpoints[identity.Provider]
		if !ok {
			continue
		}

		decrypted, err := encryption.Decrypt(identity.RefreshTokenEncrypted, w.encryptionKey)
		if err != nil {
			continue
		}

		data := url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {string(decrypted)},
		}
		if secret := w.clientSecrets[identity.Provider]; secret != "" {
			data.Set("client_secret", secret)
		}

		resp, err := http.PostForm(refreshURL, data)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var tokens oauthRefreshResponse
		if err := json.Unmarshal(body, &tokens); err != nil {
			continue
		}

		if tokens.AccessToken == "" {
			continue
		}

		encryptedAT, err := encryption.Encrypt([]byte(tokens.AccessToken), w.encryptionKey)
		if err != nil {
			continue
		}

		newRefreshToken := identity.RefreshTokenEncrypted
		if tokens.RefreshToken != "" {
			encryptedRT, err := encryption.Encrypt([]byte(tokens.RefreshToken), w.encryptionKey)
			if err != nil {
				continue
			}
			newRefreshToken = encryptedRT
		}

		var expiresAt *time.Time
		if tokens.ExpiresIn > 0 {
			exp := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
			expiresAt = &exp
		}

		_ = w.oauthRepo.UpdateTokens(ctx, identity.ID, encryptedAT, newRefreshToken, expiresAt)
	}

	return nil
}
