package nf

import "net/http"

func (f *APIFramework) BindHandler(prefix string, handler http.Handler) error {
	f.router.Handle(prefix, handler)
	return nil
}
func (f *APIFramework) BindHandlerFunc(prefix string, handler http.HandlerFunc) error {
	f.router.Handle(prefix, handler)
	return nil
}

// BindStatusHandler binds the status handler for the specified pattern.
func (f *APIFramework) BindStatusHandler(status int, handler http.HandlerFunc) {
	f.router.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		handler(w, r)
	})
}

// BindStatusHandlerByMap binds multiple status handlers using a map.
func (f *APIFramework) BindStatusHandlerByMap(handlers map[int]http.HandlerFunc) {
	for status, handler := range handlers {
		f.BindStatusHandler(status, handler)
	}
}
