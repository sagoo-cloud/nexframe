package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sagoo-cloud/nexframe/configs"
)

// JWTMiddleware 定义JWT中间件接口
type JWTMiddleware interface {
	Middleware(next http.Handler) http.Handler
	GenerateTokenPair(username string) (*TokenPair, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrMissingJwtToken        = errors.New("缺少JWT令牌")
	ErrTokenInvalid           = errors.New("令牌无效")
	ErrTokenExpired           = errors.New("JWT令牌已过期")
	ErrTokenParseFail         = errors.New("解析JWT令牌失败")
	ErrUnSupportSigningMethod = errors.New("不支持的签名方法")
	ErrInvalidTokenType       = errors.New("非访问令牌")
	ErrNilConfig              = errors.New("配置为空")
)

// TokenClaimsPool 定义TokenClaims对象池,用于复用TokenClaims对象,减少内存分配和GC压力
var tokenClaimsPool = sync.Pool{
	New: func() interface{} {
		return &TokenClaims{}
	},
}

// TokenPair 定义访问令牌和刷新令牌的结构体
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type authKey struct{}

// AuthClaims 定义JWT Claims接口
type AuthClaims interface {
	jwt.Claims
	GetUsername() string
}

// TokenClaims 实现AuthClaims接口
type TokenClaims struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	Data      interface{}
	jwt.RegisteredClaims
}

// GetUsername 获取Claims中的用户名
func (tc *TokenClaims) GetUsername() string {
	return tc.Username
}

// JwtConfig 定义JWT配置结构体
type JwtConfig struct {
	configs.TokenConfig
	SigningMethod jwt.SigningMethod
	ErrHandler    func(w http.ResponseWriter, r *http.Request, err error)
}

type jwtMiddleware struct {
	conf         *JwtConfig
	keyFunc      jwt.Keyfunc
	extractToken func(*http.Request) (string, error)
}

// NewJwt 创建JWT中间件实例
func NewJwt() (*jwtMiddleware, error) {
	cfg := configs.LoadTokenConfig()
	if cfg == nil {
		return nil, ErrNilConfig
	}

	cfg.ExcludePaths = append(cfg.ExcludePaths, "/swagger/index.html", "/swagger/*")

	signingKey, err := parseSigningKey(cfg.SigningKey)
	if err != nil {
		return nil, err
	}

	config := JwtConfig{
		TokenConfig:   *cfg,
		SigningMethod: GetSigningMethod(cfg.Method),
		ErrHandler:    defaultErrorHandler,
	}
	config.SigningKey = signingKey

	jm := &jwtMiddleware{
		conf: &config,
		keyFunc: func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		},
	}

	if err := jm.initializeTokenExtractor(); err != nil {
		return nil, err
	}

	// 检查jm.conf是否为nil,避免空指针panic
	if jm.conf == nil {
		return nil, ErrNilConfig
	}

	return jm, nil
}

// parseSigningKey 解析签名密钥
func parseSigningKey(key interface{}) ([]byte, error) {
	switch k := key.(type) {
	case []byte:
		return k, nil
	case string:
		return []byte(k), nil
	default:
		return nil, errors.New("无效的签名密钥类型")
	}
}

// Middleware JWT中间件的核心处理逻辑
func (jm *jwtMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 使用defer语句捕获可能的panic,避免程序崩溃
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, fmt.Sprintf("Panic: %v", err), http.StatusInternalServerError)
			}
		}()

		// 检查请求路径是否在排除路径列表中
		if jm.isExcludedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 从请求中提取JWT令牌
		token, err := jm.extractToken(r)
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}

		// 解析JWT令牌
		claims, err := jm.parseJwtToken(token)
		if err != nil {
			jm.conf.ErrHandler(w, r, err)
			return
		}

		// 将解析后的Claims添加到请求的上下文中
		ctx := NewAuthContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// initializeTokenExtractor 初始化令牌提取器
func (jm *jwtMiddleware) initializeTokenExtractor() error {
	// 定义不同位置的令牌提取函数
	extractors := map[string]func(*http.Request, string) (string, error){
		"header": extractTokenFromHeader,
		"query":  extractTokenFromQuery,
		"cookie": extractTokenFromCookie,
	}

	parts := strings.SplitN(jm.conf.TokenLookup, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("无效的令牌查找配置: %s", jm.conf.TokenLookup)
	}

	// 根据配置选择相应的令牌提取函数
	extractor, ok := extractors[parts[0]]
	if !ok {
		return fmt.Errorf("不支持的令牌查找方法: %s", parts[0])
	}

	// 设置令牌提取函数
	jm.extractToken = func(r *http.Request) (string, error) {
		return extractor(r, parts[1])
	}

	return nil
}

