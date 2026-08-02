package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	EmailsURL    string
	Scopes       string
}

type OAuthProviderConfig struct {
	Google          OAuthConfig
	GitHub          OAuthConfig
	Slack           OAuthConfig
	StateSigningKey []byte
}

type oauthStateClaims struct {
	jwt.RegisteredClaims
	Provider      string `json:"provider"`
	RedirectTo    string `json:"redirect_to"`
	Linking       bool   `json:"linking"`
	CodeChallenge string `json:"code_challenge"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type githubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type slackUserInfo struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	Picture           string `json:"picture"`
	PreferredUsername string `json:"https://slack.com/team_name"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

func getOAuthConfigs(cfg *OAuthProviderConfig) map[string]OAuthConfig {
	return map[string]OAuthConfig{
		"google": cfg.Google,
		"github": cfg.GitHub,
		"slack":  cfg.Slack,
	}
}

func (s *AuthService) GetOAuthURL(provider, redirectTo string, linking bool) (string, string, error) {
	configs := getOAuthConfigs(s.oauthConfigs)
	oc, ok := configs[provider]
	if !ok {
		return "", "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	codeVerifier := generateCodeVerifier()
	codeChallenge := computeCodeChallenge(codeVerifier)

	stateID := generateID("state")
	stateJWT, err := s.jwtManager.SignClaims(&oauthStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   stateID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Provider:      provider,
		RedirectTo:    redirectTo,
		Linking:       linking,
		CodeChallenge: codeChallenge,
	})
	if err != nil {
		return "", "", fmt.Errorf("signing state jwt: %w", err)
	}

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		oc.AuthURL,
		url.QueryEscape(oc.ClientID),
		url.QueryEscape(oc.CallbackURL),
		url.QueryEscape(oc.Scopes),
		url.QueryEscape(stateJWT),
		url.QueryEscape(codeChallenge),
	)

	return authURL, stateJWT, nil
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider, code, state, codeVerifier string) (*User, string, string, bool, bool, error) {
	configs := getOAuthConfigs(s.oauthConfigs)
	oc, ok := configs[provider]
	if !ok {
		return nil, "", "", false, false, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	claims := &oauthStateClaims{}
	if err := s.jwtManager.ValidateClaims(state, claims); err != nil {
		return nil, "", "", false, false, ErrOAuthStateMismatch
	}

	if claims.Provider != provider {
		return nil, "", "", false, false, ErrOAuthStateMismatch
	}

	if claims.CodeChallenge != computeCodeChallenge(codeVerifier) {
		return nil, "", "", false, false, ErrOAuthStateMismatch
	}

	tokens, err := exchangeCode(oc, code, codeVerifier)
	if err != nil {
		return nil, "", "", false, false, fmt.Errorf("exchanging code for tokens: %w", err)
	}

	providerUserID, email, emailVerified, displayName, avatarURL, err := fetchUserInfo(provider, tokens.AccessToken)
	if err != nil {
		return nil, "", "", false, false, fmt.Errorf("fetching user info: %w", err)
	}

	if !emailVerified {
		return nil, "", "", false, false, ErrEmailNotVerified
	}

	existingIdentity, _ := s.oauthRepo.GetByProvider(ctx, provider, providerUserID)
	if existingIdentity != nil {
		user, err := s.userRepo.GetByID(ctx, existingIdentity.UserID)
		if err != nil {
			return nil, "", "", false, false, err
		}

		_ = s.oauthRepo.UpdateLastUsed(ctx, existingIdentity.ID)
		accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
		if err != nil {
			return nil, "", "", false, false, fmt.Errorf("generating access token: %w", err)
		}
		refreshToken, err := s.generateRefreshToken(ctx, user.ID, "")
		if err != nil {
			return nil, "", "", false, false, err
		}
		return user, accessToken, refreshToken, false, false, nil
	}

	existingUser, _ := s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		if existingUser.PasswordHash != "" {
			return existingUser, "", "", false, true, nil
		}
	}

	user := &User{
		Email:       email,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Role:        "user",
		IsActive:    true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", false, false, fmt.Errorf("creating user: %w", err)
	}

	created, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", false, false, err
	}

	identity := &OAuthIdentity{
		UserID:                created.ID,
		Provider:              provider,
		ProviderUserID:        providerUserID,
		ProviderEmail:         email,
		AccessTokenEncrypted:  tokens.AccessToken,
		RefreshTokenEncrypted: tokens.RefreshToken,
	}
	if tokens.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
		identity.ExpiresAt = &exp
	}
	if err := s.oauthRepo.Create(ctx, identity); err != nil {
		return nil, "", "", false, false, fmt.Errorf("creating oauth identity: %w", err)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(created.ID, created.Email, created.Role)
	if err != nil {
		return nil, "", "", false, false, fmt.Errorf("generating access token: %w", err)
	}
	refreshToken, err := s.generateRefreshToken(ctx, created.ID, "")
	if err != nil {
		return nil, "", "", false, false, err
	}

	return created, accessToken, refreshToken, true, false, nil
}

