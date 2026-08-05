package domain

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/schemahub/backend/internal/pkg/jwt"
)

func testKeyPair(t *testing.T) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshaling public key: %v", err)
	}
	publicPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}))
	return privatePEM, publicPEM
}

func newTestManager(t *testing.T) *jwt.Manager {
	t.Helper()
	priv, pub := testKeyPair(t)
	m, err := jwt.NewManager(priv, pub)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func routeAllTo(server *httptest.Server) *http.Client {
	target := server.Listener.Addr().String()
	transport := &http.Transport{}
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			r := req.Clone(req.Context())
			r.URL.Scheme = "http"
			r.URL.Host = target
			return transport.RoundTrip(r)
		}),
	}
}

func overrideOAuthHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	orig := oauthHTTPClient
	oauthHTTPClient = client
	t.Cleanup(func() { oauthHTTPClient = orig })
}

func overrideSlackUserInfoURL(t *testing.T, u string) {
	t.Helper()
	orig := slackUserInfoURL
	slackUserInfoURL = u
	t.Cleanup(func() { slackUserInfoURL = orig })
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type memoryOAuthRepo struct {
	identities map[string]*OAuthIdentity
	byUser     map[string][]*OAuthIdentity
	created    []*OAuthIdentity
	deleted    []string
	lastUsed   []string
}

func (f *memoryOAuthRepo) Create(ctx context.Context, i *OAuthIdentity) error {
	i.ID = fmt.Sprintf("ident_%d", len(f.created)+1)
	f.identities[i.Provider+"|"+i.ProviderUserID] = i
	f.byUser[i.UserID] = append(f.byUser[i.UserID], i)
	f.created = append(f.created, i)
	return nil
}

func (f *memoryOAuthRepo) GetByProvider(ctx context.Context, p, uid string) (*OAuthIdentity, error) {
	i, ok := f.identities[p+"|"+uid]
	if !ok {
		return nil, errors.New("not found")
	}
	return i, nil
}

func (f *memoryOAuthRepo) GetByUserID(ctx context.Context, id string) ([]*OAuthIdentity, error) {
	return f.byUser[id], nil
}

func (f *memoryOAuthRepo) Delete(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	for k, i := range f.identities {
		if i.ID == id {
			delete(f.identities, k)
		}
	}
	for uid, list := range f.byUser {
		var kept []*OAuthIdentity
		for _, i := range list {
			if i.ID != id {
				kept = append(kept, i)
			}
		}
		f.byUser[uid] = kept
	}
	return nil
}

func (f *memoryOAuthRepo) UpdateLastUsed(ctx context.Context, id string) error {
	f.lastUsed = append(f.lastUsed, id)
	return nil
}

func (f *memoryOAuthRepo) GetExpiringSoon(ctx context.Context, d time.Duration) ([]*OAuthIdentity, error) {
	return nil, nil
}

func (f *memoryOAuthRepo) UpdateTokens(ctx context.Context, id, a, r string, exp *time.Time) error {
	return nil
}

func newOAuthConfigs(tokenURL string) *OAuthProviderConfig {
	return &OAuthProviderConfig{
		Google: OAuthConfig{
			ClientID:     "google-client",
			ClientSecret: "google-secret",
			CallbackURL:  "https://app.example.com/callback/google",
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     tokenURL,
			UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scopes:       "openid email profile",
		},
		GitHub: OAuthConfig{
			ClientID:     "github-client",
			ClientSecret: "github-secret",
			CallbackURL:  "https://app.example.com/callback/github",
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     tokenURL,
			UserInfoURL:  "https://api.github.com/user",
			EmailsURL:    "https://api.github.com/user/emails",
			Scopes:       "read:user user:email",
		},
		Slack: OAuthConfig{
			ClientID:     "slack-client",
			ClientSecret: "slack-secret",
			CallbackURL:  "https://app.example.com/callback/slack",
			AuthURL:      "https://slack.com/openid/connect/authorize",
			TokenURL:     tokenURL,
			UserInfoURL:  "https://slack.com/api/openid.connect.userInfo",
			Scopes:       "openid email profile",
		},
		StateSigningKey: []byte("test-state-signing-key"),
	}
}

func newOAuthService(t *testing.T, cfg *OAuthProviderConfig) (*AuthService, *memoryOAuthRepo, *fakeUserRepo, *jwt.Manager) {
	t.Helper()
	users := &fakeUserRepo{users: map[string]*User{}}
	oauth := &memoryOAuthRepo{identities: map[string]*OAuthIdentity{}, byUser: map[string][]*OAuthIdentity{}}
	m := newTestManager(t)
	s := NewAuthService(users, &fakeTokenRepo{}, oauth, &fakeVerifyRepo{}, m, cfg)
	return s, oauth, users, m
}

func signOAuthState(t *testing.T, m *jwt.Manager, provider, codeChallenge string, linking bool) string {
	t.Helper()
	tok, err := m.SignClaims(&oauthStateClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   "state_test",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
		Provider:      provider,
		RedirectTo:    "/oauth/callback",
		Linking:       linking,
		CodeChallenge: codeChallenge,
	})
	if err != nil {
		t.Fatalf("signing oauth state: %v", err)
	}
	return tok
}

