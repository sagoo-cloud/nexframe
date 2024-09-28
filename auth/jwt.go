package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware 接口定义
type JWTMiddleware interface {
	Middleware(next http.Handler) http.Handler
	GenerateTokenPair(username string) (*TokenPair, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

// 自定义错误
var (
	ErrMissingJwtToken        = errors.New("缺少JWT令牌")
	ErrTokenInvalid           = errors.New("令牌无效")
	ErrTokenExpired           = errors.New("JWT令牌已过期")
	ErrTokenParseFail         = errors.New("解析JWT令牌失败")
	ErrUnSupportSigningMethod = errors.New("不支持的签名方法")
	ErrInvalidTokenType       = errors.New("非访问令牌")
)

// 在 jwtMiddleware 结构体外部定义 TokenClaims 的对象池
var tokenClaimsPool = sync.Pool{
	New: func() interface{} {
		return &TokenClaims{}
	},
}

// TokenPair 结构体用于存储访问令牌和刷新令牌
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type authKey struct{}

// AuthClaims 定义了认证声明的接口
type AuthClaims interface {
	jwt.Claims
	GetUsername() string
}

// TokenClaims 实现了 AuthClaims 接口
type TokenClaims struct {
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

func (tc *TokenClaims) GetUsername() string {
	return tc.Username
}

// JwtConfig 定义了JWT中间件的配置
type JwtConfig struct {
	configs.TokenConfig
	SigningMethod jwt.SigningMethod
	ErrHandler    func(w http.ResponseWriter, r *http.Request, err error)
}

// jwtMiddleware 是用于验证JWT令牌的中间件
type jwtMiddleware struct {
	conf         JwtConfig
	keyFunc      jwt.Keyfunc
	extractToken func(*http.Request) (string, error)
}

// NewJwt 创建一个新的jwtMiddleware实例
func NewJwt() (*jwtMiddleware, error) {
	cfg := configs.LoadTokenConfig()
	if cfg == nil {
		return nil, fmt.Errorf("无法加载令牌配置")
	}

	// 确保 SigningKey 是 []byte 类型
	var signingKey []byte
	switch k := cfg.SigningKey.(type) {
	case []byte:
		signingKey = k
	case string:
		signingKey = []byte(k)
	default:
		return nil, fmt.Errorf("无效的签名密钥类型")
	}

	config := JwtConfig{
		TokenConfig:   *cfg,
		SigningMethod: GetSigningMethod(cfg.Method),
		ErrHandler:    defaultErrorHandler,
	}
	config.SigningKey = signingKey

	jm := &jwtMiddleware{
		conf: config,
		keyFunc: func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		},
	}

	if err := jm.initializeTokenExtractor(); err != nil {
		return nil, err
	}

	return jm, nil
}

// Middleware 返回一个http.Handler，用于验证JWT令牌
func (jm *jwtMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jm.extractToken(r)
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}

		claims, err := jm.parseJwtToken(token)
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}

		ctx := NewAuthContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (jm *jwtMiddleware) initializeTokenExtractor() error {
	extractors := map[string]func(*http.Request, string) (string, error){
		"header": extractTokenFromHeader,
		"query":  extractTokenFromQuery,
		"cookie": extractTokenFromCookie,
	}

	parts := strings.Split(jm.conf.TokenLookup, ":")
	if len(parts) != 2 {
		return fmt.Errorf("无效的令牌查找配置: %s", jm.conf.TokenLookup)
	}

	extractor, ok := extractors[parts[0]]
	if !ok {
		return fmt.Errorf("不支持的令牌查找方法: %s", parts[0])
	}

	jm.extractToken = func(r *http.Request) (string, error) {
		return extractor(r, parts[1])
	}

	return nil
}

func extractTokenFromHeader(r *http.Request, header string) (string, error) {
	authHeader := r.Header.Get(header)
	if authHeader == "" {
		return "", ErrMissingJwtToken
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrTokenInvalid
	}
	return parts[1], nil
}

func extractTokenFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", ErrMissingJwtToken
	}
	return token, nil
}

func extractTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", ErrMissingJwtToken
	}
	return cookie.Value, nil
}

func (jm *jwtMiddleware) parseJwtToken(tokenString string) (AuthClaims, error) {
	claims := tokenClaimsPool.Get().(*TokenClaims)
	defer tokenClaimsPool.Put(claims)

	*claims = TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jm.conf.SigningMethod.Alg() {
			return nil, ErrUnSupportSigningMethod
		}
		return jm.conf.SigningKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenParseFail, err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// 移除这个检查，因为我们现在处理两种类型的令牌
	// if claims.TokenType != "access" {
	//     return nil, ErrInvalidTokenType
	// }

	returnClaims := &TokenClaims{
		Username:         claims.Username,
		TokenType:        claims.TokenType,
		RegisteredClaims: claims.RegisteredClaims,
	}

	return returnClaims, nil
}

// GenerateTokenPair 生成新的JWT令牌
func (jm *jwtMiddleware) GenerateTokenPair(username string) (*TokenPair, error) {
	accessToken, err := jm.createToken(username, "access", jm.conf.ExpiresTime)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jm.createToken(username, "refresh", jm.conf.RefreshExpiresTime)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (jm *jwtMiddleware) createToken(username, tokenType string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := &TokenClaims{
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jm.conf.Issuer,
		},
	}

	token := jwt.NewWithClaims(jm.conf.SigningMethod, claims)
	return token.SignedString(jm.conf.SigningKey)
}

// RefreshToken 方法用于刷新访问令牌
func (jm *jwtMiddleware) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := jm.parseJwtToken(refreshToken)
	if err != nil {
		return nil, err
	}

	tokenClaims, ok := claims.(*TokenClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	if tokenClaims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	// 检查刷新令牌是否过期
	if tokenClaims.ExpiresAt != nil && tokenClaims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return jm.GenerateTokenPair(tokenClaims.Username)
}

// NewAuthContext 将认证声明添加到上下文中
func NewAuthContext(ctx context.Context, claims AuthClaims) context.Context {
	return context.WithValue(ctx, authKey{}, claims)
}

// ClaimsFromContext 从上下文中获取认证声明
func ClaimsFromContext(ctx context.Context) (AuthClaims, bool) {
	claims, ok := ctx.Value(authKey{}).(AuthClaims)
	return claims, ok
}

// GetCurrentUser 是一个高级辅助函数，用于获取当前用户名
func GetCurrentUser(ctx context.Context) (string, error) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", errors.New("未找到认证信息")
	}
	return claims.GetUsername(), nil
}

// defaultErrorHandler 是默认的错误处理函数
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// GetSigningMethod 根据提供的方法名返回对应的 JWT 签名方法。
// 支持的方法包括 "HS256"、"HS384" 和 "HS512"。
// 如果提供了不支持的方法，默认返回 HS256。
func GetSigningMethod(Method string) jwt.SigningMethod {
	switch Method {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		return jwt.SigningMethodHS256
	}
}