func (s *AuthService) LinkOAuthIdentity(ctx context.Context, userID, provider, code, state string) error {
	claims := &oauthStateClaims{}
	if err := s.jwtManager.ValidateClaims(state, claims); err != nil {
		return ErrOAuthStateMismatch
	}

	if !claims.Linking {
		return fmt.Errorf("state was not created for linking")
	}

	configs := getOAuthConfigs(s.oauthConfigs)
	oc, ok := configs[provider]
	if !ok {
		return fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	tokens, err := exchangeCode(oc, code, "")
	if err != nil {
		return fmt.Errorf("exchanging code for tokens: %w", err)
	}

	providerUserID, email, _, _, _, err := fetchUserInfo(provider, tokens.AccessToken)
	if err != nil {
		return fmt.Errorf("fetching user info: %w", err)
	}

	existing, _ := s.oauthRepo.GetByProvider(ctx, provider, providerUserID)
	if existing != nil {
		return ErrProviderAlreadyLinked
	}

	identity := &OAuthIdentity{
		UserID:                userID,
		Provider:              provider,
		ProviderUserID:        providerUserID,
		ProviderEmail:         email,
		AccessTokenEncrypted:  tokens.AccessToken,
		RefreshTokenEncrypted: tokens.RefreshToken,
	}
	if tokens.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
		identity.ExpiresAt = &exp
	}

	return s.oauthRepo.Create(ctx, identity)
}

func (s *AuthService) UnlinkOAuthIdentity(ctx context.Context, userID, provider string) error {
	identities, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	var target *OAuthIdentity
	var otherCount int
	for _, id := range identities {
		if id.Provider == provider {
			target = id
		} else {
			otherCount++
		}
	}
	if target == nil {
		return fmt.Errorf("no linked identity for provider %s", provider)
	}

	if otherCount == 0 && user.PasswordHash == "" {
		return ErrLastAuthMethod
	}

	return s.oauthRepo.Delete(ctx, target.ID)
}

func (s *AuthService) ListLinkedIdentities(ctx context.Context, userID string) ([]*OAuthIdentity, error) {
	return s.oauthRepo.GetByUserID(ctx, userID)
}

func exchangeCode(cfg OAuthConfig, code, codeVerifier string) (*tokenResponse, error) {
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {cfg.CallbackURL},
		"grant_type":    {"authorization_code"},
	}
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	resp, err := http.PostForm(cfg.TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	tokens := &tokenResponse{}
	if err := json.Unmarshal(body, tokens); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}
	return tokens, nil
}

func fetchUserInfo(provider, accessToken string) (userID, email string, emailVerified bool, displayName, avatarURL string, err error) {
	switch provider {
	case "google":
		return fetchGoogleUserInfo(accessToken)
	case "github":
		return fetchGitHubUserInfo(accessToken)
	case "slack":
		return fetchSlackUserInfo(accessToken)
	default:
		return "", "", false, "", "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func fetchGoogleUserInfo(accessToken string) (string, string, bool, string, string, error) {
	info, err := doGet[googleUserInfo]("https://openidconnect.googleapis.com/v1/userinfo", accessToken)
	if err != nil {
		return "", "", false, "", "", err
	}
	return info.Sub, info.Email, info.EmailVerified, info.Name, info.Picture, nil
}

func fetchGitHubUserInfo(accessToken string) (string, string, bool, string, string, error) {
	user, err := doGet[githubUserInfo]("https://api.github.com/user", accessToken)
	if err != nil {
		return "", "", false, "", "", err
	}

	userID := fmt.Sprintf("%d", user.ID)
	displayName := user.Name
	if displayName == "" {
		displayName = user.Login
	}

	if user.Email != "" && user.Email != "null" {
		return userID, user.Email, true, displayName, user.AvatarURL, nil
	}

	emails, err := doGet[[]githubEmail]("https://api.github.com/user/emails", accessToken)
	if err != nil {
		return userID, "", false, displayName, user.AvatarURL, nil
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return userID, e.Email, true, displayName, user.AvatarURL, nil
		}
	}
	if len(emails) > 0 {
		return userID, emails[0].Email, emails[0].Verified, displayName, user.AvatarURL, nil
	}

	return userID, "", false, displayName, user.AvatarURL, nil
}

func fetchSlackUserInfo(accessToken string) (string, string, bool, string, string, error) {
	type slackResponse struct {
		OK    bool          `json:"ok"`
		User  slackUserInfo `json:"user,omitempty"`
		Error string        `json:"error,omitempty"`
	}

	req, _ := http.NewRequest("GET", "https://slack.com/api/openid.connect.userInfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", false, "", "", fmt.Errorf("slack userinfo request: %w", err)
	}
	defer resp.Body.Close()

	var sr slackResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", "", false, "", "", fmt.Errorf("parsing slack response: %w", err)
	}
	if !sr.OK {
		return "", "", false, "", "", fmt.Errorf("slack api error: %s", sr.Error)
	}

	u := sr.User
	return u.Sub, u.Email, u.EmailVerified, u.Name, u.Picture, nil
}

func doGet[T any](url, accessToken string) (T, error) {
	var zero T
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	if strings.Contains(url, "api.github.com") {
		req.Header.Set("User-Agent", "SchemaHub")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("reading response from %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("%s returned %d: %s", url, resp.StatusCode, string(body))
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, fmt.Errorf("parsing response from %s: %w", url, err)
	}
	return result, nil
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func computeCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateID(prefix string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