func newCallbackServer(t *testing.T, userInfo http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, tokenResponse{
			AccessToken:  "at_1",
			TokenType:    "Bearer",
			Scope:        "openid email",
			ExpiresIn:    3600,
			RefreshToken: "rt_1",
		})
	})
	mux.HandleFunc("/v1/userinfo", userInfo)
	server := httptest.NewServer(mux)
	overrideOAuthHTTPClient(t, routeAllTo(server))
	return server
}

func TestGetOAuthURLBuildsValidAuthURL(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))

	authURL, state, err := s.GetOAuthURL("google", "/dashboard", false, "")
	if err != nil {
		t.Fatalf("GetOAuthURL: %v", err)
	}
	if state == "" {
		t.Fatal("state JWT is empty")
	}

	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth URL %q: %v", authURL, err)
	}
	if u.Scheme != "https" || u.Host != "accounts.google.com" {
		t.Errorf("auth URL = %s://%s, want https://accounts.google.com", u.Scheme, u.Host)
	}
	q := u.Query()
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q, want code", q.Get("response_type"))
	}
	if q.Get("client_id") != "google-client" {
		t.Errorf("client_id = %q, want google-client", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "https://app.example.com/callback/google" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
	if q.Get("scope") != "openid email profile" {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	if q.Get("state") != state {
		t.Error("state query param does not match returned state JWT")
	}
	if q.Get("code_challenge") == "" {
		t.Error("code_challenge is empty")
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method = %q, want S256", q.Get("code_challenge_method"))
	}

	claims := &oauthStateClaims{}
	if err := m.ValidateClaims(state, claims); err != nil {
		t.Fatalf("validating state JWT: %v", err)
	}
	if claims.Provider != "google" {
		t.Errorf("claims.Provider = %q, want google", claims.Provider)
	}
	if claims.RedirectTo != "/dashboard" {
		t.Errorf("claims.RedirectTo = %q, want /dashboard", claims.RedirectTo)
	}
	if claims.Linking {
		t.Error("claims.Linking = true, want false")
	}
	if claims.CodeChallenge != q.Get("code_challenge") {
		t.Error("claims.CodeChallenge does not match URL code_challenge")
	}
}

func TestGetOAuthURLAllProvidersLinking(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	providers := map[string]string{
		"google": "google-client",
		"github": "github-client",
		"slack":  "slack-client",
	}

	for provider, clientID := range providers {
		authURL, state, err := s.GetOAuthURL(provider, "/settings", true, "")
		if err != nil {
			t.Fatalf("GetOAuthURL(%s): %v", provider, err)
		}
		q, err := url.Parse(authURL)
		if err != nil {
			t.Fatalf("parsing auth URL for %s: %v", provider, err)
		}
		if q.Query().Get("client_id") != clientID {
			t.Errorf("%s client_id = %q, want %q", provider, q.Query().Get("client_id"), clientID)
		}
		if q.Query().Get("response_type") != "code" {
			t.Errorf("%s response_type = %q", provider, q.Query().Get("response_type"))
		}
		claims := &oauthStateClaims{}
		if err := m.ValidateClaims(state, claims); err != nil {
			t.Fatalf("validating %s state JWT: %v", provider, err)
		}
		if claims.Provider != provider {
			t.Errorf("%s claims.Provider = %q", provider, claims.Provider)
		}
		if !claims.Linking {
			t.Errorf("%s claims.Linking = false, want true", provider)
		}
		if claims.RedirectTo != "/settings" {
			t.Errorf("%s claims.RedirectTo = %q", provider, claims.RedirectTo)
		}
	}
}

func TestGetOAuthURLUnsupportedProvider(t *testing.T) {
	t.Parallel()
	s, _, _, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))

	_, _, err := s.GetOAuthURL("twitter", "", false, "")
	if err == nil || !strings.Contains(err.Error(), "unsupported OAuth provider: twitter") {
		t.Errorf("err = %v, want unsupported provider error", err)
	}
}

