package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestExtractorMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(*http.Request) *http.Request
		extractors     []ValuesExtractor
		expectedValues []string
	}{
		{
			name: "Extract from header",
			setupRequest: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Bearer token123")
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromHeader("Authorization", "Bearer "),
			},
			expectedValues: []string{"token123"},
		},
		{
			name: "Extract from query",
			setupRequest: func(r *http.Request) *http.Request {
				q := r.URL.Query()
				q.Add("token", "querytoken123")
				r.URL.RawQuery = q.Encode()
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromQuery("token"),
			},
			expectedValues: []string{"querytoken123"},
		},
		{
			name: "Extract from param",
			setupRequest: func(r *http.Request) *http.Request {
				vars := map[string]string{"id": "123"}
				return mux.SetURLVars(r, vars)
			},
			extractors: []ValuesExtractor{
				valuesFromParam("id"),
			},
			expectedValues: []string{"123"},
		},
		{
			name: "Extract from cookie",
			setupRequest: func(r *http.Request) *http.Request {
				r.AddCookie(&http.Cookie{Name: "session", Value: "cookietoken123"})
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromCookie("session"),
			},
			expectedValues: []string{"cookietoken123"},
		},
		{
			name: "Extract from form",
			setupRequest: func(r *http.Request) *http.Request {
				r.PostForm = make(map[string][]string)
				r.PostForm.Add("token", "formtoken123")
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromForm("token"),
			},
			expectedValues: []string{"formtoken123"},
		},
		{
			name: "Multiple extractors",
			setupRequest: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Bearer headertoken123")
				q := r.URL.Query()
				q.Add("token", "querytoken123")
				r.URL.RawQuery = q.Encode()
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromHeader("Authorization", "Bearer "),
				valuesFromQuery("token"),
			},
			expectedValues: []string{"headertoken123"},
		},
		{
			name: "No matching extractor",
			setupRequest: func(r *http.Request) *http.Request {
				return r
			},
			extractors: []ValuesExtractor{
				valuesFromHeader("NonExistent", ""),
			},
			expectedValues: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			r = tt.setupRequest(r)

			rr := httptest.NewRecorder()

			middleware := ExtractorMiddleware(tt.extractors...)

			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				values, ok := r.Context().Value(ExtractedValuesKey).([]string)
				if !ok {
					if tt.expectedValues != nil {
						t.Error("Failed to retrieve extracted values from context")
					}
					return
				}
				if len(values) != len(tt.expectedValues) {
					t.Errorf("Expected %d values, got %d", len(tt.expectedValues), len(values))
					return
				}
				for i, v := range values {
					if v != tt.expectedValues[i] {
						t.Errorf("Expected value %s, got %s", tt.expectedValues[i], v)
					}
				}
			})

			middleware(handler).ServeHTTP(rr, r)

			if !handlerCalled {
				t.Error("Handler was not called")
			}
		})
	}
}

func TestCreateExtractors(t *testing.T) {
	tests := []struct {
		name          string
		lookups       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Single extractor",
			lookups:       "header:Authorization",
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Multiple extractors",
			lookups:       "header:Authorization,query:token,param:id",
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "Invalid lookup",
			lookups:       "invalid:field",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Empty lookup",
			lookups:       "",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Invalid format",
			lookups:       "header",
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractors, err := CreateExtractors(tt.lookups)

			if tt.expectError && err == nil {
				t.Error("Expected an error, but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(extractors) != tt.expectedCount {
				t.Errorf("Expected %d extractors, got %d", tt.expectedCount, len(extractors))
			}
		})
	}
}

func TestExtractorMiddlewareWithNoExtractors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware := ExtractorMiddleware()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		values, ok := r.Context().Value(ExtractedValuesKey).([]string)
		if ok {
			t.Error("Expected no values in context, but found some")
		}
		if values != nil {
			t.Error("Expected nil values, but got non-nil")
		}
	})

	middleware(handler).ServeHTTP(rr, r)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}
