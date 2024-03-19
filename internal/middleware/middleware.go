package middleware

import "net/http"

// Middleware Функция выполняемая перед пападанием в обработчик
type Middleware func(http.Handler) http.Handler