func TestExchangeCodeSuccess(t *testing.T) {
	var gotGrant, gotClientID, gotCode, gotRedirect, gotVerifier string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("token request path = %q", r.URL.Path)
		}
		gotGrant = r.FormValue("grant_type")
		gotClientID = r.FormValue("client_id")
		gotCode = r.FormValue("code")
		gotRedirect = r.FormValue("redirect_uri")
		gotVerifier = r.FormValue("code_verifier")
		writeJSON(w, http.StatusOK, tokenResponse{
			AccessToken:  "at_1",
			TokenType:    "Bearer",
			Scope:        "openid email",
			ExpiresIn:    3600,
			RefreshToken: "rt_1",
			IDToken:      "idtok",
		})
	}))
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	cfg := newOAuthConfigs(server.URL + "/token").Google
	tokens, err := exchangeCode(cfg, "auth-code-1", "verifier-1")
	if err != nil {
		t.Fatalf("exchangeCode: %v", err)
	}
	if gotGrant != "authorization_code" {
		t.Errorf("grant_type = %q, want authorization_code", gotGrant)
	}
	if gotClientID != "google-client" {
		t.Errorf("client_id = %q, want google-client", gotClientID)
	}
	if gotCode != "auth-code-1" {
		t.Errorf("code = %q, want auth-code-1", gotCode)
	}
	if gotRedirect != "https://app.example.com/callback/google" {
		t.Errorf("redirect_uri = %q", gotRedirect)
	}
	if gotVerifier != "verifier-1" {
		t.Errorf("code_verifier = %q, want verifier-1", gotVerifier)
	}
	if tokens.AccessToken != "at_1" {
		t.Errorf("AccessToken = %q, want at_1", tokens.AccessToken)
	}
	if tokens.RefreshToken != "rt_1" {
		t.Errorf("RefreshToken = %q, want rt_1", tokens.RefreshToken)
	}
	if tokens.TokenType != "Bearer" || tokens.Scope != "openid email" || tokens.ExpiresIn != 3600 {
		t.Errorf("tokens = %+v", tokens)
	}
}

func TestExchangeCodeOmitsVerifierWhenEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("code_verifier") != "" {
			t.Errorf("code_verifier = %q, want empty", r.FormValue("code_verifier"))
		}
		writeJSON(w, http.StatusOK, tokenResponse{AccessToken: "at"})
	}))
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	cfg := newOAuthConfigs(server.URL + "/token").Google
	tokens, err := exchangeCode(cfg, "auth-code-1", "")
	if err != nil {
		t.Fatalf("exchangeCode: %v", err)
	}
	if tokens.AccessToken != "at" {
		t.Errorf("AccessToken = %q", tokens.AccessToken)
	}
}

func TestExchangeCodeNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("token expired"))
	}))
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	cfg := newOAuthConfigs(server.URL + "/token").Google
	_, err := exchangeCode(cfg, "auth-code-1", "")
	if err == nil || !strings.Contains(err.Error(), "token endpoint returned 502") {
		t.Errorf("err = %v, want 502 error", err)
	}
}

func TestExchangeCodeBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	cfg := newOAuthConfigs(server.URL + "/token").Google
	_, err := exchangeCode(cfg, "auth-code-1", "")
	if err == nil || !strings.Contains(err.Error(), "parsing token response") {
		t.Errorf("err = %v, want parsing error", err)
	}
}

func TestFetchGoogleUserInfo(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/userinfo" {
			t.Errorf("google userinfo path = %q", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "google-1",
			Email:         "g@example.com",
			EmailVerified: true,
			Name:          "Gopher",
			Picture:       "https://pic.example.com/g.png",
		})
	}))
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, name, pic, err := fetchUserInfo("google", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(google): %v", err)
	}
	if uid != "google-1" || email != "g@example.com" || !verified || name != "Gopher" || pic != "https://pic.example.com/g.png" {
		t.Errorf("got (%q, %q, %v, %q, %q)", uid, email, verified, name, pic)
	}
	if gotAuth != "Bearer tok-1" {
		t.Errorf("Authorization = %q, want Bearer tok-1", gotAuth)
	}
}

func TestFetchGitHubUserWithEmail(t *testing.T) {
	emailsCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, githubUserInfo{
			ID:        42,
			Login:     "gopher",
			Email:     "g@example.com",
			Name:      "Gopher",
			AvatarURL: "https://a.example.com/g.png",
		})
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		emailsCalled = true
		writeJSON(w, http.StatusOK, []githubEmail{})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, name, pic, err := fetchUserInfo("github", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(github): %v", err)
	}
	if uid != "42" || email != "g@example.com" || !verified || name != "Gopher" || pic != "https://a.example.com/g.png" {
		t.Errorf("got (%q, %q, %v, %q, %q)", uid, email, verified, name, pic)
	}
	if emailsCalled {
		t.Error("/user/emails called even though /user returned an email")
	}
}

func TestFetchGitHubUserWithoutEmailPrimaryVerifiedSelection(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, githubUserInfo{ID: 7, Login: "nobody", Email: "null"})
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []githubEmail{
			{Email: "b@example.com", Primary: false, Verified: true},
			{Email: "a@example.com", Primary: true, Verified: true},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, name, _, err := fetchUserInfo("github", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(github): %v", err)
	}
	if uid != "7" {
		t.Errorf("uid = %q, want 7", uid)
	}
	if email != "a@example.com" {
		t.Errorf("email = %q, want primary verified a@example.com", email)
	}
	if !verified {
		t.Error("verified = false, want true")
	}
	if name != "nobody" {
		t.Errorf("name = %q, want login fallback nobody", name)
	}
}

func TestFetchGitHubUserWithoutEmailFirstEmailFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, githubUserInfo{ID: 8, Login: "anon"})
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []githubEmail{
			{Email: "x@example.com", Primary: false, Verified: false},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, _, _, err := fetchUserInfo("github", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(github): %v", err)
	}
	if uid != "8" {
		t.Errorf("uid = %q, want 8", uid)
	}
	if email != "x@example.com" {
		t.Errorf("email = %q, want first email fallback", email)
	}
	if verified {
		t.Error("verified = true, want false")
	}
}

func TestFetchGitHubUserEmailsEndpointErrorFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, githubUserInfo{ID: 9, Login: "lonely"})
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("emails unavailable"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, _, _, err := fetchUserInfo("github", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(github) must fall back on error, got %v", err)
	}
	if uid != "9" {
		t.Errorf("uid = %q, want 9", uid)
	}
	if email != "" || verified {
		t.Errorf("got email %q verified %v, want empty email unverified", email, verified)
	}
}

