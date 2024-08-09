package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	http2 "github.com/sagoo-cloud/nexframe/weaver/http"
	"strings"
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

func Server(keyFunc jwt.Keyfunc, opts ...Option) http2.Middleware {
	o := &options{
		signingMethod: jwt.SigningMethodHS256,
	}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler http2.Handler) http2.Handler {
		return func(ctx http2.Context) error {
			if keyFunc == nil {
				return ErrMissingKeyFunc
			}
			for _, path := range o.filterPath {
				if path == ctx.Request().URL.Path {
					return handler(ctx)
				}
				//log.Logger(ctx).Info(ctx.Request().URL.Path)
			}
			auths := strings.SplitN(ctx.Request().Header.Get(authorizationKey), " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
				return ErrMissingJwtToken
			}
			jwtToken := auths[1]
			var (
				tokenInfo *jwt.Token
				err       error
			)
			if o.claims != nil {
				tokenInfo, err = jwt.ParseWithClaims(jwtToken, o.claims(), keyFunc)
			} else {
				tokenInfo, err = jwt.Parse(jwtToken, keyFunc)
			}
			if err != nil {
				if ve, ok := err.(*jwt.ValidationError); ok {
					if ve.Errors&jwt.ValidationErrorMalformed != 0 {
						return ErrTokenInvalid
					} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
						return ErrTokenExpired
					} else {
						return ErrTokenParseFail
					}
				}
				return err
			} else if !tokenInfo.Valid {
				return ErrTokenInvalid
			} else if tokenInfo.Method != o.signingMethod {
				return ErrUnSupportSigningMethod
			}
			return handler(ctx)
		}
	}
}

//func Server(keyFunc jwt.Keyfunc, opts ...Option) mux.MiddlewareFunc {
//	o := &options{
//		signingMethod: jwt.SigningMethodHS256,
//	}
//	for _, opt := range opts {
//		opt(o)
//	}
//	return func(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
//			if keyFunc == nil {
//				http.Error(writer, ErrMissingKeyFunc.Error(), http.StatusInternalServerError)
//				return
//			}
//			auths := strings.SplitN(request.Header.Get(authorizationKey), " ", 2)
//			fmt.Println(3333, auths)
//			if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
//				http.Error(writer, "", http.StatusInternalServerError)
//				return
//			}
//			jwtToken := auths[1]
//			var (
//				tokenInfo *jwt.Token
//				err       error
//			)
//			if o.claims != nil {
//				tokenInfo, err = jwt.ParseWithClaims(jwtToken, o.claims(), keyFunc)
//			} else {
//				tokenInfo, err = jwt.Parse(jwtToken, keyFunc)
//			}
//			fmt.Println(tokenInfo, err)
//			if err != nil {
//				if ve, ok := err.(*jwt.ValidationError); ok {
//					if ve.Errors&jwt.ValidationErrorMalformed != 0 {
//						http.Error(writer, ErrTokenInvalid.Error(), http.StatusInternalServerError)
//					} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
//						http.Error(writer, ErrTokenExpired.Error(), http.StatusInternalServerError)
//					} else {
//						http.Error(writer, ErrTokenParseFail.Error(), http.StatusInternalServerError)
//					}
//					return
//				}
//				http.Error(writer, err.Error(), http.StatusInternalServerError)
//			} else if !tokenInfo.Valid {
//				http.Error(writer, ErrTokenInvalid.Error(), http.StatusInternalServerError)
//				return
//			} else if tokenInfo.Method != o.signingMethod {
//				http.Error(writer, ErrUnSupportSigningMethod.Error(), http.StatusInternalServerError)
//				return
//			}
//			next.ServeHTTP(writer, request)
//
//		})
//	}
//}
