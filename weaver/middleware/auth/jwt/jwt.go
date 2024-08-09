package jwt

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
)

type authKey struct{}

const (
	bearerWord       string = "Bearer"
	bearerFormat     string = "Bearer %s"
	authorizationKey string = "Authorization"
)

type UserClaims struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	more     map[string]string
	jwt.RegisteredClaims
}

func (u UserClaims) GetToken() string {
	return u.more[authorizationKey]
}

func (u *UserClaims) set(key, value string) {
	u.more[key] = value
}

func GenerateToken(key string, opts ...Option) (string, error) {
	o := &options{signingMethod: jwt.SigningMethodHS256}
	for _, opt := range opts {
		opt(o)
	}
	return jwt.NewWithClaims(o.signingMethod, o.claims()).SignedString([]byte(key))
}

// Parser is a jwt parser
type options struct {
	signingMethod jwt.SigningMethod
	claims        func() jwt.Claims
	tokenHeader   map[string]interface{}
	filterPath    []string
}

// Option is jwt option.
type Option func(*options)

func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

func WithClaims(f func() jwt.Claims) Option {
	return func(o *options) {
		o.claims = f
	}
}

func FilterPath(paths []string) Option {
	return func(o *options) {
		o.filterPath = paths
	}
}

// NewContext put auth info into context
func NewContext(ctx context.Context, info jwt.Claims) context.Context {
	return context.WithValue(ctx, authKey{}, info)
}

// FromContext extract auth info from context
func FromContext(ctx context.Context) (token jwt.Claims, ok bool) {
	token, ok = ctx.Value(authKey{}).(jwt.Claims)
	return
}