func TestFetchSlackUserInfoOK(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slack" {
			t.Errorf("slack userinfo path = %q", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
			"user": slackUserInfo{
				Sub:               "U123",
				Email:             "s@example.com",
				EmailVerified:     true,
				Name:              "Slacker",
				Picture:           "https://p.example.com/s.png",
				PreferredUsername: "team",
			},
		})
	}))
	defer server.Close()
	overrideSlackUserInfoURL(t, server.URL+"/slack")
	overrideOAuthHTTPClient(t, routeAllTo(server))

	uid, email, verified, name, pic, err := fetchUserInfo("slack", "tok-1")
	if err != nil {
		t.Fatalf("fetchUserInfo(slack): %v", err)
	}
	if uid != "U123" || email != "s@example.com" || !verified || name != "Slacker" || pic != "https://p.example.com/s.png" {
		t.Errorf("got (%q, %q, %v, %q, %q)", uid, email, verified, name, pic)
	}
	if gotAuth != "Bearer tok-1" {
		t.Errorf("Authorization = %q, want Bearer tok-1", gotAuth)
	}
}

func TestFetchSlackUserInfoErrorEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": "invalid_token"})
	}))
	defer server.Close()
	overrideSlackUserInfoURL(t, server.URL+"/slack")
	overrideOAuthHTTPClient(t, routeAllTo(server))

	_, _, _, _, _, err := fetchUserInfo("slack", "tok-1")
	if err == nil || !strings.Contains(err.Error(), "slack api error: invalid_token") {
		t.Errorf("err = %v, want slack api error", err)
	}
}

func TestFetchUserInfoUnsupportedProvider(t *testing.T) {
	t.Parallel()
	_, _, _, _, _, err := fetchUserInfo("twitter", "tok-1")
	if err == nil || !strings.Contains(err.Error(), "unsupported provider: twitter") {
		t.Errorf("err = %v, want unsupported provider error", err)
	}
}

func TestHandleOAuthCallbackUnsupportedProvider(t *testing.T) {
	t.Parallel()
	s, _, _, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))

	_, _, _, isNew, linkApproval, err := s.HandleOAuthCallback(context.Background(), "twitter", "code", "state", "v")
	if err == nil || !strings.Contains(err.Error(), "unsupported OAuth provider: twitter") {
		t.Errorf("err = %v, want unsupported provider error", err)
	}
	if isNew || linkApproval {
		t.Error("isNew/linkApproval set on error")
	}
}

func TestHandleOAuthCallbackBadState(t *testing.T) {
	t.Parallel()
	s, _, _, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))

	_, _, _, _, _, err := s.HandleOAuthCallback(context.Background(), "google", "code", "garbage.state.jwt", "v")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Errorf("err = %v, want ErrOAuthStateMismatch", err)
	}
}

func TestHandleOAuthCallbackProviderMismatch(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	state := signOAuthState(t, m, "google", computeCodeChallenge("v"), false)

	_, _, _, _, _, err := s.HandleOAuthCallback(context.Background(), "github", "code", state, "v")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Errorf("err = %v, want ErrOAuthStateMismatch", err)
	}
}

func TestHandleOAuthCallbackCodeChallengeMismatch(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-A"), false)

	_, _, _, _, _, err := s.HandleOAuthCallback(context.Background(), "google", "code", state, "verifier-B")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Errorf("err = %v, want ErrOAuthStateMismatch", err)
	}
}

func TestHandleOAuthCallbackEmailNotVerified(t *testing.T) {
	server := newCallbackServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "google-1",
			Email:         "g@example.com",
			EmailVerified: false,
			Name:          "Gopher",
		})
	})
	defer server.Close()

	s, _, _, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))
	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), false)

	_, _, _, _, _, err := s.HandleOAuthCallback(context.Background(), "google", "code", state, "verifier-1")
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Errorf("err = %v, want ErrEmailNotVerified", err)
	}
}

