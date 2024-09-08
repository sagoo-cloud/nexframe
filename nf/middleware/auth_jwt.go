package middleware

import (
	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/utils/auth"
	"log"
)

// JwtMiddleware JWT 中间件
func JwtMiddleware(config auth.JwtConfig) mux.MiddlewareFunc {
	jwtMiddleware, err := auth.New(config)
	if err != nil {
		log.Fatal(err)
	}

	return jwtMiddleware.Middleware
}
