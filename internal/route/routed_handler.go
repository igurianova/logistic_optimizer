package route

import (
	"net/http"
)

// RoutedHandler Базовый интерфейс обработчика используемый для построения таблицы маршрутизации
type RoutedHandler interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
	GetUrls() []string
}
