// main.go

package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/nf/middleware"
	"github.com/sagoo-cloud/nexframe/utils/auth"
	"log"
	"net/http"
	"time"
)

var secretKey = []byte("你的密钥")

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	r := mux.NewRouter()

	// 应用 JWT 中间件到所有以 "/api" 开头的路由
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JwtMiddleware(auth.JwtConfig{
		SigningKey: secretKey,
	}))

	// 公开路由
	r.HandleFunc("/login", loginHandler).Methods("POST")

	// 受保护的路由
	api.HandleFunc("/protected", protectedHandler).Methods("GET")

	fmt.Println("服务器正在运行，地址为 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 这里应该验证用户凭证，但为了简单起见，我们只检查用户名是否为空
	if user.Username == "" {
		http.Error(w, "无效的用户名", http.StatusUnauthorized)
		return
	}

	// 使用新的辅助函数创建和发送 JWT 令牌
	claims := auth.TokenClaims{
		Username: user.Username,
	}
	err = auth.CreateAndSendToken(w, claims, secretKey, 24*time.Hour)
	if err != nil {
		log.Printf("创建令牌时出错: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(jwt.MapClaims)
	username := claims["username"].(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("欢迎，%s！这是一个受保护的路由。", username),
	})
}