func TestHandleOAuthCallbackExistingIdentity(t *testing.T) {
	server := newCallbackServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "google-1",
			Email:         "g@example.com",
			EmailVerified: true,
			Name:          "Gopher",
			Picture:       "https://pic.example.com/g.png",
		})
	})
	defer server.Close()

	s, oauth, users, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))
	users.users["g@example.com"] = &User{ID: "user_1", Email: "g@example.com", Role: "user", IsActive: true}
	oauth.identities["google|google-1"] = &OAuthIdentity{
		ID: "ident_1", UserID: "user_1", Provider: "google", ProviderUserID: "google-1",
	}

	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), false)
	user, access, refresh, isNew, linkApproval, err := s.HandleOAuthCallback(context.Background(), "google", "code", state, "verifier-1")
	if err != nil {
		t.Fatalf("HandleOAuthCallback: %v", err)
	}
	if user == nil || user.ID != "user_1" {
		t.Errorf("user = %+v, want user_1", user)
	}
	if access == "" || refresh == "" {
		t.Error("access/refresh tokens must be generated for existing identity")
	}
	if isNew {
		t.Error("isNew = true, want false")
	}
	if linkApproval {
		t.Error("linkApproval = true, want false")
	}
	if len(oauth.lastUsed) != 1 || oauth.lastUsed[0] != "ident_1" {
		t.Errorf("UpdateLastUsed calls = %v, want [ident_1]", oauth.lastUsed)
	}
	if len(oauth.created) != 0 {
		t.Error("no identity should be created for existing user")
	}
}

func TestHandleOAuthCallbackExistingUserWithPassword(t *testing.T) {
	server := newCallbackServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "google-1",
			Email:         "g@example.com",
			EmailVerified: true,
		})
	})
	defer server.Close()

	s, _, users, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))
	users.users["g@example.com"] = &User{
		ID: "user_1", Email: "g@example.com", PasswordHash: "bcrypt-hash", Role: "user", IsActive: true,
	}

	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), false)
	user, access, refresh, isNew, linkApproval, err := s.HandleOAuthCallback(context.Background(), "google", "code", state, "verifier-1")
	if err != nil {
		t.Fatalf("HandleOAuthCallback: %v", err)
	}
	if user == nil || user.ID != "user_1" {
		t.Errorf("user = %+v, want existing user_1", user)
	}
	if access != "" || refresh != "" {
		t.Error("tokens must be empty when linking approval is required")
	}
	if isNew {
		t.Error("isNew = true, want false")
	}
	if !linkApproval {
		t.Error("linkApproval = false, want true")
	}
}

func TestHandleOAuthCallbackNewUser(t *testing.T) {
	server := newCallbackServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "new-1",
			Email:         "new@example.com",
			EmailVerified: true,
			Name:          "Newbie",
			Picture:       "https://pic.example.com/new.png",
		})
	})
	defer server.Close()

	s, oauth, users, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))

	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), false)
	user, access, refresh, isNew, linkApproval, err := s.HandleOAuthCallback(context.Background(), "google", "code", state, "verifier-1")
	if err != nil {
		t.Fatalf("HandleOAuthCallback: %v", err)
	}
	if !isNew {
		t.Error("isNew = false, want true")
	}
	if linkApproval {
		t.Error("linkApproval = true, want false")
	}
	if access == "" || refresh == "" {
		t.Error("access/refresh tokens must be generated for new user")
	}
	if user == nil || user.Email != "new@example.com" || user.DisplayName != "Newbie" || user.AvatarURL != "https://pic.example.com/new.png" {
		t.Errorf("user = %+v", user)
	}
	if user.Role != "user" || !user.IsActive {
		t.Errorf("user role/active = %q/%v, want user/true", user.Role, user.IsActive)
	}
	if _, ok := users.users["new@example.com"]; !ok {
		t.Error("new user not stored in user repo")
	}
	if len(oauth.created) != 1 {
		t.Fatalf("created identities = %d, want 1", len(oauth.created))
	}
	identity := oauth.created[0]
	if identity.Provider != "google" || identity.ProviderUserID != "new-1" || identity.ProviderEmail != "new@example.com" {
		t.Errorf("identity = %+v", identity)
	}
	if identity.UserID != user.ID {
		t.Errorf("identity.UserID = %q, user.ID = %q", identity.UserID, user.ID)
	}
	if identity.AccessTokenEncrypted != "at_1" || identity.RefreshTokenEncrypted != "rt_1" {
		t.Errorf("identity tokens = %q/%q", identity.AccessTokenEncrypted, identity.RefreshTokenEncrypted)
	}
	if identity.ExpiresAt == nil {
		t.Error("identity.ExpiresAt is nil, want set from token expiry")
	}
}

