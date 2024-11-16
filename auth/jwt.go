package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sagoo-cloud/nexframe/configs"
)

// JWTMiddleware 接口定义
type JWTMiddleware interface {
	Middleware(next http.Handler) http.Handler
	GenerateTokenPair(username string) (*TokenPair, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// 自定义错误
var (
	ErrMissingJwtToken        = errors.New("缺少JWT令牌")
	ErrTokenInvalid           = errors.New("令牌无效")
	ErrTokenExpired           = errors.New("JWT令牌已过期")
	ErrTokenParseFail         = errors.New("解析JWT令牌失败")
	ErrUnSupportSigningMethod = errors.New("不支持的签名方法")
	ErrInvalidTokenType       = errors.New("非访问令牌")
	ErrNilConfig              = errors.New("配置为空")
)

// TokenClaims 的对象池
var tokenClaimsPool = sync.Pool{
	New: func() interface{} {
		return &TokenClaims{}
	},
}

// TokenPair 存储访问令牌和刷新令牌
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type authKey struct{}

// AuthClaims 定义认证声明的接口
type AuthClaims interface {
	jwt.Claims
	GetUsername() string
}

// TokenClaims 实现 AuthClaims 接口
type TokenClaims struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
	Data      interface{}
	jwt.RegisteredClaims
}

// GetUsername 返回用户名
func (tc *TokenClaims) GetUsername() string {
	return tc.Username
}

// JwtConfig 定义JWT中间件的配置
type JwtConfig struct {
	configs.TokenConfig
	SigningMethod jwt.SigningMethod
	ErrHandler    func(w http.ResponseWriter, r *http.Request, err error)
}

// jwtMiddleware 实现JWT令牌验证的中间件
type jwtMiddleware struct {
	conf         JwtConfig
	keyFunc      jwt.Keyfunc
	extractToken func(*http.Request) (string, error)
}

// NewJwt 创建新的jwtMiddleware实例
func NewJwt() (*jwtMiddleware, error) {
	cfg := configs.LoadTokenConfig()
	if cfg == nil {
		return nil, ErrNilConfig
	}

	// 添加默认排除路径
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

// Middleware 返回用于验证JWT令牌的http.Handler
func (jm *jwtMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查当前路径是否在排除列表中
		if jm.isExcludedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

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

// initializeTokenExtractor 初始化令牌提取器
func (jm *jwtMiddleware) initializeTokenExtractor() error {
	extractors := map[string]func(*http.Request, string) (string, error){
		"header": extractTokenFromHeader, // 从请求头提取令牌
		"query":  extractTokenFromQuery,  // 从查询参数提取令牌
		"cookie": extractTokenFromCookie, // 从cookie提取令牌
	}

	parts := strings.SplitN(jm.conf.TokenLookup, ":", 2)
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

// extractTokenFromHeader 从请求头提取令牌
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

// extractTokenFromQuery 从查询参数提取令牌
func extractTokenFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", ErrMissingJwtToken
	}
	return token, nil
}

// extractTokenFromCookie 从cookie提取令牌
func extractTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", ErrMissingJwtToken
	}
	return cookie.Value, nil
}

// parseJwtToken 解析JWT令牌
func (jm *jwtMiddleware) parseJwtToken(tokenString string) (AuthClaims, error) {
	claims := tokenClaimsPool.Get().(*TokenClaims)
	defer tokenClaimsPool.Put(claims)

	*claims = TokenClaims{} // 重置claims
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

	returnClaims := &TokenClaims{
		Username:         claims.Username,
		TokenType:        claims.TokenType,
		RegisteredClaims: claims.RegisteredClaims,
		Data:             claims.Data,
	}

	return returnClaims, nil
}

// GenerateTokenPair 生成新的JWT令牌对
func (jm *jwtMiddleware) GenerateTokenPair(username string) (*TokenPair, error) {
	accessToken, err := jm.createToken(username, TokenTypeAccess, jm.conf.ExpiresTime)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jm.createToken(username, TokenTypeRefresh, jm.conf.RefreshExpiresTime)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// createToken 创建单个令牌
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

// RefreshToken 刷新访问令牌
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

	if tokenClaims.ExpiresAt != nil && tokenClaims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return jm.GenerateTokenPair(tokenClaims.Username)
}

// isExcludedPath 检查路径是否在排除列表中
func (jm *jwtMiddleware) isExcludedPath(reqPath string) bool {
	for _, excludePath := range jm.conf.ExcludePaths {
		// 处理通配符情况
		if strings.HasSuffix(excludePath, "*") {
			// 移除末尾的 "*" 并检查请求路径是否以此为前缀
			prefix := strings.TrimSuffix(excludePath, "*")
			if strings.HasPrefix(reqPath, prefix) {
				return true
			}
		} else if matched, _ := path.Match(excludePath, reqPath); matched {
			// 使用 path.Match 处理简单的模式匹配
			return true
		}
	}
	return false
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

// GetCurrentUser 获取当前用户名
func GetCurrentUser(ctx context.Context) (string, error) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", errors.New("未找到认证信息")
	}
	return claims.GetUsername(), nil
}

// defaultErrorHandler 默认错误处理函数
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// GetSigningMethod 根据方法名返回JWT签名方法
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
