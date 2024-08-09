package http

type Wrapper = wrapper

// Handler defines the handler invoked by Middleware.
type Handler = HandlerFunc

// Middleware is HTTP middleware.
type Middleware func(Handler) Handler

func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}
