package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/gorilla/mux"
)

const (
	// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource exhaustion
	// attack vector
	extractorLimit = 20
)

var errHeaderExtractorValueMissing = errors.New("missing value in request header")
var errHeaderExtractorValueInvalid = errors.New("invalid value in request header")
var errQueryExtractorValueMissing = errors.New("missing value in the query string")
var errParamExtractorValueMissing = errors.New("missing value in path params")
var errCookieExtractorValueMissing = errors.New("missing value in cookies")
var errFormExtractorValueMissing = errors.New("missing value in the form")

// ValuesExtractor defines a function for extracting values (keys/tokens) from the given request.
type ValuesExtractor func(r *http.Request) ([]string, error)

// CreateExtractors creates ValuesExtractors from given lookups.
// Lookups is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract key from the request.
// Possible values:
//   - "header:<name>" or "header:<name>:<cut-prefix>"
//   - "query:<name>"
//   - "param:<name>"
//   - "form:<name>"
//   - "cookie:<name>"
//
// Multiple sources example:
// - "header:Authorization,header:X-Api-Key"
func CreateExtractors(lookups string) ([]ValuesExtractor, error) {
	return createExtractors(lookups, "")
}

func createExtractors(lookups string, authScheme string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}
	sources := strings.Split(lookups, ",")
	var extractors = make([]ValuesExtractor, 0)
	for _, source := range sources {
		parts := strings.Split(source, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v", source)
		}

		switch parts[0] {
		case "query":
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case "param":
			extractors = append(extractors, valuesFromParam(parts[1]))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1]))
		case "header":
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			} else if authScheme != "" && parts[1] == "Authorization" {
				prefix = authScheme
				if !strings.HasSuffix(prefix, " ") {
					prefix += " "
				}
			}
			extractors = append(extractors, valuesFromHeader(parts[1], prefix))
		}
	}
	return extractors, nil
}

func valuesFromHeader(header string, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	header = textproto.CanonicalMIMEHeaderKey(header)
	return func(r *http.Request) ([]string, error) {
		values := r.Header.Values(header)
		if len(values) == 0 {
			return nil, errHeaderExtractorValueMissing
		}

		result := make([]string, 0)
		for i, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				if i >= extractorLimit-1 {
					break
				}
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
				if i >= extractorLimit-1 {
					break
				}
			}
		}

		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, errHeaderExtractorValueInvalid
			}
			return nil, errHeaderExtractorValueMissing
		}
		return result, nil
	}
}

func valuesFromQuery(param string) ValuesExtractor {
	return func(r *http.Request) ([]string, error) {
		result := r.URL.Query()[param]
		if len(result) == 0 {
			return nil, errQueryExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}
		return result, nil
	}
}

func valuesFromParam(param string) ValuesExtractor {
	return func(r *http.Request) ([]string, error) {
		vars := mux.Vars(r)
		if value, ok := vars[param]; ok {
			return []string{value}, nil
		}
		return nil, errParamExtractorValueMissing
	}
}

func valuesFromCookie(name string) ValuesExtractor {
	return func(r *http.Request) ([]string, error) {
		cookie, err := r.Cookie(name)
		if err != nil {
			return nil, errCookieExtractorValueMissing
		}
		return []string{cookie.Value}, nil
	}
}

func valuesFromForm(name string) ValuesExtractor {
	return func(r *http.Request) ([]string, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		values := r.Form[name]
		if len(values) == 0 {
			return nil, errFormExtractorValueMissing
		}
		if len(values) > extractorLimit-1 {
			values = values[:extractorLimit]
		}
		result := append([]string{}, values...)
		return result, nil
	}
}

// ExtractorMiddleware is a middleware that extracts values using the provided extractors
func ExtractorMiddleware(extractors ...ValuesExtractor) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, extractor := range extractors {
				values, err := extractor(r)
				if err == nil && len(values) > 0 {
					// Store the extracted values in the request context
					ctx := context.WithValue(r.Context(), ExtractedValuesKey, values)
					r = r.WithContext(ctx)
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ExtractedValuesKey is the key used to store extracted values in the context
var ExtractedValuesKey = &contextKey{"extracted_values"}

type contextKey struct {
	name string
}

func (k *contextKey) String() string { return "middleware context value " + k.name }