func TestLinkOAuthIdentityNonLinkingState(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	state := signOAuthState(t, m, "google", computeCodeChallenge("v"), false)

	err := s.LinkOAuthIdentity(context.Background(), "user_1", "google", "code", state)
	if err == nil || !strings.Contains(err.Error(), "state was not created for linking") {
		t.Errorf("err = %v, want non-linking state error", err)
	}
}

func TestLinkOAuthIdentityUnsupportedProvider(t *testing.T) {
	t.Parallel()
	s, _, _, m := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	state := signOAuthState(t, m, "twitter", computeCodeChallenge("v"), true)

	err := s.LinkOAuthIdentity(context.Background(), "user_1", "twitter", "code", state)
	if err == nil || !strings.Contains(err.Error(), "unsupported OAuth provider: twitter") {
		t.Errorf("err = %v, want unsupported provider error", err)
	}
}

func TestLinkOAuthIdentityAlreadyLinked(t *testing.T) {
	server := newCallbackServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "42",
			Email:         "other@example.com",
			EmailVerified: true,
		})
	})
	defer server.Close()

	s, oauth, _, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))
	oauth.identities["google|42"] = &OAuthIdentity{
		ID: "ident_9", UserID: "other_user", Provider: "google", ProviderUserID: "42",
	}

	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), true)
	err := s.LinkOAuthIdentity(context.Background(), "user_1", "google", "code", state)
	if !errors.Is(err, ErrProviderAlreadyLinked) {
		t.Errorf("err = %v, want ErrProviderAlreadyLinked", err)
	}
	if len(oauth.created) != 0 {
		t.Error("identity created for already-linked provider")
	}
}