// extractTokenFromHeader 从请求头中提取JWT令牌
func extractTokenFromHeader(r *http.Request, header string) (string, error) {
	authHeader := r.Header.Get(header)
	if authHeader == "" {
		return "", ErrMissingJwtToken
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrTokenInvalid
	}
	return parts[1], nil
}

// extractTokenFromQuery 从查询参数中提取JWT令牌
func extractTokenFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", ErrMissingJwtToken
	}
	return token, nil
}

// extractTokenFromCookie 从Cookie中提取JWT令牌
func extractTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", ErrMissingJwtToken
	}
	return cookie.Value, nil
}

// parseJwtToken 解析JWT令牌,并将结果放入对象池中复用
func (jm *jwtMiddleware) parseJwtToken(tokenString string) (AuthClaims, error) {
	// 从对象池中获取TokenClaims对象
	claims := tokenClaimsPool.Get().(*TokenClaims)
	defer tokenClaimsPool.Put(claims)

	*claims = TokenClaims{} // 重置claims

	// 使用jwt库解析令牌字符串
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法是否匹配
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

	// 创建一个新的TokenClaims对象,避免在返回时修改对象池中的对象
	returnClaims := &TokenClaims{
		Username:         claims.Username,
		TokenType:        claims.TokenType,
		RegisteredClaims: claims.RegisteredClaims,
		Data:             claims.Data,
	}

	return returnClaims, nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌
func (jm *jwtMiddleware) GenerateTokenPair(username string) (*TokenPair, error) {
	// 生成访问令牌
	accessToken, err := jm.createToken(username, TokenTypeAccess, jm.conf.ExpiresTime)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := jm.createToken(username, TokenTypeRefresh, jm.conf.RefreshExpiresTime)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// createToken 创建JWT令牌,使用对象池优化内存分配
func (jm *jwtMiddleware) createToken(username, tokenType string, expiration time.Duration) (string, error) {
	now := time.Now()

	// 从对象池中获取TokenClaims对象
	claims := tokenClaimsPool.Get().(*TokenClaims)
	defer tokenClaimsPool.Put(claims)

	*claims = TokenClaims{
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jm.conf.Issuer,
		},
	}

	// 使用配置的签名方法创建令牌
	token := jwt.NewWithClaims(jm.conf.SigningMethod, claims)
	return token.SignedString(jm.conf.SigningKey)
}

// RefreshToken 刷新访问令牌
func (jm *jwtMiddleware) RefreshToken(refreshToken string) (*TokenPair, error) {
	// 解析刷新令牌
	claims, err := jm.parseJwtToken(refreshToken)
	if err != nil {
		return nil, err
	}

	tokenClaims, ok := claims.(*TokenClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	// 检查令牌类型是否为刷新令牌
	if tokenClaims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	// 检查刷新令牌是否已过期
	if tokenClaims.ExpiresAt != nil && tokenClaims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// 生成新的访问令牌和刷新令牌
	return jm.GenerateTokenPair(tokenClaims.Username)
}

// isExcludedPath 检查请求路径是否在排除路径列表中
func (jm *jwtMiddleware) isExcludedPath(reqPath string) bool {
	for _, excludePath := range jm.conf.ExcludePaths {
		if strings.HasSuffix(excludePath, "*") {
			prefix := strings.TrimSuffix(excludePath, "*")
			if strings.HasPrefix(reqPath, prefix) {
				return true
			}
		} else if matched, _ := path.Match(excludePath, reqPath); matched {
			return true
		}
	}
	return false
}

// NewAuthContext 将认证信息添加到请求的上下文中
func NewAuthContext(ctx context.Context, claims AuthClaims) context.Context {
	return context.WithValue(ctx, authKey{}, claims)
}

// ClaimsFromContext 从请求的上下文中获取认证信息
func ClaimsFromContext(ctx context.Context) (AuthClaims, bool) {
	claims, ok := ctx.Value(authKey{}).(AuthClaims)
	return claims, ok
}

// GetCurrentUser 获取当前请求的用户名
func GetCurrentUser(ctx context.Context) (string, error) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", errors.New("未找到认证信息")
	}
	return claims.GetUsername(), nil
}

// defaultErrorHandler 默认的错误处理函数
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// GetSigningMethod 根据签名方法名获取对应的SigningMethod
func GetSigningMethod(method string) jwt.SigningMethod {
	switch method {
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
