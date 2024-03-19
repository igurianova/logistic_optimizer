package route

import (
	"github.com/igurianova/logistic_optimizer/internal/middleware"
	"net/http"
)

// HandlerRouteBuilder Построитель таблицы маршрутизации
type HandlerRouteBuilder struct {
	handlers    *[]RoutedHandler
	middlewares *[]middleware.Middleware
}

func Builder() *HandlerRouteBuilder {
	return &HandlerRouteBuilder{&[]RoutedHandler{}, &[]middleware.Middleware{}}
}

func (b *HandlerRouteBuilder) AddHandler(handler RoutedHandler) *HandlerRouteBuilder {
	*b.handlers = append(*b.handlers, handler)
	return b
}

func (b *HandlerRouteBuilder) AddMiddleware(middleware middleware.Middleware) *HandlerRouteBuilder {
	*b.middlewares = append(*b.middlewares, middleware)
	return b
}

func (b *HandlerRouteBuilder) Build() map[string]http.Handler {
	routes := make(map[string]http.Handler)
	for _, handler := range *b.handlers {
		for _, baseUrl := range handler.GetUrls() {
			httpHandler := handler.(http.Handler)
			for _, m := range *b.middlewares {
				httpHandler = m(httpHandler)
			}
			routes[baseUrl] = httpHandler
		}
	}
	return routes
}
