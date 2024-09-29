package middleware

import (
	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/auth"
	"log"
)

// JwtMiddleware JWT 中间件
func JwtMiddleware(config auth.JwtConfig) mux.MiddlewareFunc {
	jwtMiddleware, err := auth.NewJwt()
	if err != nil {
		log.Fatal(err)
	}

	return jwtMiddleware.Middleware
}
