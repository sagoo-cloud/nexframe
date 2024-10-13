package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sagoo-cloud/nexframe/configs"
	"time"
)

// Parser is a jwt parser
type options struct {
	signingMethod jwt.SigningMethod
	claims        func() jwt.Claims
	keyFunc       jwt.Keyfunc
	tokenHeader   map[string]interface{}
}

// Option is jwt option.
type Option func(*options)

func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}
func WithKeyFunc(f jwt.Keyfunc) Option {
	return func(o *options) {
		o.keyFunc = f
	}
}
func WithClaims(f func() jwt.Claims) Option {
	return func(o *options) {
		o.claims = f
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

func GenerateToken(key string, opts ...Option) (string, error) {
	config := configs.LoadTokenConfig()
	o := &options{
		claims: func() jwt.Claims {
			return &TokenClaims{}
		},
		signingMethod: jwt.SigningMethodHS256,
	}
	for _, opt := range opts {
		opt(o)
	}

	if claims, ok := o.claims().(*TokenClaims); ok {
		if claims.Issuer == "" {
			claims.Issuer = config.Issuer
		}
		if claims.NotBefore == nil {
			claims.NotBefore = jwt.NewNumericDate(time.Now().Add(config.BufferTime))
		}
		if claims.ExpiresAt == nil {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(config.ExpiresTime))
		}
		return jwt.NewWithClaims(o.signingMethod, claims).SignedString([]byte(key))
	}
	return jwt.NewWithClaims(o.signingMethod, o.claims()).SignedString([]byte(key))
}

// ParseJwtToken 解析和验证令牌
func ParseJwtToken(jwtToken string, opts ...Option) (jwt.Claims, error) {
	var (
		err       error
		tokenInfo *jwt.Token
		o         = &options{
			claims: func() jwt.Claims {
				return &TokenClaims{}
			},
			signingMethod: jwt.SigningMethodHS256,
		}
	)

	for _, opt := range opts {
		opt(o)
	}
	if o.claims != nil {
		tokenInfo, err = jwt.ParseWithClaims(jwtToken, o.claims(), o.keyFunc)
	} else {
		tokenInfo, err = jwt.Parse(jwtToken, o.keyFunc)
	}
	if err != nil {
		return nil, err
	} else if !tokenInfo.Valid {
		return nil, ErrTokenInvalid
	} else if tokenInfo.Method != o.signingMethod {
		return nil, ErrUnSupportSigningMethod
	}
	return tokenInfo.Claims, nil
}
