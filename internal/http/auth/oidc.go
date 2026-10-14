package httpauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/gi8lino/motus/internal/http/response"
)

const (
	oidcStateCookie = "motus_oidc_state"
	oidcStateTTL    = 10 * time.Minute
)

// OIDCConfig configures the authorization-code flow for one trusted provider.
type OIDCConfig struct {
	// Issuer is the trusted OpenID Connect issuer URL.
	Issuer string
	// ClientID identifies Motus to the provider.
	ClientID string
	// ClientSecret authenticates Motus to the provider.
	ClientSecret string
	// SessionSecret signs transient browser login state.
	SessionSecret string
	// PublicURL is the externally visible Motus base URL.
	PublicURL string
}

// OIDC implements a browser authorization-code flow with PKCE and nonce binding.
type OIDC struct {
	// loginService resolves verified OIDC identities through application policy.
	loginService oidcLoginService
	local        *Local
	// verifier validates provider identity tokens.
	verifier *oidc.IDTokenVerifier
	// oauth contains provider endpoints and client credentials.
	oauth *oauth2.Config
	// issuer is the normalized trusted issuer.
	issuer string
	// secret signs transient login state.
	secret []byte
}

// oidcLoginState is signed into the transient browser state cookie.
type oidcLoginState struct {
	// State binds the callback to the initiating browser request.
	State string `json:"s"`
	// Verifier is the PKCE verifier exchanged with the authorization code.
	Verifier string `json:"v"`
	// Next is the local path restored after successful login.
	Next string `json:"n"`
	// Expires is the Unix timestamp after which the state is rejected.
	Expires int64 `json:"e"`
}

// oidcClaims contains the provider claims Motus uses for account creation.
type oidcClaims struct {
	PreferredUsername string `json:"preferred_username"`
	// Email identifies the external user.
	Email string `json:"email"`
	// Name is the optional provider display name.
	Name string `json:"name"`
}

// NewOIDC discovers the provider and constructs the verified login flow.
func NewOIDC(ctx context.Context, config OIDCConfig, loginService oidcLoginService, local *Local) (*OIDC, error) {
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, err
	}
	return &OIDC{
		loginService: loginService,
		local:        local,
		verifier:     provider.Verifier(&oidc.Config{ClientID: config.ClientID}),
		issuer:       strings.TrimSpace(config.Issuer),
		secret:       []byte(config.SessionSecret),
		oauth: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  strings.TrimRight(config.PublicURL, "/") + "/auth/oidc/callback",
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}, nil
}

// Login redirects the browser to the provider after storing signed transient state.
func (o *OIDC) login(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken(24)
	if err != nil {
		response.InternalServerError(nil, w, err)
		return
	}
	verifier := oauth2.GenerateVerifier()
	next := r.URL.Query().Get("next")
	if !response.IsLocalPath(next) {
		next = "/"
	}
	if err := o.setState(w, oidcLoginState{State: state, Verifier: verifier, Next: next, Expires: time.Now().Add(oidcStateTTL).Unix()}); err != nil {
		response.InternalServerError(nil, w, err)
		return
	}
	http.Redirect(w, r, o.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oidc.Nonce(state)), http.StatusFound)
}

// Callback validates the authorization response, persists the identity mapping, and creates a Motus session.
func (o *OIDC) callback(w http.ResponseWriter, r *http.Request) {
	state, err := o.readState(r)
	o.clearState(w)
	if err != nil || state.State != r.URL.Query().Get("state") || state.Expires <= time.Now().Unix() {
		response.Problem(w, http.StatusBadRequest, "Invalid login state.")
		return
	}
	token, err := o.oauth.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(state.Verifier))
	if err != nil {
		response.Problem(w, http.StatusUnauthorized, "Login failed.")
		return
	}
	raw, _ := token.Extra("id_token").(string)
	idToken, err := o.verifier.Verify(r.Context(), raw)
	if err != nil || idToken.Issuer != o.issuer || idToken.Nonce != state.State {
		response.Problem(w, http.StatusUnauthorized, "Invalid identity.")
		return
	}
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		response.Problem(w, http.StatusUnauthorized, "Invalid identity claims.")
		return
	}
	user, err := o.loginService.Login(r.Context(), idToken.Issuer, idToken.Subject, claims.PreferredUsername, claims.Email, claims.Name)
	if err != nil {
		response.Problem(w, http.StatusForbidden, "Login is not permitted.")
		return
	}
	if user == nil {
		response.Problem(w, http.StatusForbidden, "Login is not permitted.")
		return
	}
	session, err := o.local.StartSession(r.Context(), *user)
	if err != nil {
		response.InternalServerError(nil, w, err)
		return
	}
	o.local.WriteSessionCookie(w, r, session)
	http.Redirect(w, r, state.Next, http.StatusFound)
}

// randomToken returns a URL-safe cryptographically random token.
func randomToken(bytes int) (string, error) {
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// setState signs and stores transient OIDC login state.
func (o *OIDC) setState(w http.ResponseWriter, state oidcLoginState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, o.secret)
	_, _ = mac.Write([]byte(encoded))
	value := encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   int(oidcStateTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(o.oauth.RedirectURL, "https://"),
	})
	return nil
}

// readState verifies and decodes transient OIDC login state.
func (o *OIDC) readState(r *http.Request) (oidcLoginState, error) {
	cookie, err := r.Cookie(oidcStateCookie)
	if err != nil {
		return oidcLoginState{}, err
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return oidcLoginState{}, errors.New("invalid state")
	}
	mac := hmac.New(sha256.New, o.secret)
	_, _ = mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return oidcLoginState{}, errors.New("invalid state signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return oidcLoginState{}, err
	}
	var state oidcLoginState
	err = json.Unmarshal(payload, &state)
	return state, err
}

// clearState expires the transient OIDC login-state cookie.
func (o *OIDC) clearState(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: oidcStateCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}
