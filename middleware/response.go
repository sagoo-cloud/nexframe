package middleware

import (
	"context"
	"encoding/json"
	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"net/http"
)

// DefaultHandlerResponse is the default implementation of HandlerResponse.
type DefaultHandlerResponse struct {
	Code    int         `json:"code"    dc:"Error code"`
	Message string      `json:"message" dc:"Error message"`
	Data    interface{} `json:"data"    dc:"Result data for certain request according API definition"`
}

// HandlerResponse is the middleware function for handling response and errors
func HandlerResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture the status code and content
		crw := &customResponseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(crw, r)

		// If there's custom buffer content, then exit current handler
		if crw.buf.Len() > 0 {
			w.Write(crw.buf.Bytes())
			return
		}

		var (
			msg  string
			err  = GetError(r)
			res  = GetHandlerResponse(r)
			code = gerror.Code(err)
		)

		if err != nil {
			if code == gcode.CodeNil {
				code = gcode.CodeInternalError
			}
			msg = err.Error()
		} else {
			if crw.status > 0 && crw.status != http.StatusOK {
				msg = http.StatusText(crw.status)
				switch crw.status {
				case http.StatusNotFound:
					code = gcode.CodeNotFound
				case http.StatusForbidden:
					code = gcode.CodeNotAuthorized
				default:
					code = gcode.CodeUnknown
				}
				// Create error and set it in the request context
				err = gerror.NewCode(code, msg)
				SetError(r, err)
			} else {
				code = gcode.CodeOK
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(crw.status)
		json.NewEncoder(w).Encode(DefaultHandlerResponse{
			Code:    code.Code(),
			Message: msg,
			Data:    res,
		})
	})
}

// Helper functions to get and set values in request context

func GetError(r *http.Request) error {
	if err, ok := r.Context().Value("error").(error); ok {
		return err
	}
	return nil
}

func SetError(r *http.Request, err error) {
	*r = *r.WithContext(context.WithValue(r.Context(), "error", err))
}

func GetHandlerResponse(r *http.Request) interface{} {
	return r.Context().Value("handlerResponse")
}

func SetHandlerResponse(r *http.Request, response interface{}) {
	*r = *r.WithContext(context.WithValue(r.Context(), "handlerResponse", response))
}
