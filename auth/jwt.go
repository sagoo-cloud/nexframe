// auth/jwt.go

package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type authKey struct{}

const (
	bearerWord       string = "Bearer"
	bearerFormat     string = "Bearer %s"
	authorizationKey string = "Authorization"
)

var (
	ErrMissingJwtToken        = errors.New("JWT token is missing")
	ErrMissingKeyFunc         = errors.New("keyFunc is missing")
	ErrTokenInvalid           = errors.New("Token is invalid")
	ErrTokenExpired           = errors.New("JWT token has expired")
	ErrTokenParseFail         = errors.New("Fail to parse JWT token ")
	ErrUnSupportSigningMethod = errors.New("Wrong signing method")
	ErrWrongContext           = errors.New("Wrong context for middleware")
	ErrNeedTokenProvider      = errors.New("Token provider is missing")
	ErrSignToken              = errors.New("Can not sign token.Is the key correct?")
	ErrGetKey                 = errors.New("Can not get key while signing token")
)

var (
	issuer                    = "qmPlus"
	bufferTime  time.Duration = -time.Second
	expiresTime time.Duration = 7 * 24 * time.Hour
)

// JwtConfig 定义了 JWT 中间件的配置
type JwtConfig struct {
	SigningKey    interface{} // 用于签名的密钥
	TokenLookup   string      // 定义如何查找令牌
	SigningMethod string      // 签名方法
	BufferTime    string      // 生效时间
	ExpiresTime   string      // 过期时间
	Issuer        string      // 签发者
	ErrHandler    func(w http.ResponseWriter, r *http.Request, err error)
}

// jwtMiddleware 是验证 JWT 令牌的中间件
type jwtMiddleware struct {
	opt  options
	conf JwtConfig
}

// New 创建一个新的 jwtMiddleware 实例
func New(config JwtConfig) (*jwtMiddleware, error) {
	o := options{signingMethod: jwt.SigningMethodHS256}
	if config.SigningKey == nil {
		return nil, errors.New("jwt 中间件需要签名密钥")
	} else {
		o.keyFunc = func(token *jwt.Token) (interface{}, error) {
			return config.SigningKey, nil
		}
	}
	if config.SigningMethod == "" {
		config.SigningMethod = "HS256"
	}
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}
	if config.ErrHandler == nil {
		config.ErrHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
	}
	var err error
	if config.Issuer != "" {
		issuer = config.Issuer
	}
	if config.BufferTime != "" {
		bufferTime, err = time.ParseDuration(config.BufferTime)
		if err != nil {
			return nil, fmt.Errorf("Error parsing bufferTime: %v", err)
		}
	}
	if config.ExpiresTime != "" {
		expiresTime, err = time.ParseDuration(config.ExpiresTime)
		if err != nil {
			return nil, fmt.Errorf("Error parsing expiresTime: %v", err)
		}
	}
	return &jwtMiddleware{
		conf: config,
		opt:  o,
	}, nil
}

// Middleware 返回一个可以与 gorilla/mux 的 Use() 方法一起使用的 http.Handler
func (jm *jwtMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jm.extractToken(r)
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}

		claims, err := ParseJwtToken(token, WithSigningMethod(jm.opt.signingMethod), WithKeyFunc(jm.opt.keyFunc))
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), claims)))
	})
}

// extractToken 从请求中提取令牌
func (jm *jwtMiddleware) extractToken(r *http.Request) (string, error) {
	parts := strings.Split(jm.conf.TokenLookup, ":")
	switch parts[0] {
	case "header":
		return extractTokenFromHeader(r, parts[1])
	case "query":
		return extractTokenFromQuery(r, parts[1])
	case "cookie":
		return extractTokenFromCookie(r, parts[1])
	}
	return "", errors.New("无效的令牌查找配置")
}

// extractTokenFromHeader 从请求头中提取令牌
func extractTokenFromHeader(r *http.Request, header string) (string, error) {
	authHeader := r.Header.Get(header)
	if authHeader == "" {
		return "", errors.New("缺少认证头")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("无效的认证头")
	}
	return parts[1], nil
}

// extractTokenFromQuery 从查询参数中提取令牌
func extractTokenFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", errors.New("查询参数中缺少令牌")
	}
	return token, nil
}

// extractTokenFromCookie 从 cookie 中提取令牌
func extractTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", errors.New("cookie 中缺少令牌")
	}
	return cookie.Value, nil
}

// Parser is a jwt parser
type options struct {
	signingMethod jwt.SigningMethod
	claims        func() jwt.Claims
	keyFunc       jwt.Keyfunc
	tokenHeader   map[string]interface{}
}

type TokenClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	// 在这里添加你想包含的其他声明
	jwt.RegisteredClaims
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
			claims.Issuer = issuer
		}
		if claims.NotBefore == nil {
			claims.NotBefore = jwt.NewNumericDate(time.Now().Add(bufferTime))
		}
		if claims.ExpiresAt == nil {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresTime))
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
