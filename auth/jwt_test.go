package auth

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

// TestNewJwt 测试创建新的JWT中间件实例
// 测试确保NewJwt函数能够正确创建中间件实例，并检查关键配置是否正确设置
func TestNewJwt(t *testing.T) {
	middleware, err := NewJwt()
	if err != nil {
		t.Fatalf("创建中间件失败: %v", err)
	}

	if middleware == nil {
		t.Fatal("中间件不应为nil")
	}

	if middleware.conf.SigningMethod == nil {
		t.Error("签名方法不应为空")
	}
	if middleware.conf.TokenLookup == "" {
		t.Error("令牌查找配置不应为空")
	}
}

// TestMiddleware 测试JWT中间件的主要功能
// 测试涵盖了中间件处理各种情况的能力，包括有效令牌、无效令牌、缺少令牌和过期令牌
func TestMiddleware(t *testing.T) {
	middleware, err := NewJwt()
	if err != nil {
		t.Fatalf("创建中间件失败: %v", err)
	}

	// 创建一个受保护的处理程序，用于测试
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := GetCurrentUser(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user, ok := ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
		log.Print(user.GetUsername(), ":", user.GetUserID())
		w.Write([]byte("Welcome, " + username))
	})

	r := mux.NewRouter()
	r.Handle("/protected", middleware.Middleware(protectedHandler)).Methods("GET")

	// 测试有效令牌
	t.Run("ValidToken", func(t *testing.T) {
		tokenPair, err := middleware.GenerateTokenPair(UserInfo{
			ID:       11,
			Username: "testuser",
		})
		if err != nil {
			t.Fatalf("生成令牌对失败: %v", err)
		}

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v, 错误信息: %s", rr.Code, http.StatusOK, rr.Body.String())
		} else {
			expected := "Welcome, testuser"
			if rr.Body.String() != expected {
				t.Errorf("处理程序返回了意外的正文: 得到 %v 想要 %v", rr.Body.String(), expected)
			}
		}
	})
	// 测试无效令牌
	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusUnauthorized)
		}
	})
	// 测试缺少令牌
	t.Run("MissingToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusUnauthorized)
		}
	})
	// 测试过期令牌
	t.Run("ExpiredToken", func(t *testing.T) {
		expiredToken := createExpiredToken(t, middleware)
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusUnauthorized)
		}
	})

}

// TestTokenExtraction 测试从不同来源提取令牌的能力
// 测试检查中间件是否能够正确地从HTTP头、查询参数或cookie中提取令牌
func TestTokenExtraction(t *testing.T) {
	middleware, _ := NewJwt()
	tokenPair, _ := middleware.GenerateTokenPair(UserInfo{
		ID:       11,
		Username: "testuser",
	})

	testCases := []struct {
		name        string
		setToken    func(*http.Request)
		expectError bool
	}{
		{
			name: "从头部提取",
			setToken: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
			},
			expectError: false,
		},
		{
			name: "缺少令牌",
			setToken: func(r *http.Request) {
				// 不设置任何令牌
			},
			expectError: true,
		},
		{
			name: "无效的令牌格式",
			setToken: func(r *http.Request) {
				r.Header.Set("Authorization", "InvalidFormat "+tokenPair.AccessToken)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			tc.setToken(req)

			_, err := middleware.extractToken(req)
			if tc.expectError && err == nil {
				t.Error("预期错误，但没有得到")
			}
			if !tc.expectError && err != nil {
				t.Errorf("未预期的错误: %v", err)
			}
		})
	}
}

// TestGetSigningMethod 测试获取正确的签名方法
// 测试确保GetSigningMethod函数能够正确返回对应的JWT签名方法
func TestGetSigningMethod(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		expectedMethod jwt.SigningMethod
	}{
		{"HS256", "HS256", jwt.SigningMethodHS256},
		{"HS384", "HS384", jwt.SigningMethodHS384},
		{"HS512", "HS512", jwt.SigningMethodHS512},
		{"默认", "未知方法", jwt.SigningMethodHS256},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method := GetSigningMethod(tc.method)
			if method != tc.expectedMethod {
				t.Errorf("对于方法 %s，获得了错误的签名方法: 得到 %v, 期望 %v", tc.method, method, tc.expectedMethod)
			}
		})
	}
}

// TestClaimsFromContext 测试从上下文中提取声明
// 测试检查是否能够正确地从上下文中提取和验证JWT声明
func TestClaimsFromContext(t *testing.T) {
	ctx := context.Background()
	claims := &TokenClaims{Username: "testuser"}
	ctx = NewAuthContext(ctx, claims)

	retrievedClaims, ok := ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("无法从上下文中获取声明")
	}

	if retrievedClaims.GetUsername() != "testuser" {
		t.Errorf("获取到的用户名不正确: 得到 %s, 期望 testuser", retrievedClaims.GetUsername())
	}
}

// TestGetCurrentUser 测试获取当前用户
// 测试验证GetCurrentUser函数是否能够正确地从上下文中提取用户信息
func TestGetCurrentUser(t *testing.T) {
	ctx := context.Background()
	claims := &TokenClaims{Username: "testuser"}
	ctx = NewAuthContext(ctx, claims)

	username, err := GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("获取当前用户失败: %v", err)
	}

	if username != "testuser" {
		t.Errorf("获取到的用户名不正确: 得到 %s, 期望 testuser", username)
	}
}

// TestRefreshToken 测试刷新令牌的功能
// 测试检查刷新令牌的各种情况，包括有效的刷新令牌、过期的刷新令牌和无效的刷新令牌
func TestRefreshToken(t *testing.T) {
	middleware, _ := NewJwt()
	// 测试有效的刷新令牌
	t.Run("ValidRefreshToken", func(t *testing.T) {
		tokenPair, _ := middleware.GenerateTokenPair(UserInfo{
			Username: "testuser",
		})
		newTokenPair, err := middleware.RefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Fatalf("刷新令牌失败: %v", err)
		}
		if newTokenPair.AccessToken == "" || newTokenPair.RefreshToken == "" {
			t.Error("新的令牌对不应为空")
		}
	})
	// 测试过期的刷新令牌
	t.Run("ExpiredRefreshToken", func(t *testing.T) {
		expiredToken := createExpiredToken(t, middleware)
		_, err := middleware.RefreshToken(expiredToken)
		if err == nil {
			t.Error("使用过期的刷新令牌应该失败")
		}
	})
	// 测试无效的刷新令牌
	t.Run("InvalidRefreshToken", func(t *testing.T) {
		_, err := middleware.RefreshToken("invalid-token")
		if err == nil {
			t.Error("使用无效的刷新令牌应该失败")
		}
	})
}

// createExpiredToken 是一个辅助函数，用于创建过期的测试令牌
func createExpiredToken(t *testing.T, middleware *jwtMiddleware) string {
	now := time.Now().Add(-time.Hour) // 设置为1小时前
	claims := &TokenClaims{
		Username:  "testuser",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(middleware.conf.SigningMethod, claims)
	signedToken, err := token.SignedString(middleware.conf.SigningKey)
	if err != nil {
		t.Fatalf("创建过期令牌失败: %v", err)
	}
	return signedToken
}
