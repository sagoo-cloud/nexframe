package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var testSecretKey = []byte("test-secret-key")

// TestNewMiddleware 测试创建新的中间件实例
func TestNewMiddleware(t *testing.T) {
	// 测试成功创建中间件
	_, err := New(JwtConfig{SigningKey: testSecretKey})
	if err != nil {
		t.Fatalf("创建中间件失败: %v", err)
	}

	// 测试没有签名密钥时的失败情况
	_, err = New(JwtConfig{})
	if err == nil {
		t.Fatal("预期没有签名密钥时会失败，但没有")
	}

	// 测试默认值
	middleware, _ := New(JwtConfig{SigningKey: testSecretKey})
	if middleware.conf.SigningMethod != "HS256" {
		t.Errorf("默认签名方法错误: 得到 %v, 期望 HS256", middleware.conf.SigningMethod)
	}
	if middleware.conf.TokenLookup != "header:Authorization" {
		t.Errorf("默认令牌查找错误: 得到 %v, 期望 header:Authorization", middleware.conf.TokenLookup)
	}
}

// TestMiddleware 测试中间件的主要功能
func TestMiddleware(t *testing.T) {
	middleware, _ := New(JwtConfig{
		SigningKey:    testSecretKey,
		SigningMethod: "HS256",
	})

	// 创建一个受保护的处理程序
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext(r.Context())
		if !ok {
			w.WriteHeader(401)
			return
		}
		w.Write([]byte("Welcome, " + claims.(*TokenClaims).Username))
	})

	// 创建路由器并应用中间件
	r := mux.NewRouter()
	r.Handle("/protected", middleware.Middleware(protectedHandler)).Methods("GET")

	// 测试有效令牌
	t.Run("ValidToken", func(t *testing.T) {
		token := createToken(t, "testuser", time.Hour)
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusOK)
		}
		expected := "Welcome, testuser"
		if rr.Body.String() != expected {
			t.Errorf("处理程序返回了意外的正文: 得到 %v 想要 %v", rr.Body.String(), expected)
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
		token := createToken(t, "testuser", -time.Hour) // 创建一个已经过期的令牌
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusUnauthorized)
		}
	})
}

// TestTokenExtraction 测试从不同来源提取令牌
func TestTokenExtraction(t *testing.T) {
	token := createToken(t, "testuser", time.Hour)

	testCases := []struct {
		name        string
		tokenLookup string
		setToken    func(*http.Request)
		expectError bool
	}{
		{
			name:        "从头部提取",
			tokenLookup: "header:Authorization",
			setToken: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+token)
			},
			expectError: false,
		},
		{
			name:        "从查询参数提取",
			tokenLookup: "query:token",
			setToken: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
			},
			expectError: false,
		},
		{
			name:        "从cookie提取",
			tokenLookup: "cookie:token",
			setToken: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "token", Value: token})
			},
			expectError: false,
		},
		{
			name:        "无效的提取方法",
			tokenLookup: "invalid:token",
			setToken:    func(r *http.Request) {},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware, _ := New(JwtConfig{SigningKey: testSecretKey, TokenLookup: tc.tokenLookup})

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

// TestCreateAndSendToken 测试创建和发送令牌
func TestCreateAndSendToken(t *testing.T) {
	w := httptest.NewRecorder()
	token, err := GenerateToken(string(testSecretKey), WithClaims(func() jwt.Claims {
		return &TokenClaims{
			Username: "testuser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
	}))
	if err != nil {
		t.Fatalf("创建令牌失败: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("意外的状态码: 得到 %v 想要 %v", w.Code, http.StatusOK)
	}

	// 验证令牌
	claims, err := ParseJwtToken(token, WithKeyFunc(func(token *jwt.Token) (interface{}, error) {
		return testSecretKey, nil
	}))
	if err != nil {
		t.Fatalf("解析令牌失败: %v", err)
	}
	if claims, ok := claims.(*TokenClaims); ok {
		if claims.Username != "testuser" {
			t.Errorf("意外的用户名: 得到 %v 想要 %v", claims.Username, "testuser")
		}
	} else {
		t.Fatal("无法获取令牌声明")
	}
}

// TestCustomErrorHandler 测试自定义错误处理
func TestCustomErrorHandler(t *testing.T) {
	customErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "自定义错误: "+err.Error(), http.StatusForbidden)
	}

	middleware, _ := New(JwtConfig{
		SigningKey: testSecretKey,
		ErrHandler: customErrorHandler,
	})

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("处理程序返回了错误的状态码: 得到 %v 想要 %v", rr.Code, http.StatusForbidden)
	}

	if !strings.Contains(rr.Body.String(), "自定义错误") {
		t.Errorf("处理程序没有使用自定义错误消息: %v", rr.Body.String())
	}
}

// createToken 是一个辅助函数，用于创建有效的测试令牌
func createToken(t *testing.T, username string, expiration time.Duration) string {
	token, err := GenerateToken(string(testSecretKey), WithClaims(func() jwt.Claims {
		return &TokenClaims{
			Username: "testuser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			},
		}
	}))
	if err != nil {
		t.Fatalf("创建令牌失败: %v", err)
	}

	return token
}
