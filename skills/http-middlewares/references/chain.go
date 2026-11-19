package middleware

import "net/http"

// Middleware — функция, которая оборачивает следующий обработчик.
type Middleware func(http.Handler) http.Handler

// Chain применяет middleware так, что первый в списке — самый внешний.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
