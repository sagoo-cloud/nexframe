package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	http2 "github.com/sagoo-cloud/nexframe/weaver/http"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type CustomerClaims struct {
	ID   int64  `json:"Id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

//func TestJWTServerParse(t *testing.T) {
//	//var (
//	//	errConcurrentWrite = errors.New("concurrent write claims")
//	//	errParseClaims     = errors.New("bad result, token claims is not CustomerClaims")
//	//)
//
//	testKey := "testKey"
//	tests := []struct {
//		name         string
//		token        func() string
//		claims       func() jwt.Claims
//		exceptErr    error
//		key          string
//		goroutineNum int
//	}{
//		{
//			name: "normal",
//			token: func() string {
//				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomerClaims{}).SignedString([]byte(testKey))
//				if err != nil {
//					panic(err)
//				}
//				return fmt.Sprintf(bearerFormat, token)
//			},
//			claims: func() jwt.Claims {
//				return &CustomerClaims{}
//			},
//			exceptErr:    nil,
//			key:          testKey,
//			goroutineNum: 1,
//		},
//		{
//			name: "concurrent request",
//			token: func() string {
//				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomerClaims{
//					Username: strconv.Itoa(rand.Int()),
//				}).SignedString([]byte(testKey))
//				if err != nil {
//					panic(err)
//				}
//				return fmt.Sprintf(bearerFormat, token)
//			},
//			claims: func() jwt.Claims {
//				return &CustomerClaims{}
//			},
//			exceptErr:    nil,
//			key:          testKey,
//			goroutineNum: 10000,
//		},
//	}
//
//	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
//		//// mock biz
//		//time.Sleep(100 * time.Millisecond)
//		//
//		//if customerClaims, ok := testToken.(*CustomerClaims); ok {
//		//	if name != customerClaims.Username {
//		//		return nil, errConcurrentWrite
//		//	}
//		//} else {
//		//	return nil, errParseClaims
//		//}
//		//return "reply", nil
//	})
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			server := Server(
//				func(token *jwt.Token) (interface{}, error) { return []byte(testKey), nil },
//				WithClaims(test.claims),
//			)
//			req, _ := http.NewRequest("GET", "/", nil)
//			req.Header.Add(authorizationKey, test.token())
//			writer := httptest.NewRecorder()
//			server.Middleware(next).ServeHTTP(writer, req)
//		})
//	}
//}

func TestJWTServerParse(t *testing.T) {
	errParseClaims := errors.New("bad result, token claims is not UserClaims")
	testKey := "testKey"
	tests := []struct {
		name      string
		token     func() string
		claims    func() jwt.Claims
		exceptErr error
		key       string
	}{
		{
			name: "正常的",
			token: func() string {
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{}).SignedString([]byte(testKey))
				if err != nil {
					panic(err)
				}
				return fmt.Sprintf(bearerFormat, token)
			},
			claims: func() jwt.Claims {
				return &UserClaims{}
			},
			exceptErr: nil,
			key:       testKey,
		},
		{
			name: "标准的",
			token: func() string {
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
					Username: strconv.Itoa(rand.Int()),
				}).SignedString([]byte(testKey))
				if err != nil {
					panic(err)
				}
				return fmt.Sprintf(bearerFormat, token)
			},
			claims: func() jwt.Claims {
				return &UserClaims{}
			},
			exceptErr: nil,
			key:       testKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := Server(
				func(token *jwt.Token) (interface{}, error) { return []byte(testKey), nil },
				WithClaims(test.claims),
			)
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Add(authorizationKey, test.token())
			ctx := &http2.Wrapper{}
			ctx.Reset(httptest.NewRecorder(), req)
			err := server(func(ctx http2.Context) error {
				testToken, ok := FromContext(ctx)
				if !ok {
					return errors.New("先登录")
				}
				if user, ok := testToken.(*UserClaims); ok {
					return user.Valid()
				}
				return errParseClaims
			})(ctx)
			if test.exceptErr != err {
				t.Error(err)
			}
		})
	}
}