func TestLinkOAuthIdentitySuccess(t *testing.T) {
	var gotVerifier string
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		gotVerifier = r.FormValue("code_verifier")
		writeJSON(w, http.StatusOK, tokenResponse{AccessToken: "at_9", RefreshToken: "rt_9", ExpiresIn: 1200})
	})
	mux.HandleFunc("/v1/userinfo", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, googleUserInfo{
			Sub:           "99",
			Email:         "new@example.com",
			EmailVerified: true,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	overrideOAuthHTTPClient(t, routeAllTo(server))

	s, oauth, _, m := newOAuthService(t, newOAuthConfigs(server.URL+"/token"))
	state := signOAuthState(t, m, "google", computeCodeChallenge("verifier-1"), true)

	if err := s.LinkOAuthIdentity(context.Background(), "user_1", "google", "code", state); err != nil {
		t.Fatalf("LinkOAuthIdentity: %v", err)
	}
	if gotVerifier != "" {
		t.Errorf("code_verifier = %q, want empty for linking flow", gotVerifier)
	}
	if len(oauth.created) != 1 {
		t.Fatalf("created identities = %d, want 1", len(oauth.created))
	}
	identity := oauth.created[0]
	if identity.UserID != "user_1" || identity.Provider != "google" || identity.ProviderUserID != "99" || identity.ProviderEmail != "new@example.com" {
		t.Errorf("identity = %+v", identity)
	}
	if identity.AccessTokenEncrypted != "at_9" || identity.RefreshTokenEncrypted != "rt_9" {
		t.Errorf("identity tokens = %q/%q", identity.AccessTokenEncrypted, identity.RefreshTokenEncrypted)
	}
	if identity.ExpiresAt == nil {
		t.Error("identity.ExpiresAt is nil")
	}
}

func TestUnlinkOAuthIdentityLastAuthMethod(t *testing.T) {
	t.Parallel()
	s, oauth, users, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	users.users["u@example.com"] = &User{ID: "user_1", Email: "u@example.com", Role: "user"}
	oauth.byUser["user_1"] = []*OAuthIdentity{
		{ID: "ident_1", UserID: "user_1", Provider: "google"},
	}

	err := s.UnlinkOAuthIdentity(context.Background(), "user_1", "google")
	if !errors.Is(err, ErrLastAuthMethod) {
		t.Errorf("err = %v, want ErrLastAuthMethod", err)
	}
	if len(oauth.deleted) != 0 {
		t.Error("identity deleted despite last-auth-method guard")
	}
}

func TestUnlinkOAuthIdentitySuccessWithPassword(t *testing.T) {
	t.Parallel()
	s, oauth, users, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	users.users["u@example.com"] = &User{ID: "user_1", Email: "u@example.com", PasswordHash: "hash", Role: "user"}
	oauth.byUser["user_1"] = []*OAuthIdentity{
		{ID: "ident_1", UserID: "user_1", Provider: "google"},
	}

	if err := s.UnlinkOAuthIdentity(context.Background(), "user_1", "google"); err != nil {
		t.Fatalf("UnlinkOAuthIdentity: %v", err)
	}
	if len(oauth.deleted) != 1 || oauth.deleted[0] != "ident_1" {
		t.Errorf("deleted = %v, want [ident_1]", oauth.deleted)
	}
}

func TestUnlinkOAuthIdentitySuccessWithOtherProvider(t *testing.T) {
	t.Parallel()
	s, oauth, users, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	users.users["u@example.com"] = &User{ID: "user_1", Email: "u@example.com", Role: "user"}
	oauth.byUser["user_1"] = []*OAuthIdentity{
		{ID: "ident_1", UserID: "user_1", Provider: "google"},
		{ID: "ident_2", UserID: "user_1", Provider: "github"},
	}

	if err := s.UnlinkOAuthIdentity(context.Background(), "user_1", "google"); err != nil {
		t.Fatalf("UnlinkOAuthIdentity: %v", err)
	}
	if len(oauth.deleted) != 1 || oauth.deleted[0] != "ident_1" {
		t.Errorf("deleted = %v, want [ident_1]", oauth.deleted)
	}
}

func TestUnlinkOAuthIdentityProviderNotLinked(t *testing.T) {
	t.Parallel()
	s, oauth, users, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	users.users["u@example.com"] = &User{ID: "user_1", Email: "u@example.com", PasswordHash: "hash", Role: "user"}
	oauth.byUser["user_1"] = []*OAuthIdentity{
		{ID: "ident_1", UserID: "user_1", Provider: "github"},
	}

	err := s.UnlinkOAuthIdentity(context.Background(), "user_1", "google")
	if err == nil || !strings.Contains(err.Error(), "no linked identity for provider google") {
		t.Errorf("err = %v, want not-linked error", err)
	}
	if len(oauth.deleted) != 0 {
		t.Error("identity deleted for unlinked provider")
	}
}

func TestListLinkedIdentities(t *testing.T) {
	t.Parallel()
	s, oauth, _, _ := newOAuthService(t, newOAuthConfigs("https://token.example.com"))
	want := []*OAuthIdentity{
		{ID: "ident_1", UserID: "user_1", Provider: "google", ProviderUserID: "g1"},
		{ID: "ident_2", UserID: "user_1", Provider: "github", ProviderUserID: "gh2"},
	}
	oauth.byUser["user_1"] = want

	got, err := s.ListLinkedIdentities(context.Background(), "user_1")
	if err != nil {
		t.Fatalf("ListLinkedIdentities: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("identities = %d, want 2", len(got))
	}
	if got[0].ID != "ident_1" || got[0].Provider != "google" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].ID != "ident_2" || got[1].Provider != "github" {
		t.Errorf("got[1] = %+v", got[1])
	}

	none, err := s.ListLinkedIdentities(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("ListLinkedIdentities(ghost): %v", err)
	}
	if len(none) != 0 {
		t.Errorf("ghost identities = %d, want 0", len(none))
	}
}
