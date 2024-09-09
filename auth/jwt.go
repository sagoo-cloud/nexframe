// auth/jwt.go

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtConfig 定义了 JWT 中间件的配置
type JwtConfig struct {
	SigningKey    interface{}  // 用于签名的密钥
	SigningMethod string       // 签名方法
	TokenLookup   string       // 定义如何查找令牌
	ContextKey    string       // 用于在上下文中存储用户信息的键
	ErrorHandler  ErrorHandler // 错误处理函数
}

// ErrorHandler 是一个处理中间件错误的函数类型
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// JWTMiddleware 是验证 JWT 令牌的中间件
type JWTMiddleware struct {
	config JwtConfig
}

// TokenClaims 表示我们想要存储在令牌中的声明
type TokenClaims struct {
	Username string `json:"username"`
	// 在这里添加你想包含的其他声明
	jwt.RegisteredClaims
}

// TokenResponse 表示发送回客户端的响应
type TokenResponse struct {
	Token string `json:"token"`
}

// New 创建一个新的 JWTMiddleware 实例
func New(config JwtConfig) (*JWTMiddleware, error) {
	if config.SigningKey == nil {
		return nil, errors.New("jwt 中间件需要签名密钥")
	}

	if config.SigningMethod == "" {
		config.SigningMethod = "HS256"
	}

	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}

	if config.ContextKey == "" {
		config.ContextKey = "user"
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}

	return &JWTMiddleware{config: config}, nil
}

// Middleware 返回一个可以与 gorilla/mux 的 Use() 方法一起使用的 http.Handler
func (jm *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jm.extractToken(r)
		if err != nil {
			jm.config.ErrorHandler(w, r, err)
			return
		}

		claims, err := jm.parseToken(token)
		if err != nil {
			jm.config.ErrorHandler(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), jm.config.ContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken 从请求中提取令牌
func (jm *JWTMiddleware) extractToken(r *http.Request) (string, error) {
	parts := strings.Split(jm.config.TokenLookup, ":")
	switch parts[0] {
	case "header":
		return jm.extractTokenFromHeader(r, parts[1])
	case "query":
		return jm.extractTokenFromQuery(r, parts[1])
	case "cookie":
		return jm.extractTokenFromCookie(r, parts[1])
	}
	return "", errors.New("无效的令牌查找配置")
}

// extractTokenFromHeader 从请求头中提取令牌
func (jm *JWTMiddleware) extractTokenFromHeader(r *http.Request, header string) (string, error) {
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
func (jm *JWTMiddleware) extractTokenFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", errors.New("查询参数中缺少令牌")
	}
	return token, nil
}

// extractTokenFromCookie 从 cookie 中提取令牌
func (jm *JWTMiddleware) extractTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", errors.New("cookie 中缺少令牌")
	}
	return cookie.Value, nil
}

// parseToken 解析和验证令牌
func (jm *JWTMiddleware) parseToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jm.config.SigningMethod {
			return nil, fmt.Errorf("意外的 jwt 签名方法=%v", token.Header["alg"])
		}
		return jm.config.SigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// defaultErrorHandler 是默认的错误处理函数
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// CreateAndSendToken 创建一个 JWT 令牌并将其作为 JSON 响应发送
func CreateAndSendToken(w http.ResponseWriter, claims TokenClaims, signingKey []byte, expirationTime time.Duration) error {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expirationTime))

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码令牌作为字符串
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// 创建响应
	response := TokenResponse{
		Token: tokenString,
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}
