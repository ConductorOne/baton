package uhttp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/jwt"
)

type AuthCredentials interface {
	GetClient(ctx context.Context, options ...Option) (*http.Client, error)
}

type NoAuth struct{}

var _ AuthCredentials = (*NoAuth)(nil)

func (n *NoAuth) GetClient(ctx context.Context, options ...Option) (*http.Client, error) {
	return getHttpClient(ctx, options...)
}

type BearerAuth struct {
	Token string
}

var _ AuthCredentials = (*BearerAuth)(nil)

func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		Token: token,
	}
}

func (b *BearerAuth) GetClient(ctx context.Context, options ...Option) (*http.Client, error) {
	httpClient, err := getHttpClient(ctx, options...)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: b.Token},
	)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

type BasicAuth struct {
	Username string
	Password string
}

var _ AuthCredentials = (*BasicAuth)(nil)

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

func (b *BasicAuth) GetClient(ctx context.Context, options ...Option) (*http.Client, error) {
	httpClient, err := getHttpClient(ctx, options...)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	auth := b.Username + ":" + b.Password
	token := base64.StdEncoding.EncodeToString([]byte(auth))
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token, TokenType: "basic"},
	)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

type OAuth2ClientCredentials struct {
	cfg *clientcredentials.Config
}

var _ AuthCredentials = (*OAuth2ClientCredentials)(nil)

func NewOAuth2ClientCredentials(clientId, clientSecret string, tokenURL *url.URL, scopes []string) *OAuth2ClientCredentials {
	return &OAuth2ClientCredentials{
		cfg: &clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL.String(),
			Scopes:       scopes,
		},
	}
}

func (o *OAuth2ClientCredentials) GetClient(ctx context.Context, options ...Option) (*http.Client, error) {
	httpClient, err := getHttpClient(ctx, options...)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := o.cfg.TokenSource(ctx)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

type CreateJWTConfig func(credentials []byte, scopes ...string) (*jwt.Config, error)

type OAuth2JWT struct {
	Credentials     []byte
	Scopes          []string
	CreateJWTConfig CreateJWTConfig
}

var _ AuthCredentials = (*OAuth2JWT)(nil)

func NewOAuth2JWT(credentials []byte, scopes []string, createfn CreateJWTConfig) *OAuth2JWT {
	return &OAuth2JWT{
		Credentials:     credentials,
		Scopes:          scopes,
		CreateJWTConfig: createfn,
	}
}

func (o *OAuth2JWT) GetClient(ctx context.Context, options ...Option) (*http.Client, error) {
	httpClient, err := getHttpClient(ctx, options...)
	if err != nil {
		return nil, err
	}

	jwt, err := o.CreateJWTConfig(o.Credentials, o.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("creating JWT config failed: %w", err)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := jwt.TokenSource(ctx)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

func getHttpClient(ctx context.Context, options ...Option) (*http.Client, error) {
	options = append(options, WithLogger(true, nil))

	httpClient, err := NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client failed: %w", err)
	}

	return httpClient, nil
}
